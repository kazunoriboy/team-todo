package handler

import (
	"net/http"

	"backend/ent"
	"backend/ent/organizationmember"
	"backend/ent/project"
	"backend/ent/projectmember"
	"backend/internal/auth"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ContextHandler handles context-related requests (for restoring user's last state)
type ContextHandler struct {
	client *ent.Client
}

// NewContextHandler creates a new context handler
func NewContextHandler(client *ent.Client) *ContextHandler {
	return &ContextHandler{client: client}
}

// ContextResponse represents the user's current context
type ContextResponse struct {
	HasContext   bool                  `json:"has_context"`
	Organization *OrganizationResponse `json:"organization,omitempty"`
	Project      *ProjectResponse      `json:"project,omitempty"`
	RedirectURL  string                `json:"redirect_url,omitempty"`
}

// GetCurrentContext returns the user's last accessed organization and project
func (h *ContextHandler) GetCurrentContext(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	ctx := c.Request().Context()

	// Get user with last context
	user, err := h.client.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user")
	}

	response := ContextResponse{
		HasContext: false,
	}

	// Check if user has last org
	if user.LastOrgID == nil {
		// No context, check if user has any orgs
		memberships, err := h.client.OrganizationMember.Query().
			Where(organizationmember.UserIDEQ(userID)).
			WithOrganization().
			All(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get memberships")
		}

		if len(memberships) == 0 {
			// User has no orgs, redirect to create org
			response.RedirectURL = "/org/new"
		} else {
			// Redirect to first org
			org := memberships[0].Edges.Organization
			response.RedirectURL = "/org/" + org.Slug
		}
		return c.JSON(http.StatusOK, response)
	}

	// Verify user still has access to the last org
	org, err := h.client.Organization.Get(ctx, *user.LastOrgID)
	if err != nil {
		if ent.IsNotFound(err) {
			// Org no longer exists, clear context and redirect
			_, _ = h.client.User.UpdateOneID(userID).
				ClearLastOrgID().
				ClearLastProjectID().
				Save(ctx)
			response.RedirectURL = "/org/new"
			return c.JSON(http.StatusOK, response)
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
			// No longer a member, clear context
			_, _ = h.client.User.UpdateOneID(userID).
				ClearLastOrgID().
				ClearLastProjectID().
				Save(ctx)
			response.RedirectURL = "/org/new"
			return c.JSON(http.StatusOK, response)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check membership")
	}

	response.HasContext = true
	response.Organization = &OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Role:      string(membership.Role),
		CreatedAt: org.CreatedAt,
	}
	response.RedirectURL = "/org/" + org.Slug

	// Check last project if exists
	if user.LastProjectID != nil {
		proj, err := h.client.Project.Query().
			Where(
				project.IDEQ(*user.LastProjectID),
				project.OrganizationIDEQ(org.ID),
			).
			Only(ctx)
		if err == nil {
			// Check project access
			hasAccess := !proj.IsPrivate // Public projects are accessible to all org members
			permission := "view"

			if proj.IsPrivate {
				pm, err := h.client.ProjectMember.Query().
					Where(
						projectmember.UserIDEQ(userID),
						projectmember.ProjectIDEQ(proj.ID),
					).
					Only(ctx)
				if err == nil {
					hasAccess = true
					permission = string(pm.Permission)
				}
			} else {
				// Check for explicit membership for higher permission
				pm, err := h.client.ProjectMember.Query().
					Where(
						projectmember.UserIDEQ(userID),
						projectmember.ProjectIDEQ(proj.ID),
					).
					Only(ctx)
				if err == nil {
					permission = string(pm.Permission)
				}
			}

			if hasAccess {
				response.Project = &ProjectResponse{
					ID:             proj.ID,
					Name:           proj.Name,
					IsPrivate:      proj.IsPrivate,
					OrganizationID: proj.OrganizationID,
					Permission:     permission,
					CreatedAt:      proj.CreatedAt,
				}
				response.RedirectURL = "/org/" + org.Slug + "/projects/" + proj.ID.String()
			}
		}
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateContextRequest represents the request to update context
type UpdateContextRequest struct {
	OrgID     *string `json:"org_id,omitempty"`
	ProjectID *string `json:"project_id,omitempty"`
}

// UpdateContext updates the user's last accessed organization and project
func (h *ContextHandler) UpdateContext(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req UpdateContextRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()
	update := h.client.User.UpdateOneID(userID)

	if req.OrgID != nil {
		orgID, err := uuid.Parse(*req.OrgID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid org_id format")
		}

		// Verify membership
		exists, err := h.client.OrganizationMember.Query().
			Where(
				organizationmember.UserIDEQ(userID),
				organizationmember.OrganizationIDEQ(orgID),
			).
			Exist(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify membership")
		}
		if !exists {
			return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this organization")
		}

		update.SetLastOrgID(orgID)
	}

	if req.ProjectID != nil {
		projectID, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id format")
		}

		// Get project and verify access
		proj, err := h.client.Project.Get(ctx, projectID)
		if err != nil {
			if ent.IsNotFound(err) {
				return echo.NewHTTPError(http.StatusNotFound, "project not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get project")
		}

		// Check org membership
		exists, err := h.client.OrganizationMember.Query().
			Where(
				organizationmember.UserIDEQ(userID),
				organizationmember.OrganizationIDEQ(proj.OrganizationID),
			).
			Exist(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify membership")
		}
		if !exists {
			return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this organization")
		}

		// Check project access for private projects
		if proj.IsPrivate {
			exists, err := h.client.ProjectMember.Query().
				Where(
					projectmember.UserIDEQ(userID),
					projectmember.ProjectIDEQ(projectID),
				).
				Exist(ctx)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify project access")
			}
			if !exists {
				return echo.NewHTTPError(http.StatusForbidden, "you do not have access to this project")
			}
		}

		update.SetLastProjectID(projectID)
		update.SetLastOrgID(proj.OrganizationID)
	}

	_, err := update.Save(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update context")
	}

	return c.NoContent(http.StatusNoContent)
}

