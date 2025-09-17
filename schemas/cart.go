package schemas

import "time"

type Item struct {
	ID       string  `json:"id" validate:"required"`
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	Price    float64 `json:"price" validate:"required,min=0"`
	Quantity int     `json:"quantity" validate:"required,min=1"`
}

type CartRequest struct {
	UserID string `json:"userId" validate:"required"`
	Items  []Item `json:"items" validate:"required,min=1,dive"`
}

type CartResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Items     []Item    `json:"items"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorResponse struct {
	Error     bool      `json:"error"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}