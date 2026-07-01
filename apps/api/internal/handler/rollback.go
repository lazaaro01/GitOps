package handler

import (
	"context"
	"encoding/json"

	"gitops-lite/pkg/model"
	"gitops-lite/pkg/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type RollbackPublisher interface {
	PublishDeploy(ctx context.Context, payload model.DeployJobPayload) error
}

type DeployHandlerOps struct {
	deployRepo *repository.DeploymentRepo
	jobRepo    *repository.JobRepo
	publisher  RollbackPublisher
}

func NewDeployHandlerOps(deployRepo *repository.DeploymentRepo, jobRepo *repository.JobRepo, publisher RollbackPublisher) *DeployHandlerOps {
	return &DeployHandlerOps{deployRepo: deployRepo, jobRepo: jobRepo, publisher: publisher}
}

func (h *DeployHandlerOps) Rollback(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		TargetVersion string `json:"target_version"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
	}
	if req.TargetVersion == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Error:   "target_version is required",
		})
	}

	deploy, err := h.deployRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to get deployment",
		})
	}
	if deploy == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Error:   "deployment not found",
		})
	}

	payload := model.DeployJobPayload{
		DeploymentID: deploy.ID,
		AppName:      deploy.AppName,
		ImageTag:     req.TargetVersion,
	}

	if deploy.ParamsJSON != nil {
		var params model.DeployParams
		if err := json.Unmarshal(deploy.ParamsJSON, &params); err == nil {
			payload.EnvVars = params.EnvVars
		}
	}

	payloadBytes, _ := json.Marshal(payload)
	job, err := h.jobRepo.Create(c.Context(), deploy.ID, model.JobTypeRollback, payloadBytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to create rollback job record")
	}

	if err := h.publisher.PublishDeploy(c.Context(), payload); err != nil {
		log.Error().Err(err).Msg("failed to publish rollback job")
		if job != nil {
			h.jobRepo.UpdateStatus(c.Context(), job.ID, model.JobFailed, strPtr("queue publish failed"))
		}
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to queue rollback",
		})
	}

	if job != nil {
		h.jobRepo.UpdateStatus(c.Context(), job.ID, model.JobRunning, nil)
	}

	return c.Status(202).JSON(model.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"rollback_job_id": job.ID,
			"status":          "rollback_queued",
			"message":         "Rollback para " + req.TargetVersion + " foi enfileirado",
		},
	})
}

func (h *DeployHandlerOps) Retry(c *fiber.Ctx) error {
	id := c.Params("id")

	original, err := h.deployRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to get deployment",
		})
	}
	if original == nil {
		return c.Status(404).JSON(model.ErrorResponse{
			Success: false,
			Error:   "deployment not found",
		})
	}

	var params model.DeployParams
	if original.ParamsJSON != nil {
		if err := json.Unmarshal(original.ParamsJSON, &params); err != nil {
			params = model.DeployParams{
				AppName:  original.AppName,
				ImageTag: original.ImageTag,
			}
		}
	} else {
		params = model.DeployParams{
			AppName:  original.AppName,
			ImageTag: original.ImageTag,
		}
	}

	newDeploy, err := h.deployRepo.Create(c.Context(), params)
	if err != nil {
		log.Error().Err(err).Msg("failed to create retry deployment")
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to create retry deployment",
		})
	}

	payload := model.DeployJobPayload{
		DeploymentID: newDeploy.ID,
		AppName:      params.AppName,
		ImageTag:     params.ImageTag,
		EnvVars:      params.EnvVars,
	}

	payloadBytes, _ := json.Marshal(payload)
	job, err := h.jobRepo.Create(c.Context(), newDeploy.ID, model.JobTypeRetry, payloadBytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to create retry job record")
	}

	if err := h.publisher.PublishDeploy(c.Context(), payload); err != nil {
		log.Error().Err(err).Msg("failed to publish retry job")
		h.deployRepo.UpdateStatus(c.Context(), newDeploy.ID, model.StatusFailed, strPtr("queue publish failed"))
		if job != nil {
			h.jobRepo.UpdateStatus(c.Context(), job.ID, model.JobFailed, strPtr("queue publish failed"))
		}
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to queue retry",
		})
	}

	h.deployRepo.UpdateStatus(c.Context(), newDeploy.ID, model.StatusQueued, nil)
	if job != nil {
		h.jobRepo.UpdateStatus(c.Context(), job.ID, model.JobRunning, nil)
	}

	return c.Status(202).JSON(model.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"retry_job_id":  job.ID,
			"status":        "retry_queued",
			"message":       "Reexecução do deploy enfileirada",
			"deployment_id": newDeploy.ID,
		},
	})
}
