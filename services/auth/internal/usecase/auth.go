// Package usecase implements the auth service business logic.
package usecase

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/your-org/app/auth/internal/domain"
)

// TokenManager issues and verifies session tokens.
type TokenManager interface {
	Issue(role string) (string, time.Time, error)
	Verify(token string) (string, error)
}

// AuthService authenticates the admin with a single shared password.
// Replace it with a user-based strategy when the project needs one.
type AuthService struct {
	tokens        TokenManager
	adminPassword string
	adminRole     string
}

// NewAuthService creates an AuthService granting adminRole on successful login.
func NewAuthService(tokens TokenManager, adminPassword, adminRole string) *AuthService {
	return &AuthService{tokens: tokens, adminPassword: adminPassword, adminRole: adminRole}
}

// Login checks the password and returns a new session token.
func (s *AuthService) Login(password string) (domain.Token, error) {
	if subtle.ConstantTimeCompare([]byte(password), []byte(s.adminPassword)) != 1 {
		return domain.Token{}, domain.ErrInvalidCredentials
	}

	value, expiresAt, err := s.tokens.Issue(s.adminRole)
	if err != nil {
		return domain.Token{}, fmt.Errorf("issue token: %w", err)
	}
	return domain.Token{Value: value, ExpiresAt: expiresAt}, nil
}

// Authenticate validates a session token.
func (s *AuthService) Authenticate(token string) (domain.Session, error) {
	if token == "" {
		return domain.Session{}, domain.ErrUnauthenticated
	}
	role, err := s.tokens.Verify(token)
	if err != nil {
		return domain.Session{}, domain.ErrUnauthenticated
	}
	return domain.Session{Role: role}, nil
}
