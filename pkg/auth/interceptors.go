package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthUnaryInterceptor creates a gRPC unary interceptor for JWT authentication
func AuthUnaryInterceptor(tm *TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract token from metadata
		token, err := extractToken(ctx)
		if err != nil {
			return nil, err
		}

		// Validate token
		claims, err := tm.ValidateAccessToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Add claims to context
		newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)
		newCtx = context.WithValue(newCtx, UserRolesKey, claims.Roles)

		return handler(newCtx, req)
	}
}

// RBACUnaryInterceptor creates a gRPC unary interceptor for RBAC
func RBACUnaryInterceptor(rbac *RBAC, resource Resource, action Action) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		roles, err := GetUserRoles(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing user roles: %v", err)
		}

		if !rbac.IsAllowed(roles, resource, action) {
			return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
		}

		return handler(ctx, req)
	}
}

// AuthStreamInterceptor creates a gRPC stream interceptor for JWT authentication
func AuthStreamInterceptor(tm *TokenManager) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Extract token from metadata
		token, err := extractToken(ss.Context())
		if err != nil {
			return err
		}

		// Validate token
		claims, err := tm.ValidateAccessToken(token)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Create new context with claims
		newCtx := context.WithValue(ss.Context(), UserIDKey, claims.UserID)
		newCtx = context.WithValue(newCtx, UserRolesKey, claims.Roles)

		// Wrap ServerStream to use new context
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrappedStream)
	}
}

// RBACStreamInterceptor creates a gRPC stream interceptor for RBAC
func RBACStreamInterceptor(rbac *RBAC, resource Resource, action Action) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		roles, err := GetUserRoles(ss.Context())
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "missing user roles: %v", err)
		}

		if !rbac.IsAllowed(roles, resource, action) {
			return status.Error(codes.PermissionDenied, "insufficient permissions")
		}

		return handler(srv, ss)
	}
}

// wrappedServerStream wraps grpc.ServerStream to modify context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// extractToken extracts JWT token from gRPC metadata
func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != BearerSchema {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return parts[1], nil
}
