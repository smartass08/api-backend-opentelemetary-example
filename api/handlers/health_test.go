package handlers

import (
	"fiber-api/telemetry"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_GetHealth(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	app := fiber.New()
	handler := NewHealthHandler(mockProvider)

	app.Get("/health", handler.GetHealth)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestHealthHandler_GetError(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	app := fiber.New()
	handler := NewHealthHandler(mockProvider)

	app.Get("/error", handler.GetError)

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
