package repository

import (
	"context"

	"gitops-lite/pkg/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LogRepo struct {
	pool *pgxpool.Pool
}

func NewLogRepo(pool *pgxpool.Pool) *LogRepo {
	return &LogRepo{pool: pool}
}

func (r *LogRepo) Append(ctx context.Context, deploymentID, step string, level model.LogLevel, message string, seq int64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO deployment_logs (deployment_id, step, level, message, sequence)
		 VALUES ($1, $2, $3, $4, $5)`,
		deploymentID, step, level, message, seq,
	)
	return err
}

func (r *LogRepo) GetByDeploymentID(ctx context.Context, deploymentID string, offset int) ([]*model.DeploymentLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, deployment_id, step, level, message, sequence, created_at
		 FROM deployment_logs
		 WHERE deployment_id = $1
		 ORDER BY sequence ASC
		 OFFSET $2`,
		deploymentID, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*model.DeploymentLog
	for rows.Next() {
		l := &model.DeploymentLog{}
		if err := rows.Scan(&l.ID, &l.DeploymentID, &l.Step, &l.Level, &l.Message, &l.Sequence, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
