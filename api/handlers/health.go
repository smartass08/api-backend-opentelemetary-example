package handlers

import (
	"fiber-api/schemas"
	"fiber-api/telemetry"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
)

type HealthHandler struct {
	metricsExporter telemetry.MetricsExporter
}

func NewHealthHandler(telemetryProvider telemetry.TelemetryProvider) *HealthHandler {
	return &HealthHandler{
		metricsExporter: telemetryProvider.GetMetricsExporter(),
	}
}

func (h *HealthHandler) GetHealth(c *fiber.Ctx) error {
	// Record health check metric
	attributes := []attribute.KeyValue{
		attribute.String("endpoint", "health"),
		attribute.String("status", "ok"),
	}
	h.metricsExporter.RecordCounter(c.Context(), schemas.HealthChecksTotal, 1, attributes)

	// Log health check
	ctx := c.UserContext()
	slog.InfoContext(ctx, "Health check endpoint called")

	response := schemas.HealthResponse{
		Status:    "ok",
		Message:   "Server is running",
		Timestamp: time.Now(),
	}

	return c.JSON(response)
}

func (h *HealthHandler) GetError(c *fiber.Ctx) error {
	// Record intentional error metric
	attributes := []attribute.KeyValue{
		attribute.String("endpoint", "error"),
		attribute.String("type", "intentional_error"),
	}
	h.metricsExporter.RecordCounter(c.Context(), schemas.IntentionalErrorsTotal, 1, attributes)

	// Log error endpoint call
	ctx := c.UserContext()
	slog.ErrorContext(ctx, "Error endpoint called - this always logs as error")

	response := schemas.ErrorResponse{
		Error:     true,
		Message:   "This endpoint always returns an error",
		Timestamp: time.Now(),
	}

	return c.Status(500).JSON(response)
}
