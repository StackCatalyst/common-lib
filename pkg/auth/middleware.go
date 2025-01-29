package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/StackCatalyst/common-lib/pkg/errors"
	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UserRolesKey is the context key for user roles
	UserRolesKey contextKey = "user_roles"
	// AuthHeaderKey is the authorization header key
	AuthHeaderKey = "Authorization"
	// BearerSchema is the bearer token schema
	BearerSchema = "Bearer"
)

// AuthMiddleware creates a Gin middleware for JWT authentication
func AuthMiddleware(tm *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from header
		authHeader := c.GetHeader(AuthHeaderKey)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "no authorization header",
			})
			return
		}

		// Check bearer schema
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != BearerSchema {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		// Validate token
		claims, err := tm.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Set claims in context
		ctx := context.WithValue(c.Request.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireRole creates a Gin middleware for role-based authorization
func RequireRole(rbac *RBAC, role Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Request.Context().Value(UserRolesKey).([]string)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "user roles not found in context",
			})
			return
		}

		if !rbac.HasRole(userRoles, role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// RequirePermission creates a Gin middleware for permission-based authorization
func RequirePermission(rbac *RBAC, resource Resource, action Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Request.Context().Value(UserRolesKey).([]string)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "user roles not found in context",
			})
			return
		}

		if !rbac.IsAllowed(userRoles, resource, action) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", errors.New(errors.ErrUnauthorized, "user ID not found in context")
	}
	return userID, nil
}

// GetUserRoles retrieves the user roles from the context
func GetUserRoles(ctx context.Context) ([]string, error) {
	roles, ok := ctx.Value(UserRolesKey).([]string)
	if !ok {
		return nil, errors.New(errors.ErrUnauthorized, "user roles not found in context")
	}
	return roles, nil
}
