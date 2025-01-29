package auth

import (
	"testing"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMetricsReporter() *metrics.Reporter {
	registry := prometheus.NewRegistry()
	return metrics.New(metrics.Options{
		Namespace: "test",
		Subsystem: "auth",
		Registry:  registry,
	})
}

func setupTestTokenManager(t *testing.T) *TokenManager {
	config := Config{}
	config.Token.AccessTokenSecret = "test-access-secret"
	config.Token.RefreshTokenSecret = "test-refresh-secret"
	config.Token.AccessTokenDuration = 15 * time.Minute
	config.Token.RefreshTokenDuration = 24 * time.Hour
	config.RBAC.DefaultRole = "user"
	config.RBAC.SuperAdminRole = "admin"

	tm, err := NewTokenManager(config, newTestMetricsReporter())
	require.NoError(t, err)
	require.NotNil(t, tm)
	return tm
}

func TestNewTokenManager(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := Config{}
		config.Token.AccessTokenSecret = "test-access-secret"
		config.Token.RefreshTokenSecret = "test-refresh-secret"
		config.Token.AccessTokenDuration = 15 * time.Minute
		config.Token.RefreshTokenDuration = 24 * time.Hour
		config.RBAC.DefaultRole = "user"
		config.RBAC.SuperAdminRole = "admin"

		tm, err := NewTokenManager(config, newTestMetricsReporter())
		require.NoError(t, err)
		require.NotNil(t, tm)
	})

	t.Run("empty access secret", func(t *testing.T) {
		config := Config{}
		config.Token.RefreshTokenSecret = "test-refresh-secret"
		config.RBAC.DefaultRole = "user"
		config.RBAC.SuperAdminRole = "admin"

		tm, err := NewTokenManager(config, newTestMetricsReporter())
		assert.Error(t, err)
		assert.Nil(t, tm)
	})

	t.Run("empty refresh secret", func(t *testing.T) {
		config := Config{}
		config.Token.AccessTokenSecret = "test-access-secret"
		config.RBAC.DefaultRole = "user"
		config.RBAC.SuperAdminRole = "admin"

		tm, err := NewTokenManager(config, newTestMetricsReporter())
		assert.Error(t, err)
		assert.Nil(t, tm)
	})
}

func TestTokenGeneration(t *testing.T) {
	tm := setupTestTokenManager(t)

	userID := "test-user"
	roles := []string{"admin", "user"}

	// Test access token generation
	accessToken, err := tm.GenerateAccessToken(userID, roles)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)

	// Test refresh token generation
	refreshToken, err := tm.GenerateRefreshToken(userID, roles)
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)
}

func TestTokenValidation(t *testing.T) {
	tm := setupTestTokenManager(t)

	userID := "test-user"
	roles := []string{"admin", "user"}

	t.Run("valid access token", func(t *testing.T) {
		token, err := tm.GenerateAccessToken(userID, roles)
		require.NoError(t, err)

		claims, err := tm.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, roles, claims.Roles)
		assert.Equal(t, AccessToken, claims.TokenType)
	})

	t.Run("valid refresh token", func(t *testing.T) {
		token, err := tm.GenerateRefreshToken(userID, roles)
		require.NoError(t, err)

		claims, err := tm.ValidateRefreshToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, roles, claims.Roles)
		assert.Equal(t, RefreshToken, claims.TokenType)
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := tm.ValidateAccessToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := tm.ValidateAccessToken("")
		assert.Error(t, err)
	})

	t.Run("wrong signing method", func(t *testing.T) {
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.EkN-DOsnsuRjRO6BxXemmJDm3HbxrbRzXglbN2S4sOkopdU4IsDxTI8jO19W_A4K8ZPJijNLis4EZsHeY559a4DFOd50_OqgHGuERTqYZyuhtF39yxJPAjUESwxk2J5k_4zM3O-vtd1Ghyo4IbqKKSy6J9mTniYJPenn5-HIirE"
		_, err := tm.ValidateAccessToken(token)
		assert.Error(t, err)
	})
}
