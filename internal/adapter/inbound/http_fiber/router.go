// Package http_fiber wires all route groups onto a Fiber application instance.
package http_fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/woles/woles-backend/internal/adapter/inbound/http_fiber/middleware"
)

// RegisterRoutes mounts every route group under /api/v1 with JWT protection
// applied to all groups except public auth endpoints and webhooks.
func RegisterRoutes(app *fiber.App, svc *Services) {
	v1 := app.Group("/api/v1")

	// Public auth routes (no JWT).
	RegisterAuthRoutes(v1, svc)

	// Protected routes — require valid JWT.
	protected := v1.Use(middleware.JWTMiddleware())
	RegisterReminderRoutes(protected, svc)
	RegisterDocumentRoutes(protected, svc)
	RegisterSubscriptionRoutes(protected, svc)
	RegisterGoalRoutes(protected, svc)
	RegisterFinanceRoutes(protected, svc)
	RegisterTimelineRoutes(protected, svc)
	RegisterNotificationRoutes(protected, svc)
	RegisterFamilyRoutes(protected, svc)
	RegisterChatRoutes(protected, svc)
	RegisterAccountRoutes(protected, svc)
	RegisterBillingRoutes(protected, svc)

	// Webhook is registered directly on the app (no /api/v1 prefix, no JWT).
	RegisterWebhookRoutes(app, svc)
}
