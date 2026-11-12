// Package middleware contains gateway-level HTTP middleware.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/app/pkg/authtoken"
	"github.com/your-org/app/pkg/httpserver"
)

// TokenVerifier validates a session token and returns its role.
type TokenVerifier interface {
	Verify(token string) (string, error)
}

// RequireRole rejects requests without a valid session cookie for role.
// Downstream services behind this middleware can trust that the caller is authorised.
func RequireRole(tokens TokenVerifier, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(authtoken.CookieName)
		if err != nil || token == "" {
			httpserver.Error(c, http.StatusUnauthorized, "Not authenticated")
			return
		}

		got, err := tokens.Verify(token)
		if err != nil {
			httpserver.Error(c, http.StatusUnauthorized, "Invalid or expired session")
			return
		}
		if got != role {
			httpserver.Error(c, http.StatusForbidden, "Forbidden")
			return
		}
		c.Next()
	}
}
