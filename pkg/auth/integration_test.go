package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Mock service for integration testing
type mockService struct {
	userID string
	roles  []string
}

func (s *mockService) HandleHTTP(c *gin.Context) {
	userID, err := GetUserID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	roles, err := GetUserRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.userID = userID
	s.roles = roles
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "roles": roles})
}

func (s *mockService) HandleGRPC(ctx context.Context, req interface{}) (interface{}, error) {
	userID, err := GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	roles, err := GetUserRoles(ctx)
	if err != nil {
		return nil, err
	}
	s.userID = userID
	s.roles = roles
	return map[string]interface{}{"user_id": userID, "roles": roles}, nil
}

func setupIntegrationTokenManager(t *testing.T) *TokenManager {
	cfg := DefaultConfig()
	cfg.Token.AccessTokenSecret = "test-access-secret"
	cfg.Token.RefreshTokenSecret = "test-refresh-secret"
	cfg.RBAC.DefaultRole = "user"
	cfg.RBAC.SuperAdminRole = "admin"

	tm, err := NewTokenManager(cfg, newTestMetricsReporter())
	require.NoError(t, err)
	return tm
}

func TestIntegrationAuthFlow(t *testing.T) {
	// Setup
	tm := setupIntegrationTokenManager(t)

	rbac := NewRBAC()
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddRole(RoleUser))
	require.NoError(t, rbac.AddPermission(RoleUser, BuildPermission(ResourceDocument, ActionRead)))
	require.NoError(t, rbac.AddPermission(RoleAdmin, BuildPermission(ResourceDocument, ActionWrite)))

	// Generate test tokens
	adminToken, err := tm.GenerateAccessToken("admin-user", []string{string(RoleAdmin)})
	require.NoError(t, err)
	userToken, err := tm.GenerateAccessToken("regular-user", []string{string(RoleUser)})
	require.NoError(t, err)

	// Test HTTP and gRPC with the same token
	t.Run("admin access both HTTP and gRPC", func(t *testing.T) {
		mockSvc := &mockService{}

		// Test HTTP
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/admin/docs",
			AuthMiddleware(tm),
			RequireRole(rbac, RoleAdmin),
			RequirePermission(rbac, ResourceDocument, ActionWrite),
			mockSvc.HandleHTTP,
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/docs", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "admin-user", mockSvc.userID)
		assert.Contains(t, mockSvc.roles, string(RoleAdmin))

		// Test gRPC
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
			"authorization": "Bearer " + adminToken,
		}))

		unaryInterceptor := AuthUnaryInterceptor(tm)
		rbacInterceptor := RBACUnaryInterceptor(rbac, ResourceDocument, ActionWrite)

		resp, err := unaryInterceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return rbacInterceptor(ctx, req, nil, mockSvc.HandleGRPC)
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "admin-user", mockSvc.userID)
		assert.Contains(t, mockSvc.roles, string(RoleAdmin))
	})

	t.Run("user access with limited permissions", func(t *testing.T) {
		mockSvc := &mockService{}

		// Test HTTP
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/docs",
			AuthMiddleware(tm),
			RequireRole(rbac, RoleUser),
			RequirePermission(rbac, ResourceDocument, ActionRead),
			mockSvc.HandleHTTP,
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/docs", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "regular-user", mockSvc.userID)
		assert.Contains(t, mockSvc.roles, string(RoleUser))

		// Test gRPC
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
			"authorization": "Bearer " + userToken,
		}))

		unaryInterceptor := AuthUnaryInterceptor(tm)
		rbacInterceptor := RBACUnaryInterceptor(rbac, ResourceDocument, ActionRead)

		resp, err := unaryInterceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return rbacInterceptor(ctx, req, nil, mockSvc.HandleGRPC)
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "regular-user", mockSvc.userID)
		assert.Contains(t, mockSvc.roles, string(RoleUser))
	})

	t.Run("permission denied cases", func(t *testing.T) {
		mockSvc := &mockService{}

		// Test HTTP
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/admin/docs",
			AuthMiddleware(tm),
			RequireRole(rbac, RoleAdmin),
			RequirePermission(rbac, ResourceDocument, ActionWrite),
			mockSvc.HandleHTTP,
		)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/docs", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		// Test gRPC
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
			"authorization": "Bearer " + userToken,
		}))

		unaryInterceptor := AuthUnaryInterceptor(tm)
		rbacInterceptor := RBACUnaryInterceptor(rbac, ResourceDocument, ActionWrite)

		_, err := unaryInterceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return rbacInterceptor(ctx, req, nil, mockSvc.HandleGRPC)
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient permissions")
	})
}

func TestIntegrationStreamingAuthFlow(t *testing.T) {
	// Setup
	tm := setupIntegrationTokenManager(t)

	rbac := NewRBAC()
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddPermission(RoleAdmin, BuildPermission(ResourceDocument, ActionWrite)))

	// Generate test token
	adminToken, err := tm.GenerateAccessToken("admin-user", []string{string(RoleAdmin)})
	require.NoError(t, err)

	t.Run("streaming with valid permissions", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
			"authorization": "Bearer " + adminToken,
		}))

		stream := &mockServerStream{ctx: ctx}
		authInterceptor := AuthStreamInterceptor(tm)
		rbacInterceptor := RBACStreamInterceptor(rbac, ResourceDocument, ActionWrite)

		var handlerCalled bool
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			userID, err := GetUserID(stream.Context())
			require.NoError(t, err)
			assert.Equal(t, "admin-user", userID)

			roles, err := GetUserRoles(stream.Context())
			require.NoError(t, err)
			assert.Contains(t, roles, string(RoleAdmin))

			handlerCalled = true
			return nil
		}

		// Chain stream interceptors
		err := authInterceptor(nil, stream, nil, func(srv interface{}, stream grpc.ServerStream) error {
			return rbacInterceptor(srv, stream, nil, handler)
		})

		require.NoError(t, err)
		assert.True(t, handlerCalled, "Stream handler should have been called")
	})
}
