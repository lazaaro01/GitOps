package model

import "time"

type JobType string
type JobStatus string

const (
	JobTypeDeploy JobType = "deploy"
)

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
)

type Job struct {
	ID           string     `json:"id"`
	DeploymentID string     `json:"deployment_id"`
	Type         JobType    `json:"type"`
	Status       JobStatus  `json:"status"`
	Payload      []byte     `json:"-"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type DeployJobPayload struct {
	DeploymentID string            `json:"deployment_id"`
	AppName      string            `json:"app_name"`
	ImageTag     string            `json:"image_tag"`
	EnvVars      map[string]string `json:"env_vars,omitempty"`
}
