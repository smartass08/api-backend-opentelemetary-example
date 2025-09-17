package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// MockMetricsExporter is a mock implementation of MetricsExporter
type MockMetricsExporter struct{}

func (m *MockMetricsExporter) RecordMetric(ctx context.Context, metricName string, value interface{}, attributes []attribute.KeyValue) {
	// No-op for testing
}

func (m *MockMetricsExporter) RecordCounter(ctx context.Context, name string, value int64, attributes []attribute.KeyValue) {
	// No-op for testing
}

func (m *MockMetricsExporter) RecordHistogram(ctx context.Context, name string, value float64, attributes []attribute.KeyValue) {
	// No-op for testing
}

func (m *MockMetricsExporter) RecordGauge(ctx context.Context, name string, value float64, attributes []attribute.KeyValue) {
	// No-op for testing
}

// MockTracesExporter is a mock implementation of TracesExporter
type MockTracesExporter struct{}

func (m *MockTracesExporter) StartSpan(ctx context.Context, spanName string) (context.Context, func()) {
	return ctx, func() {}
}

func (m *MockTracesExporter) AddSpanEvent(ctx context.Context, eventName string, attributes []attribute.KeyValue) {
	// No-op for testing
}

// MockTelemetryProvider is a mock implementation of TelemetryProvider
type MockTelemetryProvider struct {
	mockMetricsExporter *MockMetricsExporter
	mockTracesExporter  *MockTracesExporter
	mockLogger          *slog.Logger
}

func NewMockTelemetryProvider() *MockTelemetryProvider {
	return &MockTelemetryProvider{
		mockMetricsExporter: &MockMetricsExporter{},
		mockTracesExporter:  &MockTracesExporter{},
		mockLogger:          slog.New(slog.NewTextHandler(&mockWriter{}, &slog.HandlerOptions{})),
	}
}

func (m *MockTelemetryProvider) GetMetricsExporter() MetricsExporter {
	return m.mockMetricsExporter
}

func (m *MockTelemetryProvider) GetLogger() *slog.Logger {
	return m.mockLogger
}

func (m *MockTelemetryProvider) GetTracesExporter() TracesExporter {
	return m.mockTracesExporter
}

func (m *MockTelemetryProvider) GetTracerProvider() *sdktrace.TracerProvider {
	return nil // Not needed for testing
}

func (m *MockTelemetryProvider) Shutdown(ctx context.Context) error {
	return nil // No-op for testing
}

// mockWriter is a mock writer for the logger
type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
