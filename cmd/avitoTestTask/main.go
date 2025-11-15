package main

import (
	"avitoTestTask/internal/config"
	"avitoTestTask/internal/storage"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting avitoTestTask", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")
	log.Error("error messages are enabled")

	teamStorage, err := storage.NewTeamStorage(cfg.StoragePath)
	userStorage, err := storage.NewUserStorage(cfg.StoragePath)
	prStorage, err := storage.NewPullRequestStorage(cfg.StoragePath)
	reviewerStorage, err := storage.NewReviewerStorage(cfg.StoragePath)

	if err != nil {
		log.Error("error creating storage", err)
		os.Exit(1)
	}

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
