package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterBillingRoutes mounts all /api/v1/billing routes.
func RegisterBillingRoutes(router fiber.Router) {
	b := router.Group("/billing")

	b.Get("/plan", handleGetCurrentPlan)
	b.Post("/checkout", handleCreateCheckout)
	b.Post("/webhook", handlePaymentWebhook)
}

// handleGetCurrentPlan handles GET /api/v1/billing/plan
func handleGetCurrentPlan(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleCreateCheckout handles POST /api/v1/billing/checkout
func handleCreateCheckout(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handlePaymentWebhook handles POST /api/v1/billing/webhook
// This is an internal-facing payment provider webhook (not the WhatsApp webhook).
func handlePaymentWebhook(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
