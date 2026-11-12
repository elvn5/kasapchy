package router_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/app/gateway/internal/config"
	"github.com/your-org/app/gateway/internal/router"
	"github.com/your-org/app/pkg/authtoken"
)

const secret = "test-secret-at-least-16"

// newGateway starts the gateway in front of a fake upstream that echoes the request path.
// A real server is used because the reverse proxy needs a full http.ResponseWriter.
func newGateway(t *testing.T) *httptest.Server {
	t.Helper()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "upstream:"+r.URL.Path)
	}))
	t.Cleanup(upstream.Close)

	cfg := config.Config{
		CORSOrigins:     []string{"http://localhost:3000"},
		JWTSecret:       secret,
		AuthServiceURL:  upstream.URL,
		UsersServiceURL: upstream.URL,
	}
	r, err := router.New(cfg, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("router.New: %v", err)
	}

	gateway := httptest.NewServer(r)
	t.Cleanup(gateway.Close)
	return gateway
}

func do(t *testing.T, req *http.Request) (int, string) {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestUsersRequireSession(t *testing.T) {
	gw := newGateway(t)
	req, _ := http.NewRequest(http.MethodGet, gw.URL+"/users", nil)

	if status, _ := do(t, req); status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", status)
	}
}

func TestUsersProxiedWithValidSession(t *testing.T) {
	gw := newGateway(t)
	token, _, err := authtoken.NewManager(secret, time.Hour).Issue(authtoken.RoleAdmin)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	for _, path := range []string{"/users", "/users/123"} {
		req, _ := http.NewRequest(http.MethodGet, gw.URL+path, nil)
		req.AddCookie(&http.Cookie{Name: authtoken.CookieName, Value: token})

		if status, body := do(t, req); status != http.StatusOK || body != "upstream:"+path {
			t.Fatalf("%s: unexpected response %d %q", path, status, body)
		}
	}
}

func TestAuthRoutesArePublic(t *testing.T) {
	gw := newGateway(t)
	req, _ := http.NewRequest(http.MethodPost, gw.URL+"/auth/login", nil)

	if status, body := do(t, req); status != http.StatusOK || body != "upstream:/auth/login" {
		t.Fatalf("unexpected response %d %q", status, body)
	}
}

func TestInternalRoutesAreNotExposed(t *testing.T) {
	gw := newGateway(t)
	req, _ := http.NewRequest(http.MethodPost, gw.URL+"/internal/users/telegram", nil)

	if status, _ := do(t, req); status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
}

func TestCORSPreflight(t *testing.T) {
	gw := newGateway(t)
	req, _ := http.NewRequest(http.MethodOptions, gw.URL+"/users", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("preflight: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials to be allowed, got %q", got)
	}
}
