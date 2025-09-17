package handlers

import (
	"bytes"
	"encoding/json"
	"fiber-api/schemas"
	"fiber-api/services"
	"fiber-api/telemetry"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCartHandler_AddToCart_Success(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	app := fiber.New()
	cartService := services.NewCartService(mockProvider)
	handler := NewCartHandler(cartService, mockProvider)

	app.Post("/cart", handler.AddToCart)

	cartRequest := schemas.CartRequest{
		UserID: "user123",
		Items: []schemas.Item{
			{
				ID:       "item1",
				Name:     "Product A",
				Price:    29.99,
				Quantity: 2,
			},
			{
				ID:       "item2",
				Name:     "Product B",
				Price:    15.50,
				Quantity: 1,
			},
		},
	}

	body, err := json.Marshal(cartRequest)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/cart", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var response schemas.CartResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "user123", response.UserID)
	assert.Len(t, response.Items, 2)
	assert.InDelta(t, 75.48, response.Total, 0.01)
	assert.NotEmpty(t, response.ID)
}

func TestCartHandler_AddToCart_InvalidJSON(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	app := fiber.New()
	cartService := services.NewCartService(mockProvider)
	handler := NewCartHandler(cartService, mockProvider)

	app.Post("/cart", handler.AddToCart)

	req := httptest.NewRequest("POST", "/cart", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestCartHandler_AddToCart_MissingUserID(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	app := fiber.New()
	cartService := services.NewCartService(mockProvider)
	handler := NewCartHandler(cartService, mockProvider)

	app.Post("/cart", handler.AddToCart)

	cartRequest := schemas.CartRequest{
		Items: []schemas.Item{
			{
				ID:       "item1",
				Name:     "Product A",
				Price:    29.99,
				Quantity: 2,
			},
		},
	}

	body, err := json.Marshal(cartRequest)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/cart", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestCartHandler_AddToCart_EmptyItems(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	app := fiber.New()
	cartService := services.NewCartService(mockProvider)
	handler := NewCartHandler(cartService, mockProvider)

	app.Post("/cart", handler.AddToCart)

	cartRequest := schemas.CartRequest{
		UserID: "user123",
		Items:  []schemas.Item{},
	}

	body, err := json.Marshal(cartRequest)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/cart", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 400, resp.StatusCode)
}
