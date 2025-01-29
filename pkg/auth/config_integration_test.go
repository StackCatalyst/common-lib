package auth

import (
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfigManager(t *testing.T) *config.Manager {
	opts := config.DefaultOptions()
	opts.ConfigName = "test_config"
	opts.ConfigPaths = []string{"testdata"}

	cm, err := config.New(opts)
	require.NoError(t, err)
	return cm
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*config.Manager)
		validate  func(*testing.T, Config, error)
	}{
		{
			name: "default values",
			setupFunc: func(cm *config.Manager) {
				// No setup needed, testing defaults
			},
			validate: func(t *testing.T, cfg Config, err error) {
				require.NoError(t, err)
				assert.Equal(t, 15*time.Minute, cfg.Token.AccessTokenDuration)
				assert.Equal(t, 24*time.Hour, cfg.Token.RefreshTokenDuration)
				assert.Equal(t, "user", cfg.RBAC.DefaultRole)
				assert.Equal(t, "admin", cfg.RBAC.SuperAdminRole)
				assert.True(t, cfg.RateLimit.Enabled)
				assert.Equal(t, 10, cfg.RateLimit.RequestsPerSecond)
				assert.Equal(t, 20, cfg.RateLimit.BurstSize)
			},
		},
		{
			name: "custom values",
			setupFunc: func(cm *config.Manager) {
				cm.Set(configKeyTokenAccessDuration, "30m")
				cm.Set(configKeyTokenRefreshDuration, "48h")
				cm.Set(configKeyTokenAccessSecret, "test-access-secret")
				cm.Set(configKeyTokenRefreshSecret, "test-refresh-secret")
				cm.Set(configKeyRBACDefaultRole, "basic")
				cm.Set(configKeyRBACSuperAdminRole, "superuser")
				cm.Set(configKeyRBACRoleHierarchy, map[string]interface{}{
					"superuser": []string{"basic"},
					"basic":     []string{},
				})
				cm.Set(configKeyRateLimitEnabled, false)
				cm.Set(configKeyRateLimitRPS, 100)
				cm.Set(configKeyRateLimitBurst, 200)
			},
			validate: func(t *testing.T, cfg Config, err error) {
				require.NoError(t, err)
				assert.Equal(t, 30*time.Minute, cfg.Token.AccessTokenDuration)
				assert.Equal(t, 48*time.Hour, cfg.Token.RefreshTokenDuration)
				assert.Equal(t, "test-access-secret", cfg.Token.AccessTokenSecret)
				assert.Equal(t, "test-refresh-secret", cfg.Token.RefreshTokenSecret)
				assert.Equal(t, "basic", cfg.RBAC.DefaultRole)
				assert.Equal(t, "superuser", cfg.RBAC.SuperAdminRole)
				assert.Equal(t, []string{"basic"}, cfg.RBAC.RoleHierarchy["superuser"])
				assert.Empty(t, cfg.RBAC.RoleHierarchy["basic"])
				assert.False(t, cfg.RateLimit.Enabled)
				assert.Equal(t, 100, cfg.RateLimit.RequestsPerSecond)
				assert.Equal(t, 200, cfg.RateLimit.BurstSize)
			},
		},
		{
			name: "invalid role hierarchy",
			setupFunc: func(cm *config.Manager) {
				cm.Set(configKeyRBACRoleHierarchy, map[string]interface{}{
					"admin": 123, // Invalid type
				})
			},
			validate: func(t *testing.T, cfg Config, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid role hierarchy format")
			},
		},
		{
			name: "invalid token duration",
			setupFunc: func(cm *config.Manager) {
				cm.Set(configKeyTokenAccessDuration, "-1m")
			},
			validate: func(t *testing.T, cfg Config, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "access token duration must be positive")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := setupTestConfigManager(t)
			tt.setupFunc(cm)
			cfg, err := LoadConfig(cm)
			tt.validate(t, cfg, err)
		})
	}
}
