// Package http_fiber wires all route groups onto a Fiber application instance.
package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterRoutes mounts every route group under /api/v1 and registers the
// WhatsApp webhook at the app level (no /api/v1 prefix, no CORS/CSRF).
//
// Middleware (JWT, rate-limit, CSRF, CORS, security headers) must be applied
// to the relevant groups by the caller before invoking this function.
func RegisterRoutes(app *fiber.App) {
	v1 := app.Group("/api/v1")

	RegisterAuthRoutes(v1)
	RegisterReminderRoutes(v1)
	RegisterDocumentRoutes(v1)
	RegisterSubscriptionRoutes(v1)
	RegisterGoalRoutes(v1)
	RegisterFinanceRoutes(v1)
	RegisterTimelineRoutes(v1)
	RegisterNotificationRoutes(v1)
	RegisterFamilyRoutes(v1)
	RegisterChatRoutes(v1)
	RegisterAccountRoutes(v1)
	RegisterBillingRoutes(v1)

	// Webhook is registered directly on the app (no /api/v1 prefix).
	RegisterWebhookRoutes(app)
}
