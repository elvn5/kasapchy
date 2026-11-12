// Package proxy forwards gateway requests to downstream microservices.
package proxy

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// New returns a handler that reverse-proxies requests to target, keeping the path.
func New(target string, log *slog.Logger) (gin.HandlerFunc, error) {
	upstream, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parse upstream %q: %w", target, err)
	}

	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(upstream)
			r.SetXForwarded()
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.ErrorContext(r.Context(), "upstream request failed",
				slog.String("upstream", target), slog.Any("error", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"message":"Service unavailable"}`))
		},
	}

	return func(c *gin.Context) {
		rp.ServeHTTP(c.Writer, c.Request)
	}, nil
}
