// Command users runs the users microservice. It owns the "users" database.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/app/pkg/config"
	"github.com/your-org/app/pkg/httpserver"
	"github.com/your-org/app/pkg/logger"
	"github.com/your-org/app/pkg/postgres"
	"github.com/your-org/app/users/internal/adapter/httpapi"
	userrepo "github.com/your-org/app/users/internal/adapter/repository/postgres"
	usersconfig "github.com/your-org/app/users/internal/config"
	"github.com/your-org/app/users/internal/usecase"
	"github.com/your-org/app/users/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "users:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load[usersconfig.Config]()
	if err != nil {
		return err
	}
	log := logger.New("users", cfg.AppEnv, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.Postgres.DSN(cfg.DBName))
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool, migrations.FS); err != nil {
		return err
	}
	log.Info("database migrations applied")

	userService := usecase.NewUserService(userrepo.NewUserRepository(pool))

	router := httpserver.NewRouter(log, cfg.IsProduction())
	httpapi.NewHandler(userService, log).Register(router)

	return httpserver.Run(ctx, cfg.HTTPPort, router, log)
}
