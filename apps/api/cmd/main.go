package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitops-lite/apps/api/internal/config"
	"gitops-lite/apps/api/internal/handler"
	"gitops-lite/apps/api/internal/queue"
	"gitops-lite/pkg/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Level(cfg.LogLevel)

	log.Info().Int("port", cfg.APIPort).Msg("starting API server")

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}
	log.Info().Msg("connected to database")

	deployRepo := repository.NewDeploymentRepo(pool)
	logRepo := repository.NewLogRepo(pool)
	jobRepo := repository.NewJobRepo(pool)

	producer, err := queue.NewProducer(cfg.RabbitMQURL, cfg.DeployQueue)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create queue producer")
	}
	defer producer.Close()
	log.Info().Msg("connected to RabbitMQ")

	deployHandler := handler.NewDeployHandler(deployRepo, jobRepo, producer)
	deploymentHandler := handler.NewDeploymentHandler(deployRepo, logRepo)

	app := fiber.New(fiber.Config{
		AppName: "GitOps Lite API",
	})

	handler.SetupRouter(app, deployHandler, deploymentHandler)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.APIPort)
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	log.Info().Msg("shutting down server...")
	app.Shutdown()
}
