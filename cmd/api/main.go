package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go-api-kbt/internal/app"
	"go-api-kbt/internal/config"
	"go-api-kbt/internal/observability"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(logger)
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	provider, err := observability.Setup(ctx, cfg, logger)
	if err != nil {
		logger.Error("failed to setup observability", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown telemetry", slog.String("error", err.Error()))
		}
	}()

	application, err := app.New(ctx, cfg, logger)
	if err != nil {
		logger.Error("failed to create app", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutting down application")
		if err := application.Shutdown(context.Background()); err != nil {
			logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		}
	}()

	if err := application.Start(); err != nil {
		logger.Error("application stopped", slog.String("error", err.Error()))
	}
}
