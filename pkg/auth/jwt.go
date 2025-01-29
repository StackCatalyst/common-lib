package auth

import (
	"fmt"
	"time"

	"github.com/StackCatalyst/common-lib/pkg/metrics"
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
	config  Config
	metrics *MetricsReporter
}

// NewTokenManager creates a new token manager
func NewTokenManager(config Config, metricsReporter *metrics.Reporter) (*TokenManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &TokenManager{
		config:  config,
		metrics: NewMetricsReporter(metricsReporter),
	}, nil
}

// generateToken creates a new JWT token
func (tm *TokenManager) generateToken(userID string, roles []string, tokenType TokenType) (string, error) {
	start := time.Now()
	var secret string
	var duration time.Duration

	switch tokenType {
	case AccessToken:
		secret = tm.config.Token.AccessTokenSecret
		duration = tm.config.Token.AccessTokenDuration
	case RefreshToken:
		secret = tm.config.Token.RefreshTokenSecret
		duration = tm.config.Token.RefreshTokenDuration
	default:
		err := fmt.Errorf("invalid token type: %s", tokenType)
		tm.metrics.ObserveTokenGeneration(tokenType, err, time.Since(start))
		return "", err
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:    userID,
		Roles:     roles,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	tm.metrics.ObserveTokenGeneration(tokenType, err, time.Since(start))
	return tokenString, err
}

// validateToken validates a JWT token
func (tm *TokenManager) validateToken(tokenString string, tokenType TokenType) (*Claims, error) {
	start := time.Now()
	var secret string

	switch tokenType {
	case AccessToken:
		secret = tm.config.Token.AccessTokenSecret
	case RefreshToken:
		secret = tm.config.Token.RefreshTokenSecret
	default:
		err := fmt.Errorf("invalid token type: %s", tokenType)
		tm.metrics.ObserveTokenValidation(tokenType, err, time.Since(start))
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		tm.metrics.ObserveTokenValidation(tokenType, err, time.Since(start))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		err = fmt.Errorf("invalid claims type")
		tm.metrics.ObserveTokenValidation(tokenType, err, time.Since(start))
		return nil, err
	}

	if claims.TokenType != tokenType {
		err = fmt.Errorf("token type mismatch: expected %s, got %s", tokenType, claims.TokenType)
		tm.metrics.ObserveTokenValidation(tokenType, err, time.Since(start))
		return nil, err
	}

	tm.metrics.ObserveTokenValidation(tokenType, nil, time.Since(start))
	return claims, nil
}

// GenerateAccessToken generates a new access token
func (tm *TokenManager) GenerateAccessToken(userID string, roles []string) (string, error) {
	return tm.generateToken(userID, roles, AccessToken)
}

// GenerateRefreshToken generates a new refresh token
func (tm *TokenManager) GenerateRefreshToken(userID string, roles []string) (string, error) {
	return tm.generateToken(userID, roles, RefreshToken)
}

// ValidateAccessToken validates an access token
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return tm.validateToken(tokenString, AccessToken)
}

// ValidateRefreshToken validates a refresh token
func (tm *TokenManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return tm.validateToken(tokenString, RefreshToken)
}
