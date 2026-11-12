package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/your-org/app/marketplace/internal/domain"
	"math/big"
	"strings"
	"time"
)

type Repository interface {
	Allow(context.Context, string, int, time.Duration) (bool, error)
	CreateChallenge(context.Context, string, string, string, string) error
	ConsumeChallenge(context.Context, string, string) (string, string, error)
	RemoveChallenge(context.Context, string) error
	Login(context.Context, string, string, string, string) (domain.User, error)
	Authenticate(context.Context, string) (domain.User, error)
	Logout(context.Context, string) error
	SaveProfile(context.Context, string, string, string) (domain.User, error)
	DeleteAccount(context.Context, string, string) error
	Products(context.Context) ([]domain.Product, error)
	Publish(context.Context, domain.User, domain.NewProduct) (domain.Product, error)
	Orders(context.Context, domain.User) ([]domain.Order, error)
	Place(context.Context, domain.User, domain.NewOrder) (domain.Order, error)
	Advance(context.Context, domain.User, int64, string) (domain.Order, error)
	CreateOIDC(context.Context, string, string, string, string, string) error
	CompleteApple(context.Context, string, string, string) error
	ConsumeApple(context.Context, string, string) (string, string, error)
	ConsumeOIDC(context.Context, string, string, string) (string, error)
}
type Mailer interface {
	SendCode(context.Context, string, string) error
}
type Identity struct{ Email, Nonce string }
type IdentityVerifier interface {
	Verify(context.Context, string, string) (Identity, error)
}
type Service struct {
	Repo     Repository
	Mail     Mailer
	Identity IdentityVerifier
	Secret   []byte
}

func randomToken() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func Hash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func (s *Service) codeHash(id, code string) string {
	h := hmac.New(sha256.New, s.Secret)
	h.Write([]byte(id + ":" + code))
	return hex.EncodeToString(h.Sum(nil))
}
func (s *Service) RequestCode(ctx context.Context, email, role, ip string) (string, error) {
	email, e := domain.NormalizeEmail(email)
	if e != nil || !domain.ValidRole(role) {
		return "", domain.ErrInvalid
	}
	for _, rule := range []struct {
		k string
		n int
		d time.Duration
	}{{"ip:" + ip, 20, time.Hour}, {"email:" + email, 5, time.Hour}} {
		ok, e := s.Repo.Allow(ctx, rule.k, rule.n, rule.d)
		if e != nil {
			return "", e
		}
		if !ok {
			return "", domain.ErrRateLimit
		}
	}
	id := randomToken()
	n, e := rand.Int(rand.Reader, big.NewInt(1000000))
	if e != nil {
		return "", e
	}
	code := strings.Repeat("0", 6-len(n.String())) + n.String()
	if e = s.Repo.CreateChallenge(ctx, id, email, role, s.codeHash(id, code)); e != nil {
		return "", e
	}
	if e = s.Mail.SendCode(ctx, email, code); e != nil {
		_ = s.Repo.RemoveChallenge(ctx, id)
		return "", e
	}
	return id, nil
}
func (s *Service) VerifyCode(ctx context.Context, id, code string) (domain.Session, error) {
	if len(id) != 64 || len(code) != 6 {
		return domain.Session{}, domain.ErrInvalid
	}
	email, role, e := s.Repo.ConsumeChallenge(ctx, id, s.codeHash(id, code))
	if e != nil {
		return domain.Session{}, e
	}
	return s.login(ctx, email, role, "Email")
}
func (s *Service) login(ctx context.Context, email, role, provider string) (domain.Session, error) {
	token := randomToken()
	u, e := s.Repo.Login(ctx, email, role, provider, Hash(token))
	return domain.Session{Token: token, User: u}, e
}
func (s *Service) StartOIDC(ctx context.Context, provider, role, proofHash string) (string, string, error) {
	if (provider != "Google" && provider != "Apple") || !domain.ValidRole(role) || (provider == "Apple" && len(proofHash) != 64) {
		return "", "", domain.ErrInvalid
	}
	id, nonce := randomToken(), randomToken()
	e := s.Repo.CreateOIDC(ctx, id, provider, nonce, role, proofHash)
	return id, nonce, e
}
func (s *Service) FinishOIDC(ctx context.Context, id, provider, token string) (domain.Session, error) {
	if s.Identity == nil {
		return domain.Session{}, domain.ErrForbidden
	}
	identity, e := s.Identity.Verify(ctx, provider, token)
	if e != nil {
		return domain.Session{}, domain.ErrUnauthorized
	}
	role, e := s.Repo.ConsumeOIDC(ctx, id, provider, identity.Nonce)
	if e != nil {
		return domain.Session{}, e
	}
	email, e := domain.NormalizeEmail(identity.Email)
	if e != nil {
		return domain.Session{}, domain.ErrUnauthorized
	}
	return s.login(ctx, email, role, provider)
}

func (s *Service) AppleCallback(ctx context.Context, id, token string) error {
	identity, e := s.Identity.Verify(ctx, "Apple", token)
	if e != nil {
		return domain.ErrUnauthorized
	}
	return s.Repo.CompleteApple(ctx, id, identity.Nonce, identity.Email)
}
func (s *Service) ApplePoll(ctx context.Context, id, proof string) (domain.Session, error) {
	if len(proof) < 43 || len(proof) > 128 {
		return domain.Session{}, domain.ErrInvalid
	}
	email, role, e := s.Repo.ConsumeApple(ctx, id, Hash(proof))
	if e != nil {
		return domain.Session{}, e
	}
	return s.login(ctx, email, role, "Apple")
}
