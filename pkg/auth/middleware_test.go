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

func setupTestMiddleware(t *testing.T) (*TokenManager, *RBAC) {
	cfg := DefaultConfig()
	cfg.Token.AccessTokenSecret = "test-access-secret"
	cfg.Token.RefreshTokenSecret = "test-refresh-secret"
	cfg.RBAC.DefaultRole = "user"
	cfg.RBAC.SuperAdminRole = "admin"

	tm, err := NewTokenManager(cfg, newTestMetricsReporter())
	require.NoError(t, err)
	rbac := NewRBAC()
	return tm, rbac
}

func setupTestRouter(tm *TokenManager, rbac *RBAC, t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Protected route with auth
	r.GET("/protected", AuthMiddleware(tm), func(c *gin.Context) {
		userID, err := GetUserID(c.Request.Context())
		require.NoError(t, err)
		assert.Equal(t, "user123", userID)

		roles, err := GetUserRoles(c.Request.Context())
		require.NoError(t, err)
		assert.Equal(t, []string{"admin"}, roles)

		c.Status(http.StatusOK)
	})

	// Admin route with role check
	if rbac != nil {
		r.GET("/admin", AuthMiddleware(tm), RequireRole(rbac, "admin"), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		r.GET("/users/write", AuthMiddleware(tm), RequirePermission(rbac, "users", "write"), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
	}

	return r
}

func TestAuthMiddleware(t *testing.T) {
	// Setup
	tm, _ := setupTestMiddleware(t)
	router := setupTestRouter(tm, nil, t)

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
	}{
		{
			name: "valid token",
			setupAuth: func() string {
				token, err := tm.GenerateAccessToken("user123", []string{"admin"})
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
				return "Bearer invalid-token"
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid bearer format",
			setupAuth: func() string {
				token, err := tm.GenerateAccessToken("user123", []string{"admin"})
				require.NoError(t, err)
				return "Invalid " + token
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/protected", nil)
			if auth := tt.setupAuth(); auth != "" {
				req.Header.Set("Authorization", auth)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Test the middleware
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRoleMiddleware(t *testing.T) {
	// Setup
	tm, rbac := setupTestMiddleware(t)

	// Configure RBAC
	rbac.AddRole("admin")
	rbac.AddRole("user")

	// Create router
	router := setupTestRouter(tm, rbac, t)

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
	}{
		{
			name: "admin access granted",
			setupAuth: func() string {
				token, err := tm.GenerateAccessToken("user123", []string{"admin"})
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "admin access denied",
			setupAuth: func() string {
				token, err := tm.GenerateAccessToken("user123", []string{"user"})
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/admin", nil)
			if auth := tt.setupAuth(); auth != "" {
				req.Header.Set("Authorization", auth)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Test the middleware
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPermissionMiddleware(t *testing.T) {
	// Setup
	tm, rbac := setupTestMiddleware(t)

	// Configure RBAC with proper hierarchy
	require.NoError(t, rbac.AddRole("user"))
	require.NoError(t, rbac.AddRole("admin", "user")) // admin inherits from user
	require.NoError(t, rbac.AddPermission("admin", BuildPermission("users", "write")))
	require.NoError(t, rbac.AddPermission("user", BuildPermission("users", "read")))

	// Create router
	router := setupTestRouter(tm, rbac, t)

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
	}{
		{
			name: "user with permission",
			setupAuth: func() string {
				token, err := tm.GenerateAccessToken("user123", []string{"admin"})
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "user without permission",
			setupAuth: func() string {
				token, err := tm.GenerateAccessToken("user123", []string{"user"})
				require.NoError(t, err)
				return "Bearer " + token
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/users/write", nil)
			if auth := tt.setupAuth(); auth != "" {
				req.Header.Set("Authorization", auth)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Test the middleware
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
