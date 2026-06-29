package model

import "time"

type LogLevel string

const (
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
	LogDebug LogLevel = "debug"
)

type DeploymentLog struct {
	ID           string    `json:"id"`
	DeploymentID string    `json:"deployment_id"`
	Step         string    `json:"step,omitempty"`
	Level        LogLevel  `json:"level"`
	Message      string    `json:"message"`
	Sequence     int64     `json:"sequence"`
	CreatedAt    time.Time `json:"created_at"`
}
