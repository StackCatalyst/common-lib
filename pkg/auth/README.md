# Authentication Package

The `auth` package provides a comprehensive authentication and authorization solution for both HTTP and gRPC services. It includes JWT token management, Role-Based Access Control (RBAC), HTTP middleware, and gRPC interceptors.

## Features

- JWT token management with access and refresh tokens
- Role-Based Access Control (RBAC) with role hierarchy
- HTTP middleware for Gin framework
- gRPC interceptors for authentication and authorization
- Support for both unary and streaming gRPC calls

## Installation

```bash
go get github.com/StackCatalyst/common-lib/pkg/auth
```

## Quick Start

### 1. Token Management

```go
// Initialize token manager
tm, err := auth.NewTokenManager(auth.TokenManagerConfig{
    AccessSecret:  "your-access-secret",
    RefreshSecret: "your-refresh-secret",
})
if err != nil {
    log.Fatal(err)
}

// Generate tokens
accessToken, err := tm.GenerateAccessToken("user123", []string{"admin"})
refreshToken, err := tm.GenerateRefreshToken("user123")

// Validate tokens
claims, err := tm.ValidateAccessToken(accessToken)
```

### 2. RBAC Setup

```go
// Initialize RBAC
rbac := auth.NewRBAC()

// Add roles
rbac.AddRole(auth.Role("admin"))
rbac.AddRole(auth.Role("user"), auth.Role("admin")) // user inherits from admin

// Add permissions
rbac.AddPermission(auth.Role("admin"), 
    auth.BuildPermission(auth.Resource("documents"), auth.ActionWrite))
rbac.AddPermission(auth.Role("user"), 
    auth.BuildPermission(auth.Resource("documents"), auth.ActionRead))

// Check permissions
allowed := rbac.IsAllowed([]string{"admin"}, auth.Resource("documents"), auth.ActionWrite)
```

### 3. HTTP Middleware (Gin)

```go
router := gin.New()

// Add authentication middleware
router.Use(auth.AuthMiddleware(tm))

// Protected routes with role and permission checks
protected := router.Group("/api")
{
    protected.GET("/docs",
        auth.RequireRole(rbac, auth.Role("user")),
        auth.RequirePermission(rbac, auth.Resource("documents"), auth.ActionRead),
        handleDocs,
    )
    
    protected.POST("/docs",
        auth.RequireRole(rbac, auth.Role("admin")),
        auth.RequirePermission(rbac, auth.Resource("documents"), auth.ActionWrite),
        handleCreateDoc,
    )
}
```

### 4. gRPC Interceptors

```go
// Server setup
server := grpc.NewServer(
    grpc.UnaryInterceptor(auth.AuthUnaryInterceptor(tm)),
    grpc.StreamInterceptor(auth.AuthStreamInterceptor(tm)),
)

// Add RBAC checks to specific methods
docService := &DocumentService{
    rbac: rbac,
}

// Implement your gRPC service methods with RBAC
func (s *DocumentService) CreateDocument(ctx context.Context, req *pb.CreateDocumentRequest) (*pb.Document, error) {
    // Check permissions using the RBAC interceptor
    if err := auth.RBACUnaryInterceptor(s.rbac, auth.Resource("documents"), auth.ActionWrite)(ctx, req, nil, nil); err != nil {
        return nil, err
    }
    
    // Your implementation here
}
```

## Common Patterns

### 1. Getting User Information

```go
// From HTTP context
func handleRequest(c *gin.Context) {
    userID, err := auth.GetUserID(c.Request.Context())
    roles, err := auth.GetUserRoles(c.Request.Context())
}

// From gRPC context
func (s *Service) HandleRPC(ctx context.Context, req interface{}) {
    userID, err := auth.GetUserID(ctx)
    roles, err := auth.GetUserRoles(ctx)
}
```

### 2. Role Hierarchy

```go
rbac := auth.NewRBAC()

// Create role hierarchy
rbac.AddRole(auth.Role("super_admin"))
rbac.AddRole(auth.Role("admin"), auth.Role("super_admin"))  // admin inherits from super_admin
rbac.AddRole(auth.Role("user"), auth.Role("admin"))        // user inherits from admin

// Add permissions at different levels
rbac.AddPermission(auth.Role("super_admin"), 
    auth.BuildPermission(auth.Resource("system"), auth.ActionAll))
rbac.AddPermission(auth.Role("admin"), 
    auth.BuildPermission(auth.Resource("users"), auth.ActionAll))
rbac.AddPermission(auth.Role("user"), 
    auth.BuildPermission(auth.Resource("documents"), auth.ActionRead))
```

### 3. Custom Actions and Resources

```go
// Define custom resources
const (
    ResourceUsers     auth.Resource = "users"
    ResourceDocuments auth.Resource = "documents"
    ResourceProjects  auth.Resource = "projects"
)

// Define custom actions
const (
    ActionApprove auth.Action = "approve"
    ActionReject  auth.Action = "reject"
    ActionArchive auth.Action = "archive"
)

// Use in permissions
rbac.AddPermission(auth.Role("manager"),
    auth.BuildPermission(ResourceProjects, ActionApprove),
    auth.BuildPermission(ResourceProjects, ActionReject),
)
```

## Best Practices

1. **Token Security**
   - Store secrets securely (e.g., environment variables, secret management service)
   - Use strong, unique secrets for access and refresh tokens
   - Implement token refresh flow for better security

2. **RBAC Design**
   - Plan your role hierarchy carefully
   - Use specific permissions instead of generic ones
   - Implement the principle of least privilege

3. **Error Handling**
   - Always check for errors when validating tokens
   - Provide clear error messages for authorization failures
   - Log security-related events appropriately

4. **Performance**
   - Cache RBAC decisions when possible
   - Use appropriate token expiration times
   - Consider implementing rate limiting for token operations

## Contributing

Please read our [contributing guidelines](../CONTRIBUTING.md) before submitting pull requests.

## License

This package is part of the StackCatalyst Common Library and is licensed under the same terms. 