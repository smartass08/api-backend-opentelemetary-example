package middleware

import (
	"errors"
	"fiber-api/schemas"
	"fiber-api/telemetry"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
)

func Logger(telemetryProvider telemetry.TelemetryProvider) fiber.Handler {
	metricsExporter := telemetryProvider.GetMetricsExporter()

	// Track active requests count (gauge metric)
	var activeRequests int64 = 0

	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Increment active requests gauge
		activeRequests++
		gaugeAttrs := []attribute.KeyValue{
			attribute.String("endpoint", c.Path()),
		}

		metricsExporter.RecordGauge(c.Context(), schemas.HTTPActiveRequests, float64(activeRequests), gaugeAttrs)

		err := c.Next()

		// Decrement active requests gauge
		activeRequests--
		metricsExporter.RecordGauge(c.Context(), schemas.HTTPActiveRequests, float64(activeRequests), gaugeAttrs)

		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Record HTTP request metrics
		attributes := []attribute.KeyValue{
			attribute.String("method", c.Method()),
			attribute.String("path", c.Path()),
			attribute.Int("status", status),
		}
		metricsExporter.RecordCounter(c.Context(), schemas.HTTPRequestsTotal, 1, attributes)

		metricsExporter.RecordHistogram(c.Context(), schemas.HTTPRequestDurationSeconds, duration.Seconds(), attributes)

		// Record error metrics if applicable
		if status >= 400 {
			errorAttrs := []attribute.KeyValue{
				attribute.String("method", c.Method()),
				attribute.String("path", c.Path()),
				attribute.Int("status", status),
				attribute.String("type", "http_error"),
			}
			metricsExporter.RecordCounter(c.Context(), schemas.ErrorsTotal, 1, errorAttrs)
		}

		// Log HTTP request with trace context
		ctx := c.UserContext()
		slog.InfoContext(ctx, "HTTP Request",
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"duration", duration.String(),
			"ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
		)

		if err != nil {
			// Record middleware error metrics
			errorAttrs := []attribute.KeyValue{
				attribute.String("method", c.Method()),
				attribute.String("path", c.Path()),
				attribute.Int("status", status),
				attribute.String("type", "middleware_error"),
			}
			metricsExporter.RecordCounter(c.Context(), schemas.ErrorsTotal, 1, errorAttrs)

		}

		return err
	}
}

func ErrorHandler(telemetryProvider telemetry.TelemetryProvider) func(*fiber.Ctx, error) error {
	metricsExporter := telemetryProvider.GetMetricsExporter()

	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
		}

		// Record application error metrics
		attributes := []attribute.KeyValue{
			attribute.String("method", c.Method()),
			attribute.String("path", c.Path()),
			attribute.Int("status", code),
			attribute.String("type", "application_error"),
		}
		metricsExporter.RecordCounter(c.Context(), schemas.ErrorsTotal, 1, attributes)

		// Log application error with trace context
		ctx := c.UserContext()
		slog.ErrorContext(ctx, "Request error: "+err.Error(),
			"method", c.Method(),
			"path", c.Path(),
			"status", code,
			"ip", c.IP(),
		)

		return c.Status(code).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
}
