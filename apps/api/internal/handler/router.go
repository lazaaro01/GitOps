package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type AllHandlers struct {
	Deploy      *DeployHandler
	Deployment  *DeploymentHandler
	Logs        *LogHandler
	RollbackOps *DeployHandlerOps
	Events      *EventsHandler
	StaticDir   string
}

func SetupRouter(app *fiber.App, h *AllHandlers) {
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} ${method} ${path} ${status} ${latency}\n",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})

	api := app.Group("/api")

	// Deploy
	api.Post("/deploy", h.Deploy.Create)

	// Deployments
	api.Get("/deployments", h.Deployment.List)
	api.Get("/deployments/:id", h.Deployment.GetByID)
	api.Put("/deployments/:id/cancel", h.Deployment.Cancel)

	// Logs
	api.Get("/deployments/:id/logs", h.Logs.GetLogs)
	api.Get("/deployments/:id/logs/download", h.Logs.DownloadLogs)

	// Rollback & Retry
	api.Post("/deployments/:id/rollback", h.RollbackOps.Rollback)
	api.Post("/deployments/:id/retry", h.RollbackOps.Retry)

	// SSE Events
	api.Get("/events", h.Events.Stream)
	api.Post("/internal/events", h.Events.InternalPublish)

	// SPA static files (must be last)
	if h.StaticDir != "" {
		app.Static("/assets", h.StaticDir+"/assets")
		app.Get("/*", func(c *fiber.Ctx) error {
			return c.SendFile(h.StaticDir + "/index.html")
		})
	}
}
