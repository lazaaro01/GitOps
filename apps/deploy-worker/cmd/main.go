package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitops-lite/apps/deploy-worker/internal/config"
	"gitops-lite/apps/deploy-worker/internal/consumer"
	"gitops-lite/apps/deploy-worker/internal/executor"
	"gitops-lite/apps/deploy-worker/internal/health"
	"gitops-lite/pkg/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.Load()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Level(cfg.LogLevel)

	log.Info().Msg("starting deploy worker")

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

	tfExecutor := executor.NewTerraformExecutor(cfg.TFWorkDir)
	checker := health.NewChecker()

	c, err := consumer.NewConsumer(
		cfg.RabbitMQURL, cfg.DeployQueue,
		deployRepo, logRepo, jobRepo,
		tfExecutor, checker,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create consumer")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := c.Start(ctx); err != nil {
			log.Fatal().Err(err).Msg("consumer error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down worker...")
	c.Stop()
}
