package main

import (
	"avitoTestTask/internal/config"
	controllers "avitoTestTask/internal/http-server/controllers"
	"avitoTestTask/internal/service"
	dao "avitoTestTask/internal/storage/Postgres"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	envLocal = "local"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting avitoTestTask", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")
	log.Error("error messages are enabled")

	// делаем слой для работы с БД
	Storage, err := dao.NewPostgresStorage(cfg.StoragePath, log)

	if err != nil {
		log.Error("error creating storage", err)
		os.Exit(1)
	}

	// делаем сервисный слой
	teamService := service.CreateTeamService(Storage, log)
	userService := service.CreateUserService(Storage, log)
	pullRequestService := service.CreatePullRequestService(Storage, log)

	// делаем хэндлеры
	router := gin.Default()
	teamHandler := controllers.CreateTeamController(&teamService, router, log)
	userHandler := controllers.CreateUserController(&userService, router, log)
	pullRequestHandler := controllers.CreatePullRequestController(&pullRequestService, router, log)
	healthHandler := controllers.CreateHealthController(router, log)

	// Включаем хэндлеры
	teamHandler.EnableController()
	userHandler.EnableController()
	pullRequestHandler.EnableController()
	healthHandler.EnableController()

	// Запускаем роутер
	router.Run()

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
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
