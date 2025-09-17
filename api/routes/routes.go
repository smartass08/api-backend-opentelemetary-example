package routes

import (
	"fiber-api/api/handlers"
	"fiber-api/services"
	"fiber-api/telemetry"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, telemetryProvider telemetry.TelemetryProvider) {
	cartService := services.NewCartService(telemetryProvider)

	healthHandler := handlers.NewHealthHandler(telemetryProvider)
	cartHandler := handlers.NewCartHandler(cartService, telemetryProvider)

	api := app.Group("/api/v1")

	api.Get("/health", healthHandler.GetHealth)
	api.Get("/error", healthHandler.GetError)
	api.Post("/cart", cartHandler.AddToCart)
}
