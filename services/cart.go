package services

import (
	"context"
	"fiber-api/schemas"
	"fiber-api/telemetry"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

type CartService struct {
	metricsExporter telemetry.MetricsExporter
}

func NewCartService(telemetryProvider telemetry.TelemetryProvider) *CartService {
	return &CartService{
		metricsExporter: telemetryProvider.GetMetricsExporter(),
	}
}

func (s *CartService) ProcessCart(ctx context.Context, req schemas.CartRequest) (*schemas.CartResponse, error) {
	itemCount := len(req.Items)

	// Record cart processing metrics
	attributes := []attribute.KeyValue{
		attribute.String("user_id", req.UserID),
		attribute.Int("item_count", itemCount),
	}
	s.metricsExporter.RecordCounter(ctx, schemas.CartOperationsTotal, 1, attributes)
	if itemCount > 0 {
		s.metricsExporter.RecordHistogram(ctx, schemas.CartItemsTotal, float64(itemCount), attributes)
	}

	// Log cart processing with trace context
	slog.InfoContext(ctx, "Processing cart request", "userId", req.UserID, "itemCount", itemCount)

	total := s.calculateTotal(req.Items)

	response := &schemas.CartResponse{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Items:     req.Items,
		Total:     total,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Log successful processing
	slog.InfoContext(ctx, "Cart processed successfully",
		"cartId", response.ID,
		"total", total,
		"itemCount", itemCount)

	return response, nil
}

func (s *CartService) calculateTotal(items []schemas.Item) float64 {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}
