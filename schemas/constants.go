package schemas

// Metric name constants
const (
	HTTPRequestsTotal          = "fiber.shbm.http.requests.total.v2"
	HTTPRequestDurationSeconds = "fiber.shbm.http.request.duration.seconds"
	HTTPActiveRequests         = "fiber.shbm.http.active.requests"
	ErrorsTotal                = "fiber.shbm.errors.total"
	CartCurrentItems           = "fiber.shbm.cart.current.items"
	CartCurrentValue           = "fiber.shbm.cart.current.value"
	CartRequestsTotal          = "fiber.shbm.cart.requests.total"
	CartOperationsTotal        = "fiber.shbm.cart.operations.total"
	CartItemsTotal             = "fiber.shbm.cart.items.total"
	CartItemsPerRequest        = "fiber.shbm.cart.items.per.request"
	HealthChecksTotal          = "fiber.shbm.health.checks.total"
	IntentionalErrorsTotal     = "fiber.shbm.intentional.errors.total"
)
