package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Test roles
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"

	// Test resources
	ResourceUser     Resource = "user"
	ResourceProject  Resource = "project"
	ResourceDocument Resource = "document"
)

func TestRBACRoleManagement(t *testing.T) {
	rbac := NewRBAC()

	// Test adding roles
	err := rbac.AddRole(RoleAdmin)
	require.NoError(t, err)

	err = rbac.AddRole(RoleUser)
	require.NoError(t, err)

	// Test adding duplicate role
	err = rbac.AddRole(RoleAdmin)
	assert.Error(t, err)

	// Test role hierarchy
	err = rbac.AddRole(RoleGuest, RoleUser)
	require.NoError(t, err)
}

func TestRBACPermissionManagement(t *testing.T) {
	rbac := NewRBAC()

	// Setup roles
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddRole(RoleUser))

	// Test adding permissions
	userReadPerm := BuildPermission(ResourceUser, ActionRead)
	userWritePerm := BuildPermission(ResourceUser, ActionUpdate)

	err := rbac.AddPermission(RoleUser, userReadPerm)
	require.NoError(t, err)

	err = rbac.AddPermission(RoleAdmin, userReadPerm, userWritePerm)
	require.NoError(t, err)

	// Test adding permission to non-existent role
	err = rbac.AddPermission(Role("non-existent"), userReadPerm)
	assert.Error(t, err)

	// Test removing permissions
	err = rbac.RemovePermission(RoleUser, userReadPerm)
	require.NoError(t, err)
	assert.False(t, rbac.HasPermission(RoleUser, userReadPerm))
}

func TestRBACPermissionChecking(t *testing.T) {
	rbac := NewRBAC()

	// Setup roles and permissions
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddRole(RoleUser))
	require.NoError(t, rbac.AddRole(RoleGuest, RoleUser)) // Guest inherits from User

	// Add permissions
	projectAllPerm := BuildPermission(ResourceProject, ActionAll)
	projectReadPerm := BuildPermission(ResourceProject, ActionRead)
	projectUpdatePerm := BuildPermission(ResourceProject, ActionUpdate)

	require.NoError(t, rbac.AddPermission(RoleAdmin, projectAllPerm))
	require.NoError(t, rbac.AddPermission(RoleUser, projectReadPerm, projectUpdatePerm))

	// Test direct permissions
	assert.True(t, rbac.HasPermission(RoleAdmin, projectAllPerm))
	assert.True(t, rbac.HasPermission(RoleUser, projectReadPerm))
	assert.False(t, rbac.HasPermission(RoleUser, projectAllPerm))

	// Test inherited permissions
	assert.True(t, rbac.HasPermission(RoleGuest, projectReadPerm))
	assert.True(t, rbac.HasPermission(RoleGuest, projectUpdatePerm))

	// Test wildcard permissions
	assert.True(t, rbac.HasPermission(RoleAdmin, projectReadPerm))
	assert.True(t, rbac.HasPermission(RoleAdmin, projectUpdatePerm))
}

func TestRBACIsAllowed(t *testing.T) {
	rbac := NewRBAC()

	// Setup roles and permissions
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddRole(RoleUser))

	// Add permissions
	docReadPerm := BuildPermission(ResourceDocument, ActionRead)
	docAllPerm := BuildPermission(ResourceDocument, ActionAll)

	require.NoError(t, rbac.AddPermission(RoleAdmin, docAllPerm))
	require.NoError(t, rbac.AddPermission(RoleUser, docReadPerm))

	// Test user permissions
	userRoles := []string{"user"}
	assert.True(t, rbac.IsAllowed(userRoles, ResourceDocument, ActionRead))
	assert.False(t, rbac.IsAllowed(userRoles, ResourceDocument, ActionUpdate))

	// Test admin permissions
	adminRoles := []string{"admin"}
	assert.True(t, rbac.IsAllowed(adminRoles, ResourceDocument, ActionRead))
	assert.True(t, rbac.IsAllowed(adminRoles, ResourceDocument, ActionUpdate))
	assert.True(t, rbac.IsAllowed(adminRoles, ResourceDocument, ActionDelete))

	// Test multiple roles
	multiRoles := []string{"user", "admin"}
	assert.True(t, rbac.IsAllowed(multiRoles, ResourceDocument, ActionAll))
}

func TestRBACHasRole(t *testing.T) {
	rbac := NewRBAC()

	// Test role checking
	userRoles := []string{"admin", "user"}
	assert.True(t, rbac.HasRole(userRoles, RoleAdmin))
	assert.True(t, rbac.HasRole(userRoles, RoleUser))
	assert.False(t, rbac.HasRole(userRoles, RoleGuest))
}
