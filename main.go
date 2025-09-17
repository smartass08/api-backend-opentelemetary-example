package main

import (
	"context"
	"fiber-api/api/routes"
	"fiber-api/config"
	"fiber-api/middleware"
	"fiber-api/telemetry"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/otel"
)

func main() {
	cfg := config.LoadConfig()

	telemetryProvider, err := telemetry.NewTelemetryProvider("fiber-api", "1.0.0")
	if err != nil {

		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		telemetryProvider.Shutdown(ctx)
	}()

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler(telemetryProvider),
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Configure otelfiber with our tracer and meter providers
	app.Use(otelfiber.Middleware(
		otelfiber.WithServerName("fiber-api"),
		otelfiber.WithTracerProvider(telemetryProvider.GetTracerProvider()),
		otelfiber.WithMeterProvider(otel.GetMeterProvider()),
	))

	// Add detailed tracing middleware for granular HTTP spans
	app.Use(middleware.DetailedTracing(telemetryProvider))

	// Add our custom logger middleware for HTTP request logging and metrics
	app.Use(middleware.Logger(telemetryProvider))

	routes.SetupRoutes(app, telemetryProvider)

	go func() {
		slog.Info("Starting server", "port", cfg.Port, "environment", cfg.Environment)
		if err := app.Listen(":" + cfg.Port); err != nil {
			slog.Error("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	// Shutdown telemetry providers to flush any pending metrics, logs, and traces
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("Shutting down telemetry...")
	if err := telemetryProvider.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown telemetry", "error", err)
	} else {
		slog.Info("Telemetry shutdown completed successfully")
	}
}
