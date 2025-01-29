package auth

import (
	"fmt"
	"time"

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
	config Config
}

// NewTokenManager creates a new token manager
func NewTokenManager(config Config) (*TokenManager, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &TokenManager{config: config}, nil
}

// GenerateAccessToken generates a new access token
func (tm *TokenManager) GenerateAccessToken(userID string, roles []string) (string, error) {
	return tm.generateToken(userID, roles, AccessToken)
}

// GenerateRefreshToken generates a new refresh token
func (tm *TokenManager) GenerateRefreshToken(userID string) (string, error) {
	return tm.generateToken(userID, nil, RefreshToken)
}

// ValidateAccessToken validates an access token
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return tm.validateToken(tokenString, []byte(tm.config.Token.AccessTokenSecret))
}

// ValidateRefreshToken validates a refresh token
func (tm *TokenManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return tm.validateToken(tokenString, []byte(tm.config.Token.RefreshTokenSecret))
}

func (tm *TokenManager) generateToken(userID string, roles []string, tokenType TokenType) (string, error) {
	now := time.Now()
	ttl := tm.config.Token.AccessTokenDuration
	secret := tm.config.Token.AccessTokenSecret
	if tokenType == RefreshToken {
		ttl = tm.config.Token.RefreshTokenDuration
		secret = tm.config.Token.RefreshTokenSecret
	}

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
	return token.SignedString([]byte(secret))
}

func (tm *TokenManager) validateToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		if err.Error() == "token is expired" {
			return nil, newTokenExpiredError()
		}
		return nil, newInvalidTokenError(err.Error())
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, newInvalidTokenError("invalid token claims")
}
