package auth

import (
	"strings"

	"github.com/StackCatalyst/common-lib/pkg/errors"
)

// Permission represents an action on a resource
type Permission string

// Resource represents a protected resource
type Resource string

// Role represents a set of permissions
type Role string

// Action represents the type of operation
type Action string

const (
	// Common actions
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionList   Action = "list"
	ActionWrite  Action = "write"

	// Special actions
	ActionAll Action = "*"
)

// BuildPermission creates a permission string from resource and action
func BuildPermission(resource Resource, action Action) Permission {
	return Permission(string(resource) + ":" + string(action))
}

// RBAC manages role-based access control
type RBAC struct {
	// rolePermissions maps roles to their permissions
	rolePermissions map[Role]map[Permission]bool
	// roleHierarchy maps roles to their parent roles
	roleHierarchy map[Role][]Role
}

// NewRBAC creates a new RBAC manager
func NewRBAC() *RBAC {
	return &RBAC{
		rolePermissions: make(map[Role]map[Permission]bool),
		roleHierarchy:   make(map[Role][]Role),
	}
}

// AddRole adds a new role with optional parent roles
func (r *RBAC) AddRole(role Role, parents ...Role) error {
	if _, exists := r.rolePermissions[role]; exists {
		return errors.New(errors.ErrValidation, "role already exists")
	}

	r.rolePermissions[role] = make(map[Permission]bool)
	if len(parents) > 0 {
		r.roleHierarchy[role] = parents
	}

	return nil
}

// AddPermission adds permissions to a role
func (r *RBAC) AddPermission(role Role, permissions ...Permission) error {
	perms, exists := r.rolePermissions[role]
	if !exists {
		return errors.New(errors.ErrNotFound, "role not found")
	}

	for _, perm := range permissions {
		perms[perm] = true
	}

	return nil
}

// RemovePermission removes permissions from a role
func (r *RBAC) RemovePermission(role Role, permissions ...Permission) error {
	perms, exists := r.rolePermissions[role]
	if !exists {
		return errors.New(errors.ErrNotFound, "role not found")
	}

	for _, perm := range permissions {
		delete(perms, perm)
	}

	return nil
}

// HasPermission checks if a role has a specific permission
func (r *RBAC) HasPermission(role Role, permission Permission) bool {
	// Check direct permissions
	if r.hasDirectPermission(role, permission) {
		return true
	}

	// Check parent roles recursively
	return r.hasParentPermission(role, permission)
}

// HasRole checks if a user has a specific role
func (r *RBAC) HasRole(userRoles []string, role Role) bool {
	for _, ur := range userRoles {
		if Role(ur) == role {
			return true
		}
	}
	return false
}

// IsAllowed checks if a user has permission to perform an action
func (r *RBAC) IsAllowed(userRoles []string, resource Resource, action Action) bool {
	permission := BuildPermission(resource, action)

	// Check each role the user has
	for _, roleStr := range userRoles {
		role := Role(roleStr)
		if r.HasPermission(role, permission) || r.HasPermission(role, BuildPermission(resource, ActionAll)) {
			return true
		}
	}

	return false
}

func (r *RBAC) hasDirectPermission(role Role, permission Permission) bool {
	perms, exists := r.rolePermissions[role]
	if !exists {
		return false
	}

	// Check for exact permission match
	if perms[permission] {
		return true
	}

	// Check for wildcard permissions
	parts := strings.Split(string(permission), ":")
	if len(parts) == 2 {
		wildcardPerm := Permission(parts[0] + ":*")
		return perms[wildcardPerm]
	}

	return false
}

func (r *RBAC) hasParentPermission(role Role, permission Permission) bool {
	parents, exists := r.roleHierarchy[role]
	if !exists {
		return false
	}

	for _, parent := range parents {
		if r.hasDirectPermission(parent, permission) {
			return true
		}
		if r.hasParentPermission(parent, permission) {
			return true
		}
	}

	return false
}
