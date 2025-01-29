package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenManager(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: func() Config {
				cfg := DefaultConfig()
				cfg.Token.AccessTokenSecret = "test-access-secret"
				cfg.Token.RefreshTokenSecret = "test-refresh-secret"
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "empty access secret",
			config: func() Config {
				cfg := DefaultConfig()
				cfg.Token.RefreshTokenSecret = "test-refresh-secret"
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "empty refresh secret",
			config: func() Config {
				cfg := DefaultConfig()
				cfg.Token.AccessTokenSecret = "test-access-secret"
				return cfg
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, err := NewTokenManager(tt.config)
			if tt.wantErr {
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
	cfg := DefaultConfig()
	cfg.Token.AccessTokenSecret = "test-access-secret"
	cfg.Token.RefreshTokenSecret = "test-refresh-secret"

	tm, err := NewTokenManager(cfg)
	require.NoError(t, err)

	// Test access token generation
	accessToken, err := tm.GenerateAccessToken("user123", []string{"admin"})
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	// Test refresh token generation
	refreshToken, err := tm.GenerateRefreshToken("user123")
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)
}

func TestTokenValidation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Token.AccessTokenSecret = "test-access-secret"
	cfg.Token.RefreshTokenSecret = "test-refresh-secret"

	tm, err := NewTokenManager(cfg)
	require.NoError(t, err)

	// Generate tokens for testing
	userID := "user123"
	roles := []string{"admin"}
	accessToken, err := tm.GenerateAccessToken(userID, roles)
	require.NoError(t, err)

	refreshToken, err := tm.GenerateRefreshToken(userID)
	require.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		validate  func(string) (*Claims, error)
		wantErr   bool
		checkFunc func(*Claims)
	}{
		{
			name:     "valid access token",
			token:    accessToken,
			validate: tm.ValidateAccessToken,
			wantErr:  false,
			checkFunc: func(claims *Claims) {
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, roles, claims.Roles)
				assert.Equal(t, AccessToken, claims.TokenType)
			},
		},
		{
			name:     "valid refresh token",
			token:    refreshToken,
			validate: tm.ValidateRefreshToken,
			wantErr:  false,
			checkFunc: func(claims *Claims) {
				assert.Equal(t, userID, claims.UserID)
				assert.Empty(t, claims.Roles)
				assert.Equal(t, RefreshToken, claims.TokenType)
			},
		},
		{
			name:     "invalid token format",
			token:    "invalid-token",
			validate: tm.ValidateAccessToken,
			wantErr:  true,
		},
		{
			name:     "empty token",
			token:    "",
			validate: tm.ValidateAccessToken,
			wantErr:  true,
		},
		{
			name:     "wrong signing method",
			token:    "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJleHAiOjE3MDY1MDI0MDAsImlhdCI6MTcwNjQ5ODgwMCwibmJmIjoxNzA2NDk4ODAwLCJ1aWQiOiJ1c2VyMTIzIiwicm9sZXMiOlsiYWRtaW4iXSwidHlwZSI6ImFjY2VzcyJ9.",
			validate: tm.ValidateAccessToken,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := tt.validate(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				tt.checkFunc(claims)
			}
		})
	}
}
