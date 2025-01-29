package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test token defaults
	assert.Equal(t, 15*time.Minute, cfg.Token.AccessTokenDuration)
	assert.Equal(t, 24*time.Hour, cfg.Token.RefreshTokenDuration)
	assert.Empty(t, cfg.Token.AccessTokenSecret)
	assert.Empty(t, cfg.Token.RefreshTokenSecret)

	// Test RBAC defaults
	assert.Equal(t, "user", cfg.RBAC.DefaultRole)
	assert.Equal(t, "admin", cfg.RBAC.SuperAdminRole)
	assert.Contains(t, cfg.RBAC.RoleHierarchy, "admin")
	assert.Contains(t, cfg.RBAC.RoleHierarchy, "user")
	assert.Equal(t, []string{"user"}, cfg.RBAC.RoleHierarchy["admin"])
	assert.Empty(t, cfg.RBAC.RoleHierarchy["user"])

	// Test rate limiting defaults
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 10, cfg.RateLimit.RequestsPerSecond)
	assert.Equal(t, 20, cfg.RateLimit.BurstSize)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		modifyFn  func(*Config)
		wantError bool
	}{
		{
			name: "valid config",
			modifyFn: func(c *Config) {
				c.Token.AccessTokenSecret = "secret1"
				c.Token.RefreshTokenSecret = "secret2"
			},
			wantError: false,
		},
		{
			name: "invalid access token duration",
			modifyFn: func(c *Config) {
				c.Token.AccessTokenDuration = -1
				c.Token.AccessTokenSecret = "secret1"
				c.Token.RefreshTokenSecret = "secret2"
			},
			wantError: true,
		},
		{
			name: "invalid refresh token duration",
			modifyFn: func(c *Config) {
				c.Token.RefreshTokenDuration = -1
				c.Token.AccessTokenSecret = "secret1"
				c.Token.RefreshTokenSecret = "secret2"
			},
			wantError: true,
		},
		{
			name: "missing access token secret",
			modifyFn: func(c *Config) {
				c.Token.RefreshTokenSecret = "secret2"
			},
			wantError: true,
		},
		{
			name: "missing refresh token secret",
			modifyFn: func(c *Config) {
				c.Token.AccessTokenSecret = "secret1"
			},
			wantError: true,
		},
		{
			name: "missing default role",
			modifyFn: func(c *Config) {
				c.Token.AccessTokenSecret = "secret1"
				c.Token.RefreshTokenSecret = "secret2"
				c.RBAC.DefaultRole = ""
			},
			wantError: true,
		},
		{
			name: "missing super admin role",
			modifyFn: func(c *Config) {
				c.Token.AccessTokenSecret = "secret1"
				c.Token.RefreshTokenSecret = "secret2"
				c.RBAC.SuperAdminRole = ""
			},
			wantError: true,
		},
		{
			name: "invalid rate limit requests per second",
			modifyFn: func(c *Config) {
				c.Token.AccessTokenSecret = "secret1"
				c.Token.RefreshTokenSecret = "secret2"
				c.RateLimit.RequestsPerSecond = -1
			},
			wantError: true,
		},
		{
			name: "invalid rate limit burst size",
			modifyFn: func(c *Config) {
				c.Token.AccessTokenSecret = "secret1"
				c.Token.RefreshTokenSecret = "secret2"
				c.RateLimit.BurstSize = -1
			},
			wantError: true,
		},
		{
			name: "rate limit disabled ignores invalid values",
			modifyFn: func(c *Config) {
				c.Token.AccessTokenSecret = "secret1"
				c.Token.RefreshTokenSecret = "secret2"
				c.RateLimit.Enabled = false
				c.RateLimit.RequestsPerSecond = -1
				c.RateLimit.BurstSize = -1
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modifyFn(&cfg)
			err := cfg.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
