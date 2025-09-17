package services

import (
	"context"
	"fiber-api/schemas"
	"fiber-api/telemetry"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCartService_ProcessCart(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	service := NewCartService(mockProvider)

	req := schemas.CartRequest{
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

	response, err := service.ProcessCart(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "user123", response.UserID)
	assert.Len(t, response.Items, 2)
	assert.InDelta(t, 75.48, response.Total, 0.01)
	assert.NotEmpty(t, response.ID)
	assert.NotZero(t, response.CreatedAt)
	assert.NotZero(t, response.UpdatedAt)
}

func TestCartService_ProcessCart_SingleItem(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	service := NewCartService(mockProvider)

	req := schemas.CartRequest{
		UserID: "user456",
		Items: []schemas.Item{
			{
				ID:       "item3",
				Name:     "Product C",
				Price:    100.00,
				Quantity: 1,
			},
		},
	}

	response, err := service.ProcessCart(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "user456", response.UserID)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, 100.00, response.Total)
	assert.Equal(t, "item3", response.Items[0].ID)
	assert.Equal(t, "Product C", response.Items[0].Name)
}

func TestCartService_calculateTotal(t *testing.T) {
	mockProvider := telemetry.NewMockTelemetryProvider()

	service := NewCartService(mockProvider)

	tests := []struct {
		name     string
		items    []schemas.Item
		expected float64
	}{
		{
			name: "multiple items",
			items: []schemas.Item{
				{Price: 10.00, Quantity: 2},
				{Price: 5.50, Quantity: 3},
			},
			expected: 36.50,
		},
		{
			name: "single item",
			items: []schemas.Item{
				{Price: 25.99, Quantity: 1},
			},
			expected: 25.99,
		},
		{
			name:     "empty items",
			items:    []schemas.Item{},
			expected: 0.00,
		},
		{
			name: "zero price",
			items: []schemas.Item{
				{Price: 0.00, Quantity: 5},
			},
			expected: 0.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := service.calculateTotal(tt.items)
			assert.Equal(t, tt.expected, total)
		})
	}
}
