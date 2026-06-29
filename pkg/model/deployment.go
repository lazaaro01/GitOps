package model

import "time"

type DeploymentStatus string

const (
	StatusPending    DeploymentStatus = "pending"
	StatusQueued     DeploymentStatus = "queued"
	StatusInProgress DeploymentStatus = "in_progress"
	StatusSuccess    DeploymentStatus = "success"
	StatusFailed     DeploymentStatus = "failed"
	StatusCancelled  DeploymentStatus = "cancelled"
)

type DeployParams struct {
	AppName  string            `json:"app_name"`
	ImageTag string            `json:"image_tag"`
	EnvVars  map[string]string `json:"env_vars,omitempty"`
	Replicas int               `json:"replicas,omitempty"`
}

type Deployment struct {
	ID           string           `json:"id"`
	AppName      string           `json:"app_name"`
	ImageTag     string           `json:"image_tag"`
	Status       DeploymentStatus `json:"status"`
	ParamsJSON   []byte           `json:"-"`
	ErrorMessage *string          `json:"error_message,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	FinishedAt   *time.Time       `json:"finished_at,omitempty"`
}

type CreateDeployRequest struct {
	AppName  string            `json:"app_name"`
	ImageTag string            `json:"image_tag"`
	EnvVars  map[string]string `json:"env_vars,omitempty"`
}
