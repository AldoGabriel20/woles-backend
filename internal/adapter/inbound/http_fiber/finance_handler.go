package http_fiber

import (
	"github.com/gofiber/fiber/v2"
	appsubscription "github.com/woles/woles-backend/internal/application/subscription"
)

type financeHandler struct{ svc *appsubscription.Service }

// RegisterFinanceRoutes mounts all /api/v1/finances routes.
func RegisterFinanceRoutes(router fiber.Router, svc *Services) {
	h := &financeHandler{svc: svc.Subscription}
	f := router.Group("/finances")
	f.Get("/summary", h.summary)
	f.Get("/monthly-costs", h.monthlyCosts)
}

func (h *financeHandler) summary(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	costs, err := h.svc.GetMonthlyCostSummary(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"summary": costs})
}

func (h *financeHandler) monthlyCosts(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	costs, err := h.svc.GetMonthlyCostSummary(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"monthly_costs": costs})
}
