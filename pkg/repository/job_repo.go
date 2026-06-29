package repository

import (
	"context"

	"gitops-lite/pkg/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepo struct {
	pool *pgxpool.Pool
}

func NewJobRepo(pool *pgxpool.Pool) *JobRepo {
	return &JobRepo{pool: pool}
}

func (r *JobRepo) Create(ctx context.Context, deploymentID string, jobType model.JobType, payload []byte) (*model.Job, error) {
	j := &model.Job{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO jobs (deployment_id, type, status, payload)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, deployment_id, type, status, payload, created_at, updated_at`,
		deploymentID, jobType, model.JobPending, payload,
	).Scan(&j.ID, &j.DeploymentID, &j.Type, &j.Status, &j.Payload, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (r *JobRepo) UpdateStatus(ctx context.Context, id string, status model.JobStatus, errMsg *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE jobs SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3`,
		status, errMsg, id,
	)
	return err
}
