package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"workend/api/internal/config"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/db"
	"workend/api/internal/oauth"
	"workend/api/internal/secret"
	"workend/api/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "err", err)
		os.Exit(1)
	}

	logger.Info("running migrations")
	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		logger.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	dc := wdagger.NewClient()
	defer func() {
		if err := dc.Close(); err != nil {
			logger.Warn("dagger close", "err", err)
		}
	}()

	var gh *oauth.GitHub
	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		if cfg.TokenKey == "" {
			logger.Error("WORKEND_TOKEN_KEY required when GitHub OAuth is configured")
			os.Exit(1)
		}
		box, err := secret.NewBox(cfg.TokenKey)
		if err != nil {
			logger.Error("token key invalid", "err", err)
			os.Exit(1)
		}
		gh = &oauth.GitHub{
			Pool:         pool,
			Box:          box,
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.GitHubRedirectURL,
			Scopes:       []string{"repo", "read:user"},
			Secure:       cfg.CookieSecure,
		}
		logger.Info("github oauth enabled")
	}

	srv := server.New(cfg, pool, dc, gh, logger)

	httpServer := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", cfg.ListenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
}
