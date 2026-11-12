// Package router wires the public API routes to downstream services.
package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/your-org/app/gateway/internal/config"
	"github.com/your-org/app/gateway/internal/middleware"
	"github.com/your-org/app/gateway/internal/proxy"
	"github.com/your-org/app/pkg/authtoken"
	"github.com/your-org/app/pkg/httpserver"
)

// New builds the gateway router.
//
// Route table (public path -> service):
//
//	/auth/*   -> auth   (public: login, logout, me)
//	/users/*  -> users  (admin session required)
//
// Internal service routes (e.g. /internal/*) are intentionally not proxied.
func New(cfg config.Config, log *slog.Logger) (*gin.Engine, error) {
	router := httpserver.NewRouter(log, cfg.IsProduction())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	authProxy, err := proxy.New(cfg.AuthServiceURL, log)
	if err != nil {
		return nil, err
	}
	usersProxy, err := proxy.New(cfg.UsersServiceURL, log)
	if err != nil {
		return nil, err
	}

	tokens := authtoken.NewManager(cfg.JWTSecret, 0)
	requireAdmin := middleware.RequireRole(tokens, authtoken.RoleAdmin)

	router.Any("/auth/*path", authProxy)

	users := router.Group("/users", requireAdmin)
	users.Any("", usersProxy)
	users.Any("/*path", usersProxy)

	return router, nil
}
