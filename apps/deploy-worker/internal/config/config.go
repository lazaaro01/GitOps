package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

type Config struct {
	DatabaseURL   string
	RabbitMQURL   string
	LogLevel      zerolog.Level
	DeployQueue   string
	TFWorkDir     string
	ContainerPort int
	APIURL        string
}

func Load() *Config {
	loadEnvFile()
	cfg := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://gitops:gitops@localhost:5432/gitops?sslmode=disable"),
		RabbitMQURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672"),
		LogLevel:      parseLogLevel(getEnv("LOG_LEVEL", "debug")),
		DeployQueue:   "deploy_jobs",
		TFWorkDir:     getEnv("TF_WORK_DIR", defaultTFWorkDir()),
		ContainerPort: getEnvInt("CONTAINER_PORT", 8080),
		APIURL:        getEnv("API_URL", "http://localhost:8080"),
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

func loadEnvFile() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		path := filepath.Join(dir, ".env")
		if data, err := os.ReadFile(path); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					if os.Getenv(key) == "" {
						os.Setenv(key, val)
					}
				}
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
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
