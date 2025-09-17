package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type MetricsExporter interface {
	RecordMetric(ctx context.Context, metricName string, value interface{}, attributes []attribute.KeyValue)
	RecordCounter(ctx context.Context, name string, value int64, attributes []attribute.KeyValue)
	RecordHistogram(ctx context.Context, name string, value float64, attributes []attribute.KeyValue)
	RecordGauge(ctx context.Context, name string, value float64, attributes []attribute.KeyValue)
}

type TracesExporter interface {
	StartSpan(ctx context.Context, spanName string) (context.Context, func())
	AddSpanEvent(ctx context.Context, eventName string, attributes []attribute.KeyValue)
}

type TelemetryProvider interface {
	GetMetricsExporter() MetricsExporter
	GetLogger() *slog.Logger
	GetTracesExporter() TracesExporter
	GetTracerProvider() *sdktrace.TracerProvider
	Shutdown(ctx context.Context) error
}
