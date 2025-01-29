package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(tm *TokenManager, rbac *RBAC) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add test routes
	r.GET("/protected", AuthMiddleware(tm), func(c *gin.Context) {
		userID, err := GetUserID(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	r.GET("/admin", AuthMiddleware(tm), RequireRole(rbac, RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	r.GET("/documents", AuthMiddleware(tm), RequirePermission(rbac, ResourceDocument, ActionRead), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "document access granted"})
	})

	return r
}

func TestAuthMiddleware(t *testing.T) {
	// Setup
	tm, err := NewTokenManager(TokenManagerConfig{
		AccessSecret:  "test-secret",
		RefreshSecret: "refresh-secret",
	})
	require.NoError(t, err)

	router := setupTestRouter(tm, nil)

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
	}{
		{
			name: "valid token",
			setupAuth: func() string {
				token, err := tm.GenerateAccessToken("test-user", []string{"user"})
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "no auth header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token format",
			setupAuth: func() string {
				return "Invalid token"
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid bearer format",
			setupAuth: func() string {
				return "Bearer"
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			if auth := tt.setupAuth(); auth != "" {
				req.Header.Set("Authorization", auth)
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRoleMiddleware(t *testing.T) {
	// Setup
	tm, err := NewTokenManager(TokenManagerConfig{
		AccessSecret:  "test-secret",
		RefreshSecret: "refresh-secret",
	})
	require.NoError(t, err)

	rbac := NewRBAC()
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddRole(RoleUser))

	router := setupTestRouter(tm, rbac)

	tests := []struct {
		name           string
		userRoles      []string
		expectedStatus int
	}{
		{
			name:           "admin access granted",
			userRoles:      []string{"admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "admin access denied",
			userRoles:      []string{"user"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tm.GenerateAccessToken("test-user", tt.userRoles)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPermissionMiddleware(t *testing.T) {
	// Setup
	tm, err := NewTokenManager(TokenManagerConfig{
		AccessSecret:  "test-secret",
		RefreshSecret: "refresh-secret",
	})
	require.NoError(t, err)

	rbac := NewRBAC()
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddRole(RoleUser))
	require.NoError(t, rbac.AddPermission(RoleUser, BuildPermission(ResourceDocument, ActionRead)))

	router := setupTestRouter(tm, rbac)

	tests := []struct {
		name           string
		userRoles      []string
		expectedStatus int
	}{
		{
			name:           "user with permission",
			userRoles:      []string{"user"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user without permission",
			userRoles:      []string{"guest"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tm.GenerateAccessToken("test-user", tt.userRoles)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/documents", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestContextHelpers(t *testing.T) {
	// Test GetUserID
	t.Run("get user id", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, "test-user")
		userID, err := GetUserID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "test-user", userID)
	})

	t.Run("get user id - not found", func(t *testing.T) {
		ctx := context.Background()
		userID, err := GetUserID(ctx)
		assert.Error(t, err)
		assert.Empty(t, userID)
	})

	// Test GetUserRoles
	t.Run("get user roles", func(t *testing.T) {
		roles := []string{"admin", "user"}
		ctx := context.WithValue(context.Background(), UserRolesKey, roles)
		userRoles, err := GetUserRoles(ctx)
		assert.NoError(t, err)
		assert.Equal(t, roles, userRoles)
	})

	t.Run("get user roles - not found", func(t *testing.T) {
		ctx := context.Background()
		userRoles, err := GetUserRoles(ctx)
		assert.Error(t, err)
		assert.Nil(t, userRoles)
	})
}
