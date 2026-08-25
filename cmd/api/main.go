package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/wolf-platform/wolf-platform/internal/config"
	"github.com/wolf-platform/wolf-platform/internal/platform/database"
	"github.com/wolf-platform/wolf-platform/internal/platform/logger"
	"github.com/wolf-platform/wolf-platform/internal/server"
)

func main() {
	cfg := config.Load()

	log := logger.New(cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	srv := server.New(cfg, log, db)

	if err := srv.Run(ctx); err != nil {
		log.Error("server exited with error", "error", err)
		os.Exit(1)
	}

	slog.Info("server shutdown complete")
}
