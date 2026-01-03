package handler

import (
	"context"
	"net/http"
	"time"

	"backend/ent"
	"backend/ent/user"
	"backend/internal/auth"
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// validate is the validator instance
var validate = validator.New()

// formatValidationError formats validation errors into a user-friendly message
func formatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				return field + " is required"
			case "email":
				return field + " must be a valid email address"
			case "min":
				return field + " must be at least " + e.Param() + " characters"
			case "max":
				return field + " must be at most " + e.Param() + " characters"
			default:
				return field + " is invalid"
			}
		}
	}
	return "validation failed"
}

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	client       *ent.Client
	jwtService   *auth.JWTService
	emailService *service.EmailService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(client *ent.Client, jwtService *auth.JWTService, emailService *service.EmailService) *AuthHandler {
	return &AuthHandler{
		client:       client,
		jwtService:   jwtService,
		emailService: emailService,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest represents the token refresh request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         UserResponse     `json:"user"`
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	ExpiresIn    int64            `json:"expires_in"`
}

// UserResponse represents the user data in responses
type UserResponse struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	DisplayName   string     `json:"display_name"`
	LastOrgID     *uuid.UUID `json:"last_org_id,omitempty"`
	LastProjectID *uuid.UUID `json:"last_project_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request using validator
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
	}

	ctx := c.Request().Context()

	// Check if user already exists
	exists, err := h.client.User.Query().
		Where(user.EmailEQ(req.Email)).
		Exist(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check user existence")
	}
	if exists {
		return echo.NewHTTPError(http.StatusConflict, "user with this email already exists")
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	// Create user
	u, err := h.client.User.Create().
		SetEmail(req.Email).
		SetPasswordHash(passwordHash).
		SetDisplayName(req.DisplayName).
		Save(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(u.ID, u.Email, u.DisplayName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate tokens")
	}

	// Send welcome email (non-blocking)
	go func() {
		_ = h.emailService.SendWelcomeEmail(context.Background(), u.Email, u.DisplayName)
	}()

	return c.JSON(http.StatusCreated, AuthResponse{
		User: UserResponse{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			CreatedAt:   u.CreatedAt,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request using validator
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
	}

	ctx := c.Request().Context()

	// Find user by email
	u, err := h.client.User.Query().
		Where(user.EmailEQ(req.Email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find user")
	}

	// Verify password
	if !auth.CheckPassword(req.Password, u.PasswordHash) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	// Generate tokens
	tokens, err := h.jwtService.GenerateTokenPair(u.ID, u.Email, u.DisplayName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate tokens")
	}

	return c.JSON(http.StatusOK, AuthResponse{
		User: UserResponse{
			ID:            u.ID,
			Email:         u.Email,
			DisplayName:   u.DisplayName,
			LastOrgID:     u.LastOrgID,
			LastProjectID: u.LastProjectID,
			CreatedAt:     u.CreatedAt,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request using validator
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
	}

	// Validate refresh token
	userID, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	ctx := c.Request().Context()

	// Get user
	u, err := h.client.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find user")
	}

	// Generate new tokens
	tokens, err := h.jwtService.GenerateTokenPair(u.ID, u.Email, u.DisplayName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate tokens")
	}

	return c.JSON(http.StatusOK, AuthResponse{
		User: UserResponse{
			ID:            u.ID,
			Email:         u.Email,
			DisplayName:   u.DisplayName,
			LastOrgID:     u.LastOrgID,
			LastProjectID: u.LastProjectID,
			CreatedAt:     u.CreatedAt,
		},
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

// GetMe returns the current authenticated user
func (h *AuthHandler) GetMe(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	ctx := c.Request().Context()

	u, err := h.client.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find user")
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		LastOrgID:     u.LastOrgID,
		LastProjectID: u.LastProjectID,
		CreatedAt:     u.CreatedAt,
	})
}

// UpdateMe updates the current authenticated user
func (h *AuthHandler) UpdateMe(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req struct {
		DisplayName *string `json:"display_name,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	update := h.client.User.UpdateOneID(userID)
	if req.DisplayName != nil && *req.DisplayName != "" {
		update.SetDisplayName(*req.DisplayName)
	}

	u, err := update.Save(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user")
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		LastOrgID:     u.LastOrgID,
		LastProjectID: u.LastProjectID,
		CreatedAt:     u.CreatedAt,
	})
}

