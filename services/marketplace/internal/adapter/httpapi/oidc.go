package httpapi

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/your-org/app/marketplace/internal/usecase"
)

type OIDC struct {
	Verifiers map[string]*oidc.IDTokenVerifier
}

func NewOIDC(ctx context.Context, googleID, appleID string) (*OIDC, error) {
	result := &OIDC{Verifiers: map[string]*oidc.IDTokenVerifier{}}
	for _, p := range []struct{ name, issuer, id string }{{"Google", "https://accounts.google.com", googleID}, {"Apple", "https://appleid.apple.com", appleID}} {
		if p.id == "" {
			continue
		}
		provider, e := oidc.NewProvider(ctx, p.issuer)
		if e != nil {
			return nil, e
		}
		result.Verifiers[p.name] = provider.Verifier(&oidc.Config{ClientID: p.id})
	}
	return result, nil
}
func (v *OIDC) Verify(ctx context.Context, provider, raw string) (usecase.Identity, error) {
	verifier := v.Verifiers[provider]
	if verifier == nil {
		return usecase.Identity{}, fmt.Errorf("provider not configured")
	}
	token, e := verifier.Verify(ctx, raw)
	if e != nil {
		return usecase.Identity{}, e
	}
	var claims struct {
		Email    string `json:"email"`
		Verified any    `json:"email_verified"`
		Nonce    string `json:"nonce"`
		AZP      string `json:"azp"`
	}
	if e = token.Claims(&claims); e != nil {
		return usecase.Identity{}, e
	}
	if (claims.Verified != true && claims.Verified != "true") || claims.Email == "" || claims.Nonce == "" {
		return usecase.Identity{}, fmt.Errorf("unverified identity")
	}
	// The nonce is consumed atomically by the repository, binding this token to a one-time login challenge.
	return usecase.Identity{Email: claims.Email, Nonce: claims.Nonce}, nil
}
