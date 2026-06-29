package handler

import (
	"strconv"

	"gitops-lite/pkg/model"
	"gitops-lite/pkg/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type DeploymentHandler struct {
	deployRepo *repository.DeploymentRepo
	logRepo    *repository.LogRepo
}

func NewDeploymentHandler(deployRepo *repository.DeploymentRepo, logRepo *repository.LogRepo) *DeploymentHandler {
	return &DeploymentHandler{deployRepo: deployRepo, logRepo: logRepo}
}

func (h *DeploymentHandler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	deployments, total, err := h.deployRepo.List(c.Context(), limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("failed to list deployments")
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to list deployments",
		})
	}

	return c.JSON(model.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"deployments": deployments,
			"total":       total,
			"limit":       limit,
			"offset":      offset,
		},
	})
}

func (h *DeploymentHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	deploy, err := h.deployRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("failed to get deployment")
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

	logs, _ := h.logRepo.GetByDeploymentID(c.Context(), id, 0)

	return c.JSON(model.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"deployment": deploy,
			"logs":       logs,
		},
	})
}

func (h *DeploymentHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")

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

	if deploy.Status != model.StatusPending && deploy.Status != model.StatusQueued {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Error:   "only pending or queued deployments can be cancelled",
		})
	}

	if err := h.deployRepo.UpdateStatus(c.Context(), id, model.StatusCancelled, nil); err != nil {
		log.Error().Err(err).Str("id", id).Msg("failed to cancel deployment")
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to cancel deployment",
		})
	}

	return c.JSON(model.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "deployment cancelled"},
	})
}
