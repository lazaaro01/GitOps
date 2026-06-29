package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRouter(
	app *fiber.App,
	deployHandler *DeployHandler,
	deploymentHandler *DeploymentHandler,
) {
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} ${method} ${path} ${status} ${latency}\n",
	}))

	api := app.Group("/api")

	api.Post("/deploy", deployHandler.Create)

	api.Get("/deployments", deploymentHandler.List)
	api.Get("/deployments/:id", deploymentHandler.GetByID)
	api.Put("/deployments/:id/cancel", deploymentHandler.Cancel)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})
}
