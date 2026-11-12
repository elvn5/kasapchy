// Command gateway runs the public API gateway: CORS, session checks and routing
// of requests to internal microservices.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	gatewayconfig "github.com/your-org/app/gateway/internal/config"
	"github.com/your-org/app/gateway/internal/router"
	"github.com/your-org/app/pkg/config"
	"github.com/your-org/app/pkg/httpserver"
	"github.com/your-org/app/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gateway:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load[gatewayconfig.Config]()
	if err != nil {
		return err
	}
	log := logger.New("gateway", cfg.AppEnv, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	r, err := router.New(cfg, log)
	if err != nil {
		return err
	}
	return httpserver.Run(ctx, cfg.HTTPPort, r, log)
}
