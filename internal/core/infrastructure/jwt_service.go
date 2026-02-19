package infrastructure

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

// JWTService handles JWT token creation and validation.
type JWTService struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTService creates a new JWT service with the given secret and expiry.
func NewJWTService(secret string, accessExpiry time.Duration) *JWTService {
	return &JWTService{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: 7 * 24 * time.Hour, // 7 days
	}
}

// TokenPair holds access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// GenerateTokenPair creates access + refresh tokens for the given user.
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, tenantID, email string, permissions []string) (*TokenPair, error) {
	now := time.Now()

	// Access token with full claims
	accessClaims := &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiry)),
			ID:        uuid.NewString(),
		},
		UserID:      userID,
		TenantID:    tenantID,
		Email:       email,
		Permissions: permissions,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token with minimal claims
	refreshClaims := &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshExpiry)),
			ID:        uuid.NewString(),
		},
		UserID:   userID,
		TenantID: tenantID,
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessExpiry.Seconds()),
	}, nil
}

// ValidateToken parses and validates a JWT token string.
func (s *JWTService) ValidateToken(tokenStr string) (*auth.Claims, error) {
	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
