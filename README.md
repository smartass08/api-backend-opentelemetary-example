# Go Fiber API with OpenTelemetry

A simple example showing how to integrate OpenTelemetry with Go Fiber for observability.

## What's included

- Go Fiber web framework
- OpenTelemetry metrics, traces, and logs
- SigNoz integration
- Basic cart API endpoints
- Health check endpoint
- Request logging middleware

## Quick start

```bash
go mod download
go run main.go
```

The API runs on `http://localhost:8080`

## Endpoints

- `GET /api/v1/health` - Health check
- `GET /api/v1/error` - Intentional error endpoint for testing
- `POST /api/v1/cart` - Add item to cart

## Environment variables

```bash
LOG_LEVEL=INFO          # DEBUG, INFO, WARN, ERROR
PORT=8080               # Server port
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317  # SigNoz endpoint
```

## Testing

```bash
go test ./...
```

## SigNoz setup

1. Run SigNoz locally (Docker)
2. Import the dashboard: `signoz-dashboard.json`
3. View metrics like request count, duration, errors

## What you'll see

The app automatically tracks:
- HTTP request metrics (`fiber.shbm.http.requests.total.v2`)
- Request duration (`fiber.shbm.http.request.duration.seconds`)
- Error counts
- Active requests gauge
- Distributed traces

Built this as a reference for setting up observability in Go APIs.