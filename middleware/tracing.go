package middleware

import (
	"fiber-api/telemetry"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
)

// DetailedTracing middleware creates granular spans for HTTP request lifecycle
func DetailedTracing(telemetryProvider telemetry.TelemetryProvider) fiber.Handler {
	tracesExporter := telemetryProvider.GetTracesExporter()

	return func(c *fiber.Ctx) error {
		start := time.Now()
		ctx := c.UserContext()

		// Create a span for request receive phase
		receiveCtx, endReceive := tracesExporter.StartSpan(ctx, c.Method()+" "+c.Path()+" http receive")
		tracesExporter.AddSpanEvent(receiveCtx, "request.received", []attribute.KeyValue{
			attribute.String("http.method", c.Method()),
			attribute.String("http.path", c.Path()),
			attribute.String("http.user_agent", c.Get("User-Agent")),
			attribute.String("http.remote_addr", c.IP()),
			attribute.String("http.content_type", c.Get("Content-Type")),
			attribute.Int64("http.request.size", int64(len(c.Body()))),
		})
		endReceive()

		// Create a span for request processing phase
		processCtx, endProcess := tracesExporter.StartSpan(ctx, c.Method()+" "+c.Path()+" http process")
		tracesExporter.AddSpanEvent(processCtx, "processing.started", []attribute.KeyValue{
			attribute.String("http.method", c.Method()),
			attribute.String("http.path", c.Path()),
		})

		// Update the context for downstream handlers
		c.SetUserContext(processCtx)

		err := c.Next()

		tracesExporter.AddSpanEvent(processCtx, "processing.completed", []attribute.KeyValue{
			attribute.String("processing.status", func() string {
				if err != nil {
					return "error"
				}
				return "success"
			}()),
		})
		endProcess()

		// Create a span for response send phase
		sendCtx, endSend := tracesExporter.StartSpan(ctx, c.Method()+" "+c.Path()+" http send")

		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Add span events for response details
		tracesExporter.AddSpanEvent(sendCtx, "response.sending", []attribute.KeyValue{
			attribute.Int("http.status_code", status),
			attribute.Int64("http.response.size", int64(len(c.Response().Body()))),
			attribute.Float64("http.duration.ms", float64(duration.Nanoseconds())/1000000),
			attribute.String("response.type", func() string {
				if status >= 400 {
					return "error"
				} else if status >= 300 {
					return "redirect"
				}
				return "success"
			}()),
		})

		// End the send span
		endSend()

		return err
	}
}