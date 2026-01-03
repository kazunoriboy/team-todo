package handler

import "backend/ent/organizationmember"

// HasAdminPermission checks if the role has admin-level permission (owner or admin)
func HasAdminPermission(role organizationmember.Role) bool {
	return role == organizationmember.RoleOwner || role == organizationmember.RoleAdmin
}

// IsOwner checks if the role is owner
func IsOwner(role organizationmember.Role) bool {
	return role == organizationmember.RoleOwner
}

