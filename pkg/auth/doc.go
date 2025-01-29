/*
Package auth provides a comprehensive authentication and authorization solution for both HTTP and gRPC services.

The package includes JWT token management, Role-Based Access Control (RBAC), HTTP middleware for the Gin framework,
and gRPC interceptors for both unary and streaming calls.

# Token Management

The package provides JWT token management with support for both access and refresh tokens:

	tm, err := auth.NewTokenManager(auth.TokenManagerConfig{
		AccessSecret:  "your-access-secret",
		RefreshSecret: "your-refresh-secret",
	})

	// Generate tokens
	accessToken, err := tm.GenerateAccessToken("user123", []string{"admin"})
	refreshToken, err := tm.GenerateRefreshToken("user123")

# Role-Based Access Control

The RBAC system supports role hierarchy and fine-grained permissions:

	rbac := auth.NewRBAC()

	// Add roles with inheritance
	rbac.AddRole(auth.Role("admin"))
	rbac.AddRole(auth.Role("user"), auth.Role("admin"))

	// Add permissions
	rbac.AddPermission(auth.Role("admin"),
		auth.BuildPermission(auth.Resource("documents"), auth.ActionWrite))

# HTTP Middleware

The package provides Gin middleware for authentication and authorization:

	router := gin.New()
	router.Use(auth.AuthMiddleware(tm))

	protected := router.Group("/api")
	protected.GET("/docs",
		auth.RequireRole(rbac, auth.Role("user")),
		auth.RequirePermission(rbac, auth.Resource("documents"), auth.ActionRead),
		handleDocs,
	)

gRPC Interceptors

For gRPC services, the package provides both unary and stream interceptors:

	server := grpc.NewServer(
		grpc.UnaryInterceptor(auth.AuthUnaryInterceptor(tm)),
		grpc.StreamInterceptor(auth.AuthStreamInterceptor(tm)),
	)

# Context Helpers

Helper functions are provided to access authentication information from context:

	userID, err := auth.GetUserID(ctx)
	roles, err := auth.GetUserRoles(ctx)

# Error Handling

The package uses the common-lib error package for consistent error handling:

	if err != nil {
		switch {
		case errors.Is(err, errors.ErrUnauthorized):
			// Handle unauthorized access
		case errors.Is(err, errors.ErrValidation):
			// Handle validation errors
		default:
			// Handle other errors
		}
	}
*/
package auth
