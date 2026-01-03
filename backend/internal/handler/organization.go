package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"backend/ent"
	"backend/ent/invite"
	"backend/ent/organization"
	"backend/ent/organizationmember"
	"backend/internal/auth"
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// validate is the validator instance (shared from auth.go)
var orgValidate = validator.New()

// OrganizationHandler handles organization-related requests
type OrganizationHandler struct {
	client       *ent.Client
	emailService *service.EmailService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(client *ent.Client, emailService *service.EmailService) *OrganizationHandler {
	return &OrganizationHandler{
		client:       client,
		emailService: emailService,
	}
}

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required"`
	Slug string `json:"slug" validate:"required"`
}

// OrganizationResponse represents the organization data in responses
type OrganizationResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Role      string    `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// InviteRequest represents the request to invite a user
type InviteRequest struct {
	Email     string  `json:"email" validate:"required,email"`
	Role      string  `json:"role" validate:"required,oneof=admin member"`
	ProjectID *string `json:"project_id,omitempty"`
}

// InviteResponse represents the invite data in responses
type InviteResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request using validator
	if err := orgValidate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
	}

	ctx := c.Request().Context()

	// Check if slug is already taken
	exists, err := h.client.Organization.Query().
		Where(organization.SlugEQ(req.Slug)).
		Exist(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check slug availability")
	}
	if exists {
		return echo.NewHTTPError(http.StatusConflict, "slug is already taken")
	}

	// Create organization in a transaction
	tx, err := h.client.Tx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to start transaction")
	}

	// Create the organization
	org, err := tx.Organization.Create().
		SetName(req.Name).
		SetSlug(req.Slug).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create organization")
	}

	// Add creator as owner
	_, err = tx.OrganizationMember.Create().
		SetUserID(userID).
		SetOrganizationID(org.ID).
		SetRole(organizationmember.RoleOwner).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add owner to organization")
	}

	// Create default project
	_, err = tx.Project.Create().
		SetName("全般").
		SetOrganizationID(org.ID).
		SetIsPrivate(false).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create default project")
	}

	// Update user's last accessed org
	_, err = tx.User.UpdateOneID(userID).
		SetLastOrgID(org.ID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user context")
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction")
	}

	return c.JSON(http.StatusCreated, OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Role:      "owner",
		CreatedAt: org.CreatedAt,
	})
}

// ListOrganizations lists all organizations the user belongs to
func (h *OrganizationHandler) ListOrganizations(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	ctx := c.Request().Context()

	memberships, err := h.client.OrganizationMember.Query().
		Where(organizationmember.UserIDEQ(userID)).
		WithOrganization().
		All(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list organizations")
	}

	orgs := make([]OrganizationResponse, len(memberships))
	for i, m := range memberships {
		orgs[i] = OrganizationResponse{
			ID:        m.Edges.Organization.ID,
			Name:      m.Edges.Organization.Name,
			Slug:      m.Edges.Organization.Slug,
			Role:      string(m.Role),
			CreatedAt: m.Edges.Organization.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, orgs)
}

// GetOrganization gets an organization by slug
func (h *OrganizationHandler) GetOrganization(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	slug := c.Param("slug")
	if slug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "slug is required")
	}

	ctx := c.Request().Context()

	// Get organization
	org, err := h.client.Organization.Query().
		Where(organization.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "organization not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get organization")
	}

	// Check membership
	membership, err := h.client.OrganizationMember.Query().
		Where(
			organizationmember.UserIDEQ(userID),
			organizationmember.OrganizationIDEQ(org.ID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this organization")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check membership")
	}

	// Update user's last accessed org
	_, _ = h.client.User.UpdateOneID(userID).
		SetLastOrgID(org.ID).
		Save(ctx)

	return c.JSON(http.StatusOK, OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Role:      string(membership.Role),
		CreatedAt: org.CreatedAt,
	})
}

// InviteMember invites a user to an organization
func (h *OrganizationHandler) InviteMember(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	slug := c.Param("slug")
	if slug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "slug is required")
	}

	var req InviteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate request using validator
	if err := orgValidate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
	}

	ctx := c.Request().Context()

	// Get organization
	org, err := h.client.Organization.Query().
		Where(organization.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "organization not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get organization")
	}

	// Check if user has permission to invite (must be owner or admin)
	membership, err := h.client.OrganizationMember.Query().
		Where(
			organizationmember.UserIDEQ(userID),
			organizationmember.OrganizationIDEQ(org.ID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this organization")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check membership")
	}

	if !HasAdminPermission(membership.Role) {
		return echo.NewHTTPError(http.StatusForbidden, "only owners and admins can invite members")
	}

	// Generate invite token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate invite token")
	}
	token := hex.EncodeToString(tokenBytes)

	// Determine role
	role := invite.RoleMember
	if req.Role == "admin" {
		role = invite.RoleAdmin
	}

	// Create invite
	inv, err := h.client.Invite.Create().
		SetToken(token).
		SetEmail(req.Email).
		SetOrganizationID(org.ID).
		SetRole(role).
		SetInvitedByID(userID).
		SetExpiresAt(time.Now().Add(7 * 24 * time.Hour)).
		Save(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create invite")
	}

	// Get inviter name
	inviter, _ := h.client.User.Get(ctx, userID)
	inviterName := "Someone"
	if inviter != nil {
		inviterName = inviter.DisplayName
	}

	// Send invite email
	go func() {
		_ = h.emailService.SendInviteEmail(ctx, req.Email, inviterName, org.Name, token)
	}()

	return c.JSON(http.StatusCreated, InviteResponse{
		ID:        inv.ID,
		Email:     inv.Email,
		Role:      string(inv.Role),
		ExpiresAt: inv.ExpiresAt,
		CreatedAt: inv.CreatedAt,
	})
}

// AcceptInvite accepts an invite and joins the organization
func (h *OrganizationHandler) AcceptInvite(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	token := c.Param("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token is required")
	}

	ctx := c.Request().Context()

	// Get invite
	inv, err := h.client.Invite.Query().
		Where(
			invite.TokenEQ(token),
			invite.UsedAtIsNil(),
			invite.ExpiresAtGT(time.Now()),
		).
		WithOrganization().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "invite not found or expired")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get invite")
	}

	// Check if user is already a member
	exists, err := h.client.OrganizationMember.Query().
		Where(
			organizationmember.UserIDEQ(userID),
			organizationmember.OrganizationIDEQ(inv.OrganizationID),
		).
		Exist(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check membership")
	}
	if exists {
		return echo.NewHTTPError(http.StatusConflict, "you are already a member of this organization")
	}

	// Transaction: add member and mark invite as used
	tx, err := h.client.Tx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to start transaction")
	}

	// Add user as member
	role := organizationmember.RoleMember
	if inv.Role == invite.RoleAdmin {
		role = organizationmember.RoleAdmin
	}

	_, err = tx.OrganizationMember.Create().
		SetUserID(userID).
		SetOrganizationID(inv.OrganizationID).
		SetRole(role).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add member")
	}

	// Mark invite as used
	_, err = tx.Invite.UpdateOne(inv).
		SetUsedAt(time.Now()).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update invite")
	}

	// Update user's last accessed org
	_, err = tx.User.UpdateOneID(userID).
		SetLastOrgID(inv.OrganizationID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user context")
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction")
	}

	return c.JSON(http.StatusOK, OrganizationResponse{
		ID:        inv.Edges.Organization.ID,
		Name:      inv.Edges.Organization.Name,
		Slug:      inv.Edges.Organization.Slug,
		Role:      string(role),
		CreatedAt: inv.Edges.Organization.CreatedAt,
	})
}

// GetInviteInfo gets public info about an invite (for showing before login)
func (h *OrganizationHandler) GetInviteInfo(c echo.Context) error {
	token := c.Param("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token is required")
	}

	ctx := c.Request().Context()

	inv, err := h.client.Invite.Query().
		Where(
			invite.TokenEQ(token),
			invite.UsedAtIsNil(),
			invite.ExpiresAtGT(time.Now()),
		).
		WithOrganization().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "invite not found or expired")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get invite")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"organization_name": inv.Edges.Organization.Name,
		"organization_slug": inv.Edges.Organization.Slug,
		"email":             inv.Email,
		"expires_at":        inv.ExpiresAt,
	})
}

