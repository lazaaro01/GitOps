package handler

import (
	"context"
	"encoding/json"

	"gitops-lite/pkg/model"
	"gitops-lite/pkg/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type DeployPublisher interface {
	PublishDeploy(ctx context.Context, payload model.DeployJobPayload) error
}

type DeployHandler struct {
	deployRepo *repository.DeploymentRepo
	jobRepo    *repository.JobRepo
	publisher  DeployPublisher
}

func NewDeployHandler(deployRepo *repository.DeploymentRepo, jobRepo *repository.JobRepo, publisher DeployPublisher) *DeployHandler {
	return &DeployHandler{
		deployRepo: deployRepo,
		jobRepo:    jobRepo,
		publisher:  publisher,
	}
}

func (h *DeployHandler) Create(c *fiber.Ctx) error {
	var req model.CreateDeployRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
	}

	if req.AppName == "" || req.ImageTag == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Error:   "app_name and image_tag are required",
		})
	}

	params := model.DeployParams{
		AppName:  req.AppName,
		ImageTag: req.ImageTag,
		EnvVars:  req.EnvVars,
	}

	deploy, err := h.deployRepo.Create(c.Context(), params)
	if err != nil {
		log.Error().Err(err).Msg("failed to create deployment")
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to create deployment",
		})
	}

	jobPayload := model.DeployJobPayload{
		DeploymentID: deploy.ID,
		AppName:      req.AppName,
		ImageTag:     req.ImageTag,
		EnvVars:      req.EnvVars,
	}

	payloadBytes, _ := json.Marshal(jobPayload)

	job, err := h.jobRepo.Create(c.Context(), deploy.ID, model.JobTypeDeploy, payloadBytes)
	if err != nil {
		log.Error().Err(err).Msg("failed to create job record")
	}

	if err := h.publisher.PublishDeploy(c.Context(), jobPayload); err != nil {
		log.Error().Err(err).Msg("failed to publish deploy job, falling back")
		h.deployRepo.UpdateStatus(c.Context(), deploy.ID, model.StatusFailed, strPtr("queue publish failed"))
		h.jobRepo.UpdateStatus(c.Context(), job.ID, model.JobFailed, strPtr("queue publish failed"))
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to queue deployment",
		})
	}

	h.deployRepo.UpdateStatus(c.Context(), deploy.ID, model.StatusQueued, nil)
	if job != nil {
		h.jobRepo.UpdateStatus(c.Context(), job.ID, model.JobRunning, nil)
	}

	return c.Status(202).JSON(model.APIResponse{
		Success: true,
		Data:    deploy,
	})
}

func strPtr(s string) *string {
	return &s
}
