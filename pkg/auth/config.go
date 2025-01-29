package auth

import "time"

// Config holds authentication configuration settings
type Config struct {
	// Token settings
	Token struct {
		// AccessTokenDuration is the duration for which access tokens are valid
		AccessTokenDuration time.Duration `json:"access_token_duration" yaml:"access_token_duration"`
		// RefreshTokenDuration is the duration for which refresh tokens are valid
		RefreshTokenDuration time.Duration `json:"refresh_token_duration" yaml:"refresh_token_duration"`
		// AccessTokenSecret is the secret used to sign access tokens
		AccessTokenSecret string `json:"access_token_secret" yaml:"access_token_secret"`
		// RefreshTokenSecret is the secret used to sign refresh tokens
		RefreshTokenSecret string `json:"refresh_token_secret" yaml:"refresh_token_secret"`
	} `json:"token" yaml:"token"`

	// RBAC settings
	RBAC struct {
		// DefaultRole is the role assigned to new users
		DefaultRole string `json:"default_role" yaml:"default_role"`
		// SuperAdminRole is the role with all permissions
		SuperAdminRole string `json:"super_admin_role" yaml:"super_admin_role"`
		// RoleHierarchy defines the hierarchy of roles (higher roles inherit lower roles' permissions)
		RoleHierarchy map[string][]string `json:"role_hierarchy" yaml:"role_hierarchy"`
	} `json:"rbac" yaml:"rbac"`

	// Rate limiting settings
	RateLimit struct {
		// Enabled indicates whether rate limiting is enabled
		Enabled bool `json:"enabled" yaml:"enabled"`
		// RequestsPerSecond is the maximum number of requests allowed per second
		RequestsPerSecond int `json:"requests_per_second" yaml:"requests_per_second"`
		// BurstSize is the maximum number of requests allowed to burst
		BurstSize int `json:"burst_size" yaml:"burst_size"`
	} `json:"rate_limit" yaml:"rate_limit"`
}

// DefaultConfig returns the default authentication configuration
func DefaultConfig() Config {
	cfg := Config{}

	// Token defaults
	cfg.Token.AccessTokenDuration = 15 * time.Minute
	cfg.Token.RefreshTokenDuration = 24 * time.Hour
	cfg.Token.AccessTokenSecret = ""  // Must be provided
	cfg.Token.RefreshTokenSecret = "" // Must be provided

	// RBAC defaults
	cfg.RBAC.DefaultRole = "user"
	cfg.RBAC.SuperAdminRole = "admin"
	cfg.RBAC.RoleHierarchy = map[string][]string{
		"admin": {"user"},
		"user":  {},
	}

	// Rate limiting defaults
	cfg.RateLimit.Enabled = true
	cfg.RateLimit.RequestsPerSecond = 10
	cfg.RateLimit.BurstSize = 20

	return cfg
}

// Validate validates the authentication configuration
func (c *Config) Validate() error {
	if c.Token.AccessTokenDuration <= 0 {
		return newInvalidConfigError("access token duration must be positive")
	}
	if c.Token.RefreshTokenDuration <= 0 {
		return newInvalidConfigError("refresh token duration must be positive")
	}
	if c.Token.AccessTokenSecret == "" {
		return newInvalidConfigError("access token secret must be provided")
	}
	if c.Token.RefreshTokenSecret == "" {
		return newInvalidConfigError("refresh token secret must be provided")
	}
	if c.RBAC.DefaultRole == "" {
		return newInvalidConfigError("default role must be provided")
	}
	if c.RBAC.SuperAdminRole == "" {
		return newInvalidConfigError("super admin role must be provided")
	}
	if c.RateLimit.Enabled {
		if c.RateLimit.RequestsPerSecond <= 0 {
			return newInvalidConfigError("requests per second must be positive")
		}
		if c.RateLimit.BurstSize <= 0 {
			return newInvalidConfigError("burst size must be positive")
		}
	}
	return nil
}

// newInvalidConfigError creates a new error for invalid configuration
func newInvalidConfigError(msg string) error {
	return newInvalidActionError("invalid configuration: " + msg)
}
