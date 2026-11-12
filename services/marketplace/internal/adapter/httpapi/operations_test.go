package httpapi

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadinessDoesNotLeakDatabaseError(t *testing.T) {
	for _, healthy := range []bool{true, false} {
		router := (Handler{Readiness: func(ctx context.Context) error {
			if _, ok := ctx.Deadline(); !ok {
				t.Fatal("missing database timeout")
			}
			if !healthy {
				return errors.New("secret DSN")
			}
			return nil
		}}).Router()
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", "/readyz", nil))
		expected := 503
		if healthy {
			expected = 200
		}
		if w.Code != expected || strings.Contains(w.Body.String(), "secret DSN") {
			t.Fatalf("unexpected readiness: %d %s", w.Code, w.Body)
		}
	}
}
func TestClientIPOnlyTrustsConfiguredProxy(t *testing.T) {
	for _, tc := range []struct{ remote, want string }{
		{"172.22.0.1:12345", "198.51.100.10"},
		{"203.0.113.2:12345", "203.0.113.2"},
	} {
		router := (Handler{TrustedProxies: []string{"172.22.0.1"}}).Router()
		router.GET("/probe", func(c *gin.Context) { c.String(200, c.ClientIP()) })
		req := httptest.NewRequest("GET", "/probe", nil)
		req.RemoteAddr = tc.remote
		req.Header.Set("X-Forwarded-For", "198.51.100.10")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Body.String() != tc.want {
			t.Fatalf("got %s, want %s", w.Body, tc.want)
		}
	}
}
