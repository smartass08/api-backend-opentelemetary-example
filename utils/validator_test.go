package utils

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name     string `validate:"required,min=3,max=10"`
	Email    string `validate:"required,email"`
	Age      int    `validate:"required,min=18"`
	Optional string
}

func TestValidateStruct_Valid(t *testing.T) {
	validStruct := TestStruct{
		Name:     "John",
		Email:    "john@example.com",
		Age:      25,
		Optional: "optional",
	}

	errors := ValidateStruct(validStruct)
	assert.Empty(t, errors)
}

func TestValidateStruct_RequiredFieldMissing(t *testing.T) {
	invalidStruct := TestStruct{
		Email: "john@example.com",
		Age:   25,
	}

	errors := ValidateStruct(invalidStruct)
	assert.Contains(t, errors, "name is required")
}

func TestValidateStruct_MinLength(t *testing.T) {
	invalidStruct := TestStruct{
		Name:  "Jo",
		Email: "john@example.com",
		Age:   25,
	}

	errors := ValidateStruct(invalidStruct)
	assert.Contains(t, errors, "name must be at least 3")
}

func TestValidateStruct_MaxLength(t *testing.T) {
	invalidStruct := TestStruct{
		Name:  "ThisNameIsTooLong",
		Email: "john@example.com",
		Age:   25,
	}

	errors := ValidateStruct(invalidStruct)
	assert.Contains(t, errors, "name must be at most 10")
}

func TestValidateStruct_MinValue(t *testing.T) {
	invalidStruct := TestStruct{
		Name:  "John",
		Email: "john@example.com",
		Age:   16,
	}

	errors := ValidateStruct(invalidStruct)
	assert.Contains(t, errors, "age must be at least 18")
}

func TestValidateStruct_MultipleErrors(t *testing.T) {
	invalidStruct := TestStruct{
		Age: 16,
	}

	errors := ValidateStruct(invalidStruct)
	assert.Len(t, errors, 3)
	assert.Contains(t, errors, "name is required")
	assert.Contains(t, errors, "email is required")
	assert.Contains(t, errors, "age must be at least 18")
}

func TestParseLogLevel_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := ParseLogLevel(tt.input)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestParseLogLevel_Invalid(t *testing.T) {
	invalidInputs := []string{"", "invalid", "trace", "fatal"}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			level := ParseLogLevel(input)
			assert.Equal(t, slog.LevelInfo, level)
		})
	}
}