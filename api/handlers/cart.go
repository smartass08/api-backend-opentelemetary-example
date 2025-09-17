package handlers

import (
	"fiber-api/schemas"
	"fiber-api/services"
	"fiber-api/telemetry"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
)

type CartHandler struct {
	cartService     *services.CartService
	metricsExporter telemetry.MetricsExporter
}

func NewCartHandler(cartService *services.CartService, telemetryProvider telemetry.TelemetryProvider) *CartHandler {
	return &CartHandler{
		cartService:     cartService,
		metricsExporter: telemetryProvider.GetMetricsExporter(),
	}
}

func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	var req schemas.CartRequest
	ctx := c.UserContext()

	if err := c.BodyParser(&req); err != nil {
		slog.ErrorContext(ctx, "Failed to parse cart request", "error", err.Error(), "type", "parse_error")
		return c.Status(400).JSON(schemas.ErrorResponse{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	if req.UserID == "" {
		slog.ErrorContext(ctx, "Missing user ID in cart request", "type", "validation_error")
		return c.Status(400).JSON(schemas.ErrorResponse{
			Error:   true,
			Message: "User ID is required",
		})
	}

	if len(req.Items) == 0 {
		slog.ErrorContext(ctx, "Empty items list in cart request", "type", "validation_error")
		return c.Status(400).JSON(schemas.ErrorResponse{
			Error:   true,
			Message: "At least one item is required",
		})
	}

	response, err := h.cartService.ProcessCart(c.UserContext(), req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to process cart", "error", err.Error(), "type", "processing_error")
		return c.Status(500).JSON(schemas.ErrorResponse{
			Error:   true,
			Message: "Failed to process cart",
		})
	}

	// Record successful cart operation metrics
	attributes := []attribute.KeyValue{
		attribute.String("user_id", req.UserID),
		attribute.Int("item_count", len(req.Items)),
		attribute.String("status", "success"),
	}
	h.metricsExporter.RecordCounter(c.Context(), schemas.CartRequestsTotal, 1, attributes)
	h.metricsExporter.RecordHistogram(c.Context(), schemas.CartItemsPerRequest, float64(len(req.Items)), attributes)

	// Record gauge metrics for current cart state
	gaugeAttributes := []attribute.KeyValue{
		attribute.String("user_id", req.UserID),
	}
	h.metricsExporter.RecordGauge(c.Context(), schemas.CartCurrentValue, response.Total, gaugeAttributes)
	h.metricsExporter.RecordGauge(c.Context(), schemas.CartCurrentItems, float64(len(req.Items)), gaugeAttributes)

	// Log successful cart operation
	slog.InfoContext(ctx, "Cart processed successfully",
		"cartId", response.ID,
		"userId", req.UserID,
		"itemCount", len(req.Items),
		"total", response.Total)

	return c.Status(201).JSON(response)
}
