package auth

import (
	"fmt"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of token
type TokenType string

const (
	// AccessToken is used for API access
	AccessToken TokenType = "access"
	// RefreshToken is used to obtain new access tokens
	RefreshToken TokenType = "refresh"
)

// Claims represents the claims in a JWT token
type Claims struct {
	jwt.RegisteredClaims
	UserID    string    `json:"uid"`
	Roles     []string  `json:"roles"`
	TokenType TokenType `json:"type"`
}

// TokenManager handles JWT token operations
type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// TokenManagerConfig configures the token manager
type TokenManagerConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// NewTokenManager creates a new token manager
func NewTokenManager(config TokenManagerConfig) (*TokenManager, error) {
	if config.AccessSecret == "" || config.RefreshSecret == "" {
		return nil, errors.New(errors.ErrValidation, "secrets cannot be empty")
	}

	if config.AccessTTL == 0 {
		config.AccessTTL = 15 * time.Minute
	}

	if config.RefreshTTL == 0 {
		config.RefreshTTL = 7 * 24 * time.Hour
	}

	return &TokenManager{
		accessSecret:  []byte(config.AccessSecret),
		refreshSecret: []byte(config.RefreshSecret),
		accessTTL:     config.AccessTTL,
		refreshTTL:    config.RefreshTTL,
	}, nil
}

// GenerateAccessToken generates a new access token
func (tm *TokenManager) GenerateAccessToken(userID string, roles []string) (string, error) {
	return tm.generateToken(userID, roles, AccessToken, tm.accessSecret, tm.accessTTL)
}

// GenerateRefreshToken generates a new refresh token
func (tm *TokenManager) GenerateRefreshToken(userID string) (string, error) {
	return tm.generateToken(userID, nil, RefreshToken, tm.refreshSecret, tm.refreshTTL)
}

// ValidateAccessToken validates an access token
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return tm.validateToken(tokenString, tm.accessSecret)
}

// ValidateRefreshToken validates a refresh token
func (tm *TokenManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return tm.validateToken(tokenString, tm.refreshSecret)
}

func (tm *TokenManager) generateToken(userID string, roles []string, tokenType TokenType, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:    userID,
		Roles:     roles,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (tm *TokenManager) validateToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New(errors.ErrUnauthorized, "invalid token claims")
}
