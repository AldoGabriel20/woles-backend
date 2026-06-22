package middleware

import "github.com/gofiber/fiber/v2"

// planLevel maps plan names to a numeric rank for comparison.
// Higher rank = higher tier.
var planLevel = map[string]int{
	"free":     0,
	"premium":  1,
	"advanced": 2,
}

// PlanGateMiddleware rejects requests whose authenticated plan is below requiredPlan.
// requiredPlan must be "premium" or "advanced".
// Expects JWTMiddleware to have already set "plan" in Fiber locals.
func PlanGateMiddleware(requiredPlan string) fiber.Handler {
	required := planLevel[requiredPlan]

	return func(c *fiber.Ctx) error {
		currentPlan, _ := c.Locals("plan").(string)
		current := planLevel[currentPlan]

		if current < required {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":         "plan_required",
				"required_plan": requiredPlan,
				"upgrade_url":   "/billing/checkout",
			})
		}

		return c.Next()
	}
}
