package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gitops-lite/pkg/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeploymentRepo struct {
	pool *pgxpool.Pool
}

func NewDeploymentRepo(pool *pgxpool.Pool) *DeploymentRepo {
	return &DeploymentRepo{pool: pool}
}

func (r *DeploymentRepo) Create(ctx context.Context, params model.DeployParams) (*model.Deployment, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	d := &model.Deployment{}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO deployments (app_name, image_tag, status, params_json)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, app_name, image_tag, status, params_json, created_at, updated_at`,
		params.AppName, params.ImageTag, model.StatusPending, paramsJSON,
	).Scan(&d.ID, &d.AppName, &d.ImageTag, &d.Status, &d.ParamsJSON, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DeploymentRepo) GetByID(ctx context.Context, id string) (*model.Deployment, error) {
	d := &model.Deployment{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, app_name, image_tag, status, params_json, error_message,
		        created_at, updated_at, finished_at
		 FROM deployments WHERE id = $1`, id,
	).Scan(&d.ID, &d.AppName, &d.ImageTag, &d.Status, &d.ParamsJSON,
		&d.ErrorMessage, &d.CreatedAt, &d.UpdatedAt, &d.FinishedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DeploymentRepo) List(ctx context.Context, limit, offset int) ([]*model.Deployment, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deployments`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, app_name, image_tag, status, params_json, error_message,
		        created_at, updated_at, finished_at
		 FROM deployments ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deployments []*model.Deployment
	for rows.Next() {
		d := &model.Deployment{}
		if err := rows.Scan(&d.ID, &d.AppName, &d.ImageTag, &d.Status, &d.ParamsJSON,
			&d.ErrorMessage, &d.CreatedAt, &d.UpdatedAt, &d.FinishedAt); err != nil {
			return nil, 0, err
		}
		deployments = append(deployments, d)
	}
	return deployments, total, nil
}

func (r *DeploymentRepo) UpdateStatus(ctx context.Context, id string, status model.DeploymentStatus, errMsg *string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE deployments SET status = $1, error_message = $2, updated_at = $3,
		 finished_at = CASE WHEN $1 IN ('success','failed','cancelled') THEN $3 ELSE finished_at END
		 WHERE id = $4`,
		status, errMsg, now, id,
	)
	return err
}
