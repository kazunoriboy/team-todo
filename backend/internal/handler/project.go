package handler

import (
	"net/http"
	"time"

	"backend/ent"
	"backend/ent/organization"
	"backend/ent/organizationmember"
	"backend/ent/project"
	"backend/ent/projectmember"
	"backend/internal/auth"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ProjectHandler handles project-related requests
type ProjectHandler struct {
	client *ent.Client
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(client *ent.Client) *ProjectHandler {
	return &ProjectHandler{client: client}
}

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	Name      string `json:"name" validate:"required"`
	IsPrivate bool   `json:"is_private"`
}

// ProjectResponse represents the project data in responses
type ProjectResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	IsPrivate      bool      `json:"is_private"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Permission     string    `json:"permission,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ProjectMemberResponse represents a project member in responses
type ProjectMemberResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Permission  string    `json:"permission"`
	JoinedAt    time.Time `json:"joined_at"`
}

// CreateProject creates a new project in an organization
func (h *ProjectHandler) CreateProject(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	slug := c.Param("slug")
	if slug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "organization slug is required")
	}

	var req CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
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

	// Only owners and admins can create projects
	if !HasAdminPermission(membership.Role) {
		return echo.NewHTTPError(http.StatusForbidden, "only owners and admins can create projects")
	}

	// Create project in a transaction
	tx, err := h.client.Tx(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to start transaction")
	}

	proj, err := tx.Project.Create().
		SetName(req.Name).
		SetOrganizationID(org.ID).
		SetIsPrivate(req.IsPrivate).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create project")
	}

	// Add creator as edit member if private
	if req.IsPrivate {
		_, err = tx.ProjectMember.Create().
			SetUserID(userID).
			SetProjectID(proj.ID).
			SetPermission(projectmember.PermissionEdit).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to add project member")
		}
	}

	// Update user's last accessed project
	_, err = tx.User.UpdateOneID(userID).
		SetLastProjectID(proj.ID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user context")
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction")
	}

	return c.JSON(http.StatusCreated, ProjectResponse{
		ID:             proj.ID,
		Name:           proj.Name,
		IsPrivate:      proj.IsPrivate,
		OrganizationID: proj.OrganizationID,
		Permission:     "edit",
		CreatedAt:      proj.CreatedAt,
	})
}

// ListProjects lists all projects in an organization
func (h *ProjectHandler) ListProjects(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	slug := c.Param("slug")
	if slug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "organization slug is required")
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
	_, err = h.client.OrganizationMember.Query().
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

	// Get all projects in the organization
	projects, err := h.client.Project.Query().
		Where(project.OrganizationIDEQ(org.ID)).
		Order(ent.Asc(project.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list projects")
	}

	// Get user's project memberships for private projects
	projectMemberships, err := h.client.ProjectMember.Query().
		Where(projectmember.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get project memberships")
	}

	membershipMap := make(map[uuid.UUID]string)
	for _, pm := range projectMemberships {
		membershipMap[pm.ProjectID] = string(pm.Permission)
	}

	// Filter and build response
	var result []ProjectResponse
	for _, p := range projects {
		if p.IsPrivate {
			// Only include if user is a member
			if perm, ok := membershipMap[p.ID]; ok {
				result = append(result, ProjectResponse{
					ID:             p.ID,
					Name:           p.Name,
					IsPrivate:      p.IsPrivate,
					OrganizationID: p.OrganizationID,
					Permission:     perm,
					CreatedAt:      p.CreatedAt,
				})
			}
		} else {
			// Public project - include for all org members
			perm := "view"
			if mp, ok := membershipMap[p.ID]; ok {
				perm = mp
			}
			result = append(result, ProjectResponse{
				ID:             p.ID,
				Name:           p.Name,
				IsPrivate:      p.IsPrivate,
				OrganizationID: p.OrganizationID,
				Permission:     perm,
				CreatedAt:      p.CreatedAt,
			})
		}
	}

	return c.JSON(http.StatusOK, result)
}

// GetProject gets a project by ID
func (h *ProjectHandler) GetProject(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	slug := c.Param("slug")
	projectIDStr := c.Param("project_id")
	if slug == "" || projectIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "organization slug and project_id are required")
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id format")
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

	// Check org membership
	_, err = h.client.OrganizationMember.Query().
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

	// Get project
	proj, err := h.client.Project.Query().
		Where(
			project.IDEQ(projectID),
			project.OrganizationIDEQ(org.ID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get project")
	}

	// Check project access for private projects
	permission := "view"
	if proj.IsPrivate {
		pm, err := h.client.ProjectMember.Query().
			Where(
				projectmember.UserIDEQ(userID),
				projectmember.ProjectIDEQ(proj.ID),
			).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return echo.NewHTTPError(http.StatusForbidden, "you do not have access to this project")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to check project membership")
		}
		permission = string(pm.Permission)
	} else {
		// Check if user has explicit membership for higher permission
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

	// Update user's last accessed project
	_, _ = h.client.User.UpdateOneID(userID).
		SetLastProjectID(proj.ID).
		SetLastOrgID(org.ID).
		Save(ctx)

	return c.JSON(http.StatusOK, ProjectResponse{
		ID:             proj.ID,
		Name:           proj.Name,
		IsPrivate:      proj.IsPrivate,
		OrganizationID: proj.OrganizationID,
		Permission:     permission,
		CreatedAt:      proj.CreatedAt,
	})
}

// AddProjectMember adds a member to a project
func (h *ProjectHandler) AddProjectMember(c echo.Context) error {
	userID, ok := auth.GetUserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	slug := c.Param("slug")
	projectIDStr := c.Param("project_id")
	if slug == "" || projectIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "organization slug and project_id are required")
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id format")
	}

	var req struct {
		UserID     string `json:"user_id" validate:"required"`
		Permission string `json:"permission" validate:"required,oneof=edit view"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id format")
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

	// Check if requester is org owner/admin
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
		return echo.NewHTTPError(http.StatusForbidden, "only owners and admins can manage project members")
	}

	// Check target user is org member
	_, err = h.client.OrganizationMember.Query().
		Where(
			organizationmember.UserIDEQ(targetUserID),
			organizationmember.OrganizationIDEQ(org.ID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "target user is not a member of this organization")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check target user membership")
	}

	// Get project
	proj, err := h.client.Project.Query().
		Where(
			project.IDEQ(projectID),
			project.OrganizationIDEQ(org.ID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "project not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get project")
	}

	// Check if already a member
	exists, err := h.client.ProjectMember.Query().
		Where(
			projectmember.UserIDEQ(targetUserID),
			projectmember.ProjectIDEQ(proj.ID),
		).
		Exist(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check existing membership")
	}
	if exists {
		return echo.NewHTTPError(http.StatusConflict, "user is already a member of this project")
	}

	// Add member
	permission := projectmember.PermissionView
	if req.Permission == "edit" {
		permission = projectmember.PermissionEdit
	}

	pm, err := h.client.ProjectMember.Create().
		SetUserID(targetUserID).
		SetProjectID(proj.ID).
		SetPermission(permission).
		Save(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add project member")
	}

	targetUser, _ := h.client.User.Get(ctx, targetUserID)

	return c.JSON(http.StatusCreated, ProjectMemberResponse{
		UserID:      targetUserID,
		Email:       targetUser.Email,
		DisplayName: targetUser.DisplayName,
		Permission:  string(pm.Permission),
		JoinedAt:    pm.CreatedAt,
	})
}

