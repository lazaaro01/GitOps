package config

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

type Config struct {
	DatabaseURL    string
	RabbitMQURL    string
	APIPort        int
	LogLevel       zerolog.Level
	DeployQueue    string
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://gitops:gitops@localhost:5432/gitops?sslmode=disable"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672"),
		APIPort:     getEnvInt("API_PORT", 8080),
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "debug")),
		DeployQueue: "deploy_jobs",
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
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
