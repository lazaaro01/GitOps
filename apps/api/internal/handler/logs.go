package handler

import (
	"strconv"

	"gitops-lite/pkg/model"
	"gitops-lite/pkg/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type LogHandler struct {
	logRepo *repository.LogRepo
}

func NewLogHandler(logRepo *repository.LogRepo) *LogHandler {
	return &LogHandler{logRepo: logRepo}
}

func (h *LogHandler) GetLogs(c *fiber.Ctx) error {
	id := c.Params("id")
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logs, err := h.logRepo.GetByDeploymentID(c.Context(), id, offset)
	if err != nil {
		log.Error().Err(err).Str("deploy_id", id).Msg("failed to get logs")
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to get logs",
		})
	}

	totalLines := len(logs)

	return c.JSON(model.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"lines":       logs,
			"total_lines": totalLines,
		},
	})
}

func (h *LogHandler) DownloadLogs(c *fiber.Ctx) error {
	id := c.Params("id")

	logs, err := h.logRepo.GetByDeploymentID(c.Context(), id, 0)
	if err != nil {
		log.Error().Err(err).Str("deploy_id", id).Msg("failed to get logs for download")
		return c.Status(500).JSON(model.ErrorResponse{
			Success: false,
			Error:   "failed to get logs",
		})
	}

	var text string
	for _, l := range logs {
		timeStr := l.CreatedAt.Format("2006-01-02 15:04:05")
		text += timeStr + " [" + string(l.Level) + "] [" + l.Step + "] " + l.Message + "\n"
	}

	c.Set("Content-Type", "text/plain; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=deploy-"+id[:8]+".log")
	return c.SendString(text)
}
