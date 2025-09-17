package telemetry

import (
	"context"
	"fiber-api/config"
	"fiber-api/schemas"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const signozHeaderName = `signoz-ingestion-key`

// LoggingMetricExporter wraps the OTLP exporter and logs export attempts
type LoggingMetricExporter struct {
	exporter sdkmetric.Exporter
}

func (e *LoggingMetricExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {

	err := e.exporter.Export(ctx, rm)
	if err != nil {
		slog.Error("❌ METRIC EXPORT FAILED", "error", err)
	} else {

	}
	return err
}

func (e *LoggingMetricExporter) ForceFlush(ctx context.Context) error {

	return e.exporter.ForceFlush(ctx)
}

func (e *LoggingMetricExporter) Shutdown(ctx context.Context) error {

	return e.exporter.Shutdown(ctx)
}

func (e *LoggingMetricExporter) Temporality(ik sdkmetric.InstrumentKind) metricdata.Temporality {
	return e.exporter.Temporality(ik)
}

func (e *LoggingMetricExporter) Aggregation(ik sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return e.exporter.Aggregation(ik)
}

func parseLogLevel(logLevel string) slog.Level {
	switch strings.ToLower(logLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type DefaultTelemetryProvider struct {
	loggerProvider *log.LoggerProvider
	meterProvider  *sdkmetric.MeterProvider
	tracerProvider *sdktrace.TracerProvider
	meter          metric.Meter
	tracer         trace.Tracer
	logger         *slog.Logger
}

func NewTelemetryProvider(serviceName, serviceVersion string) (TelemetryProvider, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	loggerProvider, err := setupLogs(ctx, res)
	if err != nil {
		return nil, err
	}

	meterProvider, err := setupMetrics(ctx, res)
	if err != nil {
		return nil, err
	}

	tracerProvider, err := setupTraces(ctx, res)
	if err != nil {
		return nil, err
	}

	// Set up propagation for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	meter := meterProvider.Meter(serviceName)
	tracer := tracerProvider.Tracer(serviceName)

	// Get the configured log level
	cfg := config.GetConfig()
	logLevel := parseLogLevel(cfg.LogLevel)

	// Create otelslog handler that bridges slog to OpenTelemetry
	baseOtelHandler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(loggerProvider))

	// Apply level filtering to otel handler
	otelHandler := NewLevelFilterHandler(baseOtelHandler, logLevel)

	// Create a console handler with level filtering
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Create a custom handler that sends to both destinations
	multiHandler := NewMultiHandler(otelHandler, consoleHandler)

	logger := slog.New(multiHandler)

	// Set the default slog logger to use our multi-handler
	slog.SetDefault(logger)

	return &DefaultTelemetryProvider{
		loggerProvider: loggerProvider,
		meterProvider:  meterProvider,
		tracerProvider: tracerProvider,
		meter:          meter,
		tracer:         tracer,
		logger:         logger,
	}, nil
}

func (p *DefaultTelemetryProvider) GetMetricsExporter() MetricsExporter {
	// Pre-create all metric instruments to ensure they're cached and reused
	httpRequestsCounter, err := p.meter.Int64Counter(schemas.HTTPRequestsTotal)
	if err != nil {
		slog.Error("Failed to create http_requests_total counter", "error", err)
	}

	httpDurationHistogram, err := p.meter.Float64Histogram(schemas.HTTPRequestDurationSeconds)
	if err != nil {
		slog.Error("Failed to create http_request_duration_seconds histogram", "error", err)
	}

	activeRequestsGauge, err := p.meter.Float64Gauge(schemas.HTTPActiveRequests)
	if err != nil {
		slog.Error("Failed to create http_active_requests gauge", "error", err)
	}

	errorsCounter, err := p.meter.Int64Counter(schemas.ErrorsTotal)
	if err != nil {
		slog.Error("Failed to create errors_total counter", "error", err)
	}

	cartItemsGauge, err := p.meter.Float64Gauge(schemas.CartCurrentItems)
	if err != nil {
		slog.Error("Failed to create cart_current_items gauge", "error", err)
	}

	cartRequestsCounter, err := p.meter.Int64Counter(schemas.CartRequestsTotal)
	if err != nil {
		slog.Error("Failed to create cart_requests_total counter", "error", err)
	}

	cartItemsPerRequestHistogram, err := p.meter.Float64Histogram(schemas.CartItemsPerRequest)
	if err != nil {
		slog.Error("Failed to create cart_items_per_request histogram", "error", err)
	}

	return &DefaultMetricsExporter{
		meter:                        p.meter,
		httpRequestsCounter:          httpRequestsCounter,
		httpDurationHistogram:        httpDurationHistogram,
		activeRequestsGauge:          activeRequestsGauge,
		errorsCounter:                errorsCounter,
		cartItemsGauge:               cartItemsGauge,
		cartRequestsCounter:          cartRequestsCounter,
		cartItemsPerRequestHistogram: cartItemsPerRequestHistogram,
	}
}

func (p *DefaultTelemetryProvider) GetLogger() *slog.Logger {
	return p.logger
}

func (p *DefaultTelemetryProvider) GetTracesExporter() TracesExporter {
	return &DefaultTracesExporter{tracer: p.tracer}
}

func (p *DefaultTelemetryProvider) GetTracerProvider() *sdktrace.TracerProvider {
	return p.tracerProvider
}

func (p *DefaultTelemetryProvider) Shutdown(ctx context.Context) error {
	if err := p.loggerProvider.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown logger provider", "error", err)
		return err
	}

	if err := p.meterProvider.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown meter provider", "error", err)
		return err
	}

	if err := p.tracerProvider.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown tracer provider", "error", err)
		return err
	}

	return nil
}

