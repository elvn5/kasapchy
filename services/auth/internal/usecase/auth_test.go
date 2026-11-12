package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/your-org/app/auth/internal/domain"
	"github.com/your-org/app/auth/internal/usecase"
	"github.com/your-org/app/pkg/authtoken"
)

func newService() *usecase.AuthService {
	tokens := authtoken.NewManager("test-secret-at-least-16", time.Hour)
	return usecase.NewAuthService(tokens, "s3cret", authtoken.RoleAdmin)
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	_, err := newService().Login("wrong")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginIssuesVerifiableToken(t *testing.T) {
	svc := newService()

	token, err := svc.Login("s3cret")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	session, err := svc.Authenticate(token.Value)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if session.Role != authtoken.RoleAdmin {
		t.Fatalf("expected role %q, got %q", authtoken.RoleAdmin, session.Role)
	}
}

func TestAuthenticateRejectsGarbage(t *testing.T) {
	for _, token := range []string{"", "not-a-jwt"} {
		if _, err := newService().Authenticate(token); !errors.Is(err, domain.ErrUnauthenticated) {
			t.Fatalf("token %q: expected ErrUnauthenticated, got %v", token, err)
		}
	}
}
