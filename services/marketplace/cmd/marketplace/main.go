package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/app/marketplace/internal/adapter/httpapi"
	"github.com/your-org/app/marketplace/internal/adapter/mailer"
	"github.com/your-org/app/marketplace/internal/adapter/postgres"
	"github.com/your-org/app/marketplace/internal/usecase"
	"github.com/your-org/app/marketplace/migrations"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	secret := os.Getenv("MARKET_OTP_SECRET")
	if len(secret) < 32 {
		return errors.New("MARKET_OTP_SECRET must contain at least 32 characters")
	}
	address := os.Getenv("MARKET_SMTP_ADDRESS")
	host, _, e := net.SplitHostPort(address)
	if e != nil {
		return errors.New("MARKET_SMTP_ADDRESS must be host:port")
	}
	local := os.Getenv("APP_ENV") == "development"
	if !local {
		gin.SetMode(gin.ReleaseMode)
	}
	from := os.Getenv("MARKET_MAIL_FROM")
	if from == "" || strings.ContainsAny(from, "\r\n") {
		return errors.New("MARKET_MAIL_FROM required")
	}
	pool, e := pgxpool.New(ctx, os.Getenv("MARKET_DATABASE_URL"))
	if e != nil {
		return e
	}
	defer pool.Close()
	if e = pool.Ping(ctx); e != nil {
		return e
	}
	if _, e = pool.Exec(ctx, migrations.Initial); e != nil {
		return e
	}
	identity, e := httpapi.NewOIDC(ctx, os.Getenv("MARKET_GOOGLE_CLIENT_ID"), os.Getenv("MARKET_APPLE_CLIENT_ID"))
	if e != nil {
		return e
	}
	repo := &postgres.Repository{DB: pool}
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if e := repo.Cleanup(ctx); e != nil {
					log.Error("auth cleanup failed", "error", e)
				}
			}
		}
	}()
	service := &usecase.Service{Repo: repo, Secret: []byte(secret), Identity: identity, Mail: mailer.SMTP{Address: address, Host: host, Username: os.Getenv("MARKET_SMTP_USERNAME"), Password: os.Getenv("MARKET_SMTP_PASSWORD"), From: from, Local: local}}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: ":" + port, Handler: (httpapi.Handler{Service: service, Log: log, Operator: os.Getenv("MARKET_OPERATOR_NAME"), SupportEmail: os.Getenv("MARKET_SUPPORT_EMAIL")}).Router(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	errc := make(chan error, 1)
	go func() { log.Info("marketplace listening", "port", port); errc <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	case e := <-errc:
		if errors.Is(e, http.ErrServerClosed) {
			return nil
		}
		return e
	}
}
func main() {
	if e := run(); e != nil {
		slog.Error("marketplace stopped", "error", e)
		os.Exit(1)
	}
}
