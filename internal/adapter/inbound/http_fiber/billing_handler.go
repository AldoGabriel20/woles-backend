package http_fiber

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterBillingRoutes mounts all /api/v1/billing routes.
func RegisterBillingRoutes(router fiber.Router, svc *Services) {
	b := router.Group("/billing")
	b.Get("/plan", handleGetCurrentPlan)
}

// handleGetCurrentPlan handles GET /api/v1/billing/plan
// Returns the user's current plan from JWT locals.
func handleGetCurrentPlan(c *fiber.Ctx) error {
	plan, _ := c.Locals("plan").(string)
	if plan == "" {
		plan = "free"
	}
	return c.JSON(fiber.Map{"plan": plan})
}
