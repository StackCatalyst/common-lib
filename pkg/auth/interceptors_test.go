package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Mock gRPC server stream
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func TestAuthUnaryInterceptor(t *testing.T) {
	// Setup
	tm, err := NewTokenManager(TokenManagerConfig{
		AccessSecret:  "test-secret",
		RefreshSecret: "refresh-secret",
	})
	require.NoError(t, err)

	interceptor := AuthUnaryInterceptor(tm)

	// Test cases
	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedError  bool
		expectedCode   codes.Code
		expectedUserID string
	}{
		{
			name: "valid token",
			setupContext: func() context.Context {
				token, _ := tm.GenerateAccessToken("test-user", []string{"user"})
				md := metadata.New(map[string]string{
					"authorization": "Bearer " + token,
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError:  false,
			expectedUserID: "test-user",
		},
		{
			name: "missing metadata",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: true,
			expectedCode:  codes.Unauthenticated,
		},
		{
			name: "invalid token",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer invalid-token",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError: true,
			expectedCode:  codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				if tt.expectedUserID != "" {
					userID, err := GetUserID(ctx)
					require.NoError(t, err)
					assert.Equal(t, tt.expectedUserID, userID)
				}
				return "response", nil
			}

			_, err := interceptor(ctx, nil, nil, handler)
			if tt.expectedError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRBACUnaryInterceptor(t *testing.T) {
	// Setup
	rbac := NewRBAC()
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddRole(RoleUser))
	require.NoError(t, rbac.AddPermission(RoleUser, BuildPermission(ResourceDocument, ActionRead)))

	interceptor := RBACUnaryInterceptor(rbac, ResourceDocument, ActionRead)

	// Test cases
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name: "allowed access",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), UserRolesKey, []string{"user"})
			},
			expectedError: false,
		},
		{
			name: "missing roles",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: true,
			expectedCode:  codes.Unauthenticated,
		},
		{
			name: "insufficient permissions",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), UserRolesKey, []string{"guest"})
			},
			expectedError: true,
			expectedCode:  codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			}

			_, err := interceptor(ctx, nil, nil, handler)
			if tt.expectedError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthStreamInterceptor(t *testing.T) {
	// Setup
	tm, err := NewTokenManager(TokenManagerConfig{
		AccessSecret:  "test-secret",
		RefreshSecret: "refresh-secret",
	})
	require.NoError(t, err)

	interceptor := AuthStreamInterceptor(tm)

	// Test cases
	tests := []struct {
		name           string
		setupStream    func() grpc.ServerStream
		expectedError  bool
		expectedCode   codes.Code
		expectedUserID string
	}{
		{
			name: "valid token",
			setupStream: func() grpc.ServerStream {
				token, _ := tm.GenerateAccessToken("test-user", []string{"user"})
				md := metadata.New(map[string]string{
					"authorization": "Bearer " + token,
				})
				ctx := metadata.NewIncomingContext(context.Background(), md)
				return &mockServerStream{ctx: ctx}
			},
			expectedError:  false,
			expectedUserID: "test-user",
		},
		{
			name: "missing metadata",
			setupStream: func() grpc.ServerStream {
				return &mockServerStream{ctx: context.Background()}
			},
			expectedError: true,
			expectedCode:  codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tt.setupStream()
			handler := func(srv interface{}, stream grpc.ServerStream) error {
				if tt.expectedUserID != "" {
					userID, err := GetUserID(stream.Context())
					require.NoError(t, err)
					assert.Equal(t, tt.expectedUserID, userID)
				}
				return nil
			}

			err := interceptor(nil, stream, nil, handler)
			if tt.expectedError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRBACStreamInterceptor(t *testing.T) {
	// Setup
	rbac := NewRBAC()
	require.NoError(t, rbac.AddRole(RoleAdmin))
	require.NoError(t, rbac.AddRole(RoleUser))
	require.NoError(t, rbac.AddPermission(RoleUser, BuildPermission(ResourceDocument, ActionRead)))

	interceptor := RBACStreamInterceptor(rbac, ResourceDocument, ActionRead)

	// Test cases
	tests := []struct {
		name          string
		setupStream   func() grpc.ServerStream
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name: "allowed access",
			setupStream: func() grpc.ServerStream {
				ctx := context.WithValue(context.Background(), UserRolesKey, []string{"user"})
				return &mockServerStream{ctx: ctx}
			},
			expectedError: false,
		},
		{
			name: "missing roles",
			setupStream: func() grpc.ServerStream {
				return &mockServerStream{ctx: context.Background()}
			},
			expectedError: true,
			expectedCode:  codes.Unauthenticated,
		},
		{
			name: "insufficient permissions",
			setupStream: func() grpc.ServerStream {
				ctx := context.WithValue(context.Background(), UserRolesKey, []string{"guest"})
				return &mockServerStream{ctx: ctx}
			},
			expectedError: true,
			expectedCode:  codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tt.setupStream()
			handler := func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			}

			err := interceptor(nil, stream, nil, handler)
			if tt.expectedError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
