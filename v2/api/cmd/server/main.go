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
	"workend/api/internal/detect"
	"workend/api/internal/oauth"
	"workend/api/internal/oidc"
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

	// The Dagger session must outlive any individual HTTP request, so we
	// bind it to the main process context (cancelled on SIGTERM).
	dc := wdagger.NewClient(ctx)
	defer func() {
		if err := dc.Close(); err != nil {
			logger.Warn("dagger close", "err", err)
		}
	}()

	customDetectors, err := detect.LoadCustom(dc)
	if err != nil {
		logger.Error("custom detectors load failed", "err", err)
		os.Exit(1)
	}
	// Always-on Dagger detector (needs the client; added here rather than
	// in the static builtin list).
	extras := []detect.Detector{detect.Dagger{Dagger: dc}}
	extras = append(extras, customDetectors...)
	detect.InitDetectors(extras)
	if len(customDetectors) > 0 {
		names := make([]string, 0, len(customDetectors))
		for _, d := range customDetectors {
			names = append(names, d.Name())
		}
		logger.Info("custom detectors loaded", "detectors", names)
	}

	providers, err := oauth.LoadConfig(cfg.WebPublicURL)
	if err != nil {
		logger.Error("oauth config load failed", "err", err)
		os.Exit(1)
	}
	// Build the secret box up-front whenever WORKEND_TOKEN_KEY is set; both
	// OAuth tokens and SSH keys reuse it for at-rest encryption.
	var box *secret.Box
	if cfg.TokenKey != "" {
		box, err = secret.NewBox(cfg.TokenKey)
		if err != nil {
			logger.Error("token key invalid", "err", err)
			os.Exit(1)
		}
	}

	var oauthReg *oauth.Registry
	if len(providers) > 0 {
		if box == nil {
			logger.Error("WORKEND_TOKEN_KEY required when any OAuth provider is configured")
			os.Exit(1)
		}
		oauthReg = oauth.NewRegistry(pool, box, providers)
		ids := make([]string, 0, len(providers))
		for _, p := range providers {
			ids = append(ids, p.ID())
		}
		logger.Info("oauth providers configured", "providers", ids)
	}

	var oidcDoc *oidc.DiscoveryDoc
	if cfg.OIDCConfigured() {
		var err error
		oidcDoc, err = oidc.Discover(ctx, cfg.OIDCIssuer)
		if err != nil {
			logger.Error("oidc discovery failed", "issuer", cfg.OIDCIssuer, "err", err)
			os.Exit(1)
		}
		logger.Info("oidc configured", "issuer", cfg.OIDCIssuer, "provider", cfg.OIDCProviderName)
	}

	srv := server.New(cfg, pool, dc, oauthReg, box, logger, oidcDoc)

	stopScheduler := srv.Schedules().StartTicker(ctx, srv.Runs().EnqueueForUser)
	defer stopScheduler()

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
