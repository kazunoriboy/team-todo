package auth

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Context keys for storing auth info
const (
	UserIDKey      = "user_id"
	UserEmailKey   = "user_email"
	UserNameKey    = "user_name"
	UserClaimsKey  = "user_claims"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(jwtService *JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Check for Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := parts[1]

			// Validate the token
			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				if err == ErrExpiredToken {
					return echo.NewHTTPError(http.StatusUnauthorized, "token has expired")
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Store user info in context
			c.Set(UserIDKey, claims.UserID)
			c.Set(UserEmailKey, claims.Email)
			c.Set(UserNameKey, claims.DisplayName)
			c.Set(UserClaimsKey, claims)

			return next(c)
		}
	}
}

// OptionalAuthMiddleware allows requests without authentication but sets user info if token is present
func OptionalAuthMiddleware(jwtService *JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return next(c)
			}

			tokenString := parts[1]
			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err == nil {
				c.Set(UserIDKey, claims.UserID)
				c.Set(UserEmailKey, claims.Email)
				c.Set(UserNameKey, claims.DisplayName)
				c.Set(UserClaimsKey, claims)
			}

			return next(c)
		}
	}
}

// GetUserID retrieves the user ID from context
func GetUserID(c echo.Context) (uuid.UUID, bool) {
	userID, ok := c.Get(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetUserEmail retrieves the user email from context
func GetUserEmail(c echo.Context) (string, bool) {
	email, ok := c.Get(UserEmailKey).(string)
	return email, ok
}

// GetClaims retrieves the full claims from context
func GetClaims(c echo.Context) (*Claims, bool) {
	claims, ok := c.Get(UserClaimsKey).(*Claims)
	return claims, ok
}

