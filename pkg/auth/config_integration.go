package auth

import (
	"fmt"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/config"
)

const (
	// Configuration keys
	configKeyTokenAccessDuration  = "auth.token.access_duration"
	configKeyTokenRefreshDuration = "auth.token.refresh_duration"
	configKeyTokenAccessSecret    = "auth.token.access_secret"
	configKeyTokenRefreshSecret   = "auth.token.refresh_secret"
	configKeyRBACDefaultRole      = "auth.rbac.default_role"
	configKeyRBACSuperAdminRole   = "auth.rbac.super_admin_role"
	configKeyRBACRoleHierarchy    = "auth.rbac.role_hierarchy"
	configKeyRateLimitEnabled     = "auth.rate_limit.enabled"
	configKeyRateLimitRPS         = "auth.rate_limit.requests_per_second"
	configKeyRateLimitBurst       = "auth.rate_limit.burst_size"
)

// LoadConfig loads the authentication configuration from the config manager
func LoadConfig(cm *config.Manager) (Config, error) {
	var cfg Config

	// Token settings
	cfg.Token.AccessTokenDuration = cm.GetDuration(configKeyTokenAccessDuration)
	if cfg.Token.AccessTokenDuration == 0 {
		cfg.Token.AccessTokenDuration = 15 * time.Minute // Default value
	}

	cfg.Token.RefreshTokenDuration = cm.GetDuration(configKeyTokenRefreshDuration)
	if cfg.Token.RefreshTokenDuration == 0 {
		cfg.Token.RefreshTokenDuration = 24 * time.Hour // Default value
	}

	cfg.Token.AccessTokenSecret = cm.GetString(configKeyTokenAccessSecret)
	cfg.Token.RefreshTokenSecret = cm.GetString(configKeyTokenRefreshSecret)

	// RBAC settings
	cfg.RBAC.DefaultRole = cm.GetString(configKeyRBACDefaultRole)
	if cfg.RBAC.DefaultRole == "" {
		cfg.RBAC.DefaultRole = "user" // Default value
	}

	cfg.RBAC.SuperAdminRole = cm.GetString(configKeyRBACSuperAdminRole)
	if cfg.RBAC.SuperAdminRole == "" {
		cfg.RBAC.SuperAdminRole = "admin" // Default value
	}

	// Role hierarchy
	hierarchyMap := cm.GetStringMap(configKeyRBACRoleHierarchy)
	if len(hierarchyMap) == 0 {
		// Default hierarchy
		cfg.RBAC.RoleHierarchy = map[string][]string{
			"admin": {"user"},
			"user":  {},
		}
	} else {
		cfg.RBAC.RoleHierarchy = make(map[string][]string)
		for role, parents := range hierarchyMap {
			switch v := parents.(type) {
			case []interface{}:
				parentRoles := make([]string, 0, len(v))
				for _, p := range v {
					if str, ok := p.(string); ok {
						parentRoles = append(parentRoles, str)
					}
				}
				cfg.RBAC.RoleHierarchy[role] = parentRoles
			case []string:
				cfg.RBAC.RoleHierarchy[role] = v
			default:
				return cfg, fmt.Errorf("invalid role hierarchy format for role %s", role)
			}
		}
	}

	// Rate limiting settings
	cfg.RateLimit.Enabled = cm.GetBool(configKeyRateLimitEnabled)
	if cfg.RateLimit.Enabled {
		cfg.RateLimit.Enabled = true // Default value
	}

	cfg.RateLimit.RequestsPerSecond = cm.GetInt(configKeyRateLimitRPS)
	if cfg.RateLimit.RequestsPerSecond == 0 {
		cfg.RateLimit.RequestsPerSecond = 10 // Default value
	}

	cfg.RateLimit.BurstSize = cm.GetInt(configKeyRateLimitBurst)
	if cfg.RateLimit.BurstSize == 0 {
		cfg.RateLimit.BurstSize = 20 // Default value
	}

	// Validate the loaded configuration
	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}
