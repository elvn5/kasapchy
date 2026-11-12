// Package authtoken issues and verifies the admin session JWT shared by the auth
// service (issuer) and the gateway (verifier).
package authtoken

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CookieName is the httpOnly cookie that carries the session token.
const CookieName = "admin_token"

// RoleAdmin is the role granted to admin panel sessions.
const RoleAdmin = "admin"

// ErrInvalidToken is returned for malformed, tampered or expired tokens.
var ErrInvalidToken = errors.New("invalid or expired token")

type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Manager signs and verifies HS256 tokens with a shared secret.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager creates a Manager. ttl is only used when issuing tokens.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Issue creates a signed token for role and returns it with its expiry time.
func (m *Manager) Issue(role string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

// Verify validates the token and returns the role it was issued for.
func (m *Manager) Verify(token string) (string, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", ErrInvalidToken
	}
	return c.Role, nil
}