type DefaultMetricsExporter struct {
	meter                        metric.Meter
	httpRequestsCounter          metric.Int64Counter
	httpDurationHistogram        metric.Float64Histogram
	activeRequestsGauge          metric.Float64Gauge
	errorsCounter                metric.Int64Counter
	cartItemsGauge               metric.Float64Gauge
	cartRequestsCounter          metric.Int64Counter
	cartItemsPerRequestHistogram metric.Float64Histogram
}

func (e *DefaultMetricsExporter) RecordMetric(ctx context.Context, metricName string, value interface{}, attributes []attribute.KeyValue) {
	switch v := value.(type) {
	case int64:
		e.RecordCounter(ctx, metricName, v, attributes)
	case float64:
		e.RecordHistogram(ctx, metricName, v, attributes)
	}
}

func (e *DefaultMetricsExporter) RecordCounter(ctx context.Context, name string, value int64, attributes []attribute.KeyValue) {

	var counter metric.Int64Counter
	switch name {
	case schemas.HTTPRequestsTotal:
		counter = e.httpRequestsCounter
	case schemas.ErrorsTotal:
		counter = e.errorsCounter
	case schemas.CartRequestsTotal:
		counter = e.cartRequestsCounter
	default:
		// Fallback to creating new counter
		var err error
		counter, err = e.meter.Int64Counter(name)
		if err != nil {
			slog.Error("Failed to create counter", "name", name, "error", err)
			return
		}
	}

	counter.Add(ctx, value, metric.WithAttributes(attributes...))

}

func (e *DefaultMetricsExporter) RecordHistogram(ctx context.Context, name string, value float64, attributes []attribute.KeyValue) {
	var histogram metric.Float64Histogram
	switch name {
	case schemas.HTTPRequestDurationSeconds:
		histogram = e.httpDurationHistogram
	case schemas.CartItemsPerRequest:
		histogram = e.cartItemsPerRequestHistogram
	default:
		// Fallback to creating new histogram
		var err error
		histogram, err = e.meter.Float64Histogram(name)
		if err != nil {
			slog.Error("Failed to create histogram", "name", name, "error", err)
			return
		}
	}

	histogram.Record(ctx, value, metric.WithAttributes(attributes...))

}

func (e *DefaultMetricsExporter) RecordGauge(ctx context.Context, name string, value float64, attributes []attribute.KeyValue) {
	gauge, err := e.meter.Float64Gauge(name)
	if err != nil {
		slog.Error("Failed to create gauge", "name", name, "error", err)
		return
	}

	gauge.Record(ctx, value, metric.WithAttributes(attributes...))
}

type DefaultTracesExporter struct {
	tracer trace.Tracer
}

func (e *DefaultTracesExporter) StartSpan(ctx context.Context, spanName string) (context.Context, func()) {
	ctx, span := e.tracer.Start(ctx, spanName)
	return ctx, func() { span.End() }
}

func (e *DefaultTracesExporter) AddSpanEvent(ctx context.Context, eventName string, attributes []attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(eventName, trace.WithAttributes(attributes...))
	}
}

func setupLogs(ctx context.Context, res *resource.Resource) (*log.LoggerProvider, error) {
	var exporterOptions []otlploggrpc.Option
	cfg := config.GetConfig()

	exporterOptions = append(exporterOptions, otlploggrpc.WithEndpoint(cfg.OTLPEndpoint))
	//exporterOptions = append(exporterOptions, otlploggrpc.WithInsecure())
	// Add authentication if API key is provided
	if cfg.OtelAPIKey != "" {
		headers := map[string]string{
			signozHeaderName: cfg.OtelAPIKey,
		}
		exporterOptions = append(exporterOptions, otlploggrpc.WithHeaders(headers))
	}
	exporter, err := otlploggrpc.New(ctx, exporterOptions...)
	if err != nil {
		return nil, err
	}

	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(processor),
	)

	global.SetLoggerProvider(provider)
	return provider, nil
}

func setupMetrics(ctx context.Context, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	var exporterOptions []otlpmetricgrpc.Option
	cfg := config.GetConfig()

	exporterOptions = append(exporterOptions, otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint))
	//exporterOptions = append(exporterOptions, otlpmetricgrpc.WithInsecure())
	// Add authentication if API key is provided
	if cfg.OtelAPIKey != "" {
		headers := map[string]string{
			signozHeaderName: cfg.OtelAPIKey,
		}
		exporterOptions = append(exporterOptions, otlpmetricgrpc.WithHeaders(headers))

	}

	baseExporter, err := otlpmetricgrpc.New(ctx, exporterOptions...)
	if err != nil {
		slog.Error("Failed to create metrics exporter", "error", err)
		return nil, err
	}

	// Wrap with logging exporter to track export attempts
	loggingExporter := &LoggingMetricExporter{exporter: baseExporter}

	// Use a fast 1-second export interval for testing
	reader := sdkmetric.NewPeriodicReader(loggingExporter,
		sdkmetric.WithInterval(1*time.Second),
	)

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	otel.SetMeterProvider(provider)

	return provider, nil
}

func setupTraces(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	var exporterOptions []otlptracegrpc.Option
	cfg := config.GetConfig()

	exporterOptions = append(exporterOptions, otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint))
	//exporterOptions = append(exporterOptions, otlptracegrpc.WithInsecure())
	// Add authentication if an API key is provided
	if cfg.OtelAPIKey != "" {
		headers := map[string]string{
			signozHeaderName: cfg.OtelAPIKey,
		}
		exporterOptions = append(exporterOptions, otlptracegrpc.WithHeaders(headers))
	}

	exporter, err := otlptracegrpc.New(ctx, exporterOptions...)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(provider)
	return provider, nil
}
