package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "TestTaskSkillsrock/docs"
	"TestTaskSkillsrock/internal/config"
	"TestTaskSkillsrock/internal/handlers"
	sl "TestTaskSkillsrock/internal/lib/slog"
	psql "TestTaskSkillsrock/internal/storage/postgresql"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// @title           Task API
// @version         1.0
// @description     API для управления задачами.
// @host            localhost:9000
// @BasePath        /
func main() {
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)
	logger.Info("starting project", slog.String("env", cfg.Env))

	db, err := psql.New(cfg.StorageDSN)
	if err != nil {
		os.Exit(1)
	}

	defer db.Close(context.Background())

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	})

	h := handlers.NewHandlers(db)

	app.Post("/tasks", h.CreateTask)
	app.Get("/tasks", h.GetTasks)
	app.Put("/tasks/:id", h.UpdateTask)
	app.Delete("/tasks/:id", h.DeleteTask)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	app.Listen(cfg.Address)

	// Graceful shutdown сервера
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(cfg.Address); err != nil {
			logger.Error("failed to start server", sl.Err(err))
		}
	}()

	logger.Info("server started")

	<-done
	logger.Info("stopping server")

	shutdownCh := make(chan error, 1)
	go func() {
		shutdownCh <- app.Shutdown()
	}()

	select {
	case err := <-shutdownCh:
		if err != nil {
			logger.Error("failed to stop server", sl.Err(err))
		} else {
			logger.Info("server stopped gracefully")
		}
	case <-time.After(10 * time.Second):
		logger.Error("server shutdown timed out")
	}
}

// setupLogger устанавливает логгер в зависимости от окружения, пользуемся slog из стандартного пакета golang
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
