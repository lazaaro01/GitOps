package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

type Config struct {
	DatabaseURL   string
	RabbitMQURL   string
	LogLevel      zerolog.Level
	DeployQueue   string
	TFWorkDir     string
	ContainerPort int
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://gitops:gitops@localhost:5432/gitops?sslmode=disable"),
		RabbitMQURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672"),
		LogLevel:      parseLogLevel(getEnv("LOG_LEVEL", "debug")),
		DeployQueue:   "deploy_jobs",
		TFWorkDir:     getEnv("TF_WORK_DIR", defaultTFWorkDir()),
		ContainerPort: getEnvInt("CONTAINER_PORT", 8080),
	}
	return cfg
}

func defaultTFWorkDir() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "..", "terraform", "app")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return fallback
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
