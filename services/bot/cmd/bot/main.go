// Command bot runs the Telegram bot (long polling).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-telegram/bot"
	"github.com/your-org/app/bot/internal/adapter/telegram"
	"github.com/your-org/app/bot/internal/adapter/usersapi"
	botconfig "github.com/your-org/app/bot/internal/config"
	"github.com/your-org/app/bot/internal/usecase"
	"github.com/your-org/app/pkg/config"
	"github.com/your-org/app/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "bot:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load[botconfig.Config]()
	if err != nil {
		return err
	}
	log := logger.New("bot", cfg.AppEnv, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	onboarding := usecase.NewOnboarding(usersapi.NewClient(cfg.UsersServiceURL))
	handlers := telegram.NewHandlers(onboarding, log)

	b, err := bot.New(cfg.BotToken,
		bot.WithDefaultHandler(handlers.Default),
		bot.WithErrorsHandler(func(err error) {
			log.Error("telegram error", slog.Any("error", err))
		}),
	)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	handlers.Register(b)

	log.Info("bot started")
	b.Start(ctx)
	log.Info("bot stopped")
	return nil
}
