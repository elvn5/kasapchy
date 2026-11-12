// Command auth runs the authentication microservice: it issues and checks admin sessions.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/app/auth/internal/adapter/httpapi"
	authconfig "github.com/your-org/app/auth/internal/config"
	"github.com/your-org/app/auth/internal/usecase"
	"github.com/your-org/app/pkg/authtoken"
	"github.com/your-org/app/pkg/config"
	"github.com/your-org/app/pkg/httpserver"
	"github.com/your-org/app/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "auth:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load[authconfig.Config]()
	if err != nil {
		return err
	}
	log := logger.New("auth", cfg.AppEnv, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tokens := authtoken.NewManager(cfg.JWTSecret, cfg.JWTExpiresIn)
	authService := usecase.NewAuthService(tokens, cfg.AdminPassword, authtoken.RoleAdmin)

	router := httpserver.NewRouter(log, cfg.IsProduction())
	httpapi.NewHandler(authService, log, cfg.IsProduction()).Register(router)

	return httpserver.Run(ctx, cfg.HTTPPort, router, log)
}
