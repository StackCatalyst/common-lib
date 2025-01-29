package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenManager(t *testing.T) {
	tests := []struct {
		name        string
		config      TokenManagerConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: TokenManagerConfig{
				AccessSecret:  "test-access-secret",
				RefreshSecret: "test-refresh-secret",
				AccessTTL:     15 * time.Minute,
				RefreshTTL:    7 * 24 * time.Hour,
			},
			expectError: false,
		},
		{
			name: "empty access secret",
			config: TokenManagerConfig{
				AccessSecret:  "",
				RefreshSecret: "test-refresh-secret",
			},
			expectError: true,
		},
		{
			name: "empty refresh secret",
			config: TokenManagerConfig{
				AccessSecret:  "test-access-secret",
				RefreshSecret: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, err := NewTokenManager(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tm)
			}
		})
	}
}

func TestTokenGeneration(t *testing.T) {
	tm, err := NewTokenManager(TokenManagerConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
	})
	require.NoError(t, err)

	userID := "test-user"
	roles := []string{"admin", "user"}

	// Test access token
	accessToken, err := tm.GenerateAccessToken(userID, roles)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	claims, err := tm.ValidateAccessToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, AccessToken, claims.TokenType)

	// Test refresh token
	refreshToken, err := tm.GenerateRefreshToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	refreshClaims, err := tm.ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Empty(t, refreshClaims.Roles)
	assert.Equal(t, RefreshToken, refreshClaims.TokenType)
}

func TestTokenValidation(t *testing.T) {
	tm, err := NewTokenManager(TokenManagerConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "invalid token format",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "wrong signing method",
			token:       "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tm.ValidateAccessToken(tt.token)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
