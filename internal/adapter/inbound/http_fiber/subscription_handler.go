package http_fiber

import (
	"time"

	"github.com/gofiber/fiber/v2"
	appsubscription "github.com/woles/woles-backend/internal/application/subscription"
	domainsubscription "github.com/woles/woles-backend/internal/domain/subscription"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

type subscriptionHandler struct{ svc *appsubscription.Service }

// RegisterSubscriptionRoutes mounts all /api/v1/subscriptions routes.
func RegisterSubscriptionRoutes(router fiber.Router, svc *Services) {
	h := &subscriptionHandler{svc: svc.Subscription}
	s := router.Group("/subscriptions")
	s.Post("/", h.create)
	s.Get("/", h.list)
	s.Get("/:id", h.get)
	s.Patch("/:id", h.update)
	s.Delete("/:id", h.delete)
	s.Post("/:id/archive", h.archive)
}

type createSubscriptionBody struct {
	Name          string  `json:"name"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	BillingCycle  string  `json:"billing_cycle"`
	NextBillingAt string  `json:"next_billing_at"`
	Category      string  `json:"category"`
}

func (h *subscriptionHandler) create(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body createSubscriptionBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Name == "" || body.BillingCycle == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "name and billing_cycle are required")
	}
	req := appsubscription.CreateSubscriptionRequest{
		Name:         body.Name,
		Amount:       body.Amount,
		Currency:     body.Currency,
		BillingCycle: domainsubscription.BillingCycle(body.BillingCycle),
		Category:     domainsubscription.SubscriptionCategory(body.Category),
	}
	if body.NextBillingAt != "" {
		t, err := time.Parse(time.RFC3339, body.NextBillingAt)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "next_billing_at must be RFC3339")
		}
		req.NextBillingAt = &t
	}
	sub, err := h.svc.CreateSubscription(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"subscription": sub})
}

func (h *subscriptionHandler) list(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 20, 100)
	filter := database.SubscriptionFilter{}
	if s := c.Query("status"); s != "" {
		st := domainsubscription.SubscriptionStatus(s)
		filter.Status = &st
	}
	if cat := c.Query("category"); cat != "" {
		ct := domainsubscription.SubscriptionCategory(cat)
		filter.Category = &ct
	}
	result, err := h.svc.GetSubscriptions(c.Context(), userID, filter, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{
		"subscriptions": result.Items,
		"meta":          paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}

func (h *subscriptionHandler) get(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	sub, err := h.svc.GetSubscriptionByID(c.Context(), userID, c.Params("id"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"subscription": sub})
}

type updateSubscriptionBody struct {
	Name          *string  `json:"name"`
	Amount        *float64 `json:"amount"`
	Currency      *string  `json:"currency"`
	BillingCycle  *string  `json:"billing_cycle"`
	NextBillingAt *string  `json:"next_billing_at"`
	Category      *string  `json:"category"`
}

func (h *subscriptionHandler) update(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body updateSubscriptionBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	req := appsubscription.UpdateSubscriptionRequest{
		Name:     body.Name,
		Amount:   body.Amount,
		Currency: body.Currency,
	}
	if body.BillingCycle != nil {
		bc := domainsubscription.BillingCycle(*body.BillingCycle)
		req.BillingCycle = &bc
	}
	if body.Category != nil {
		cat := domainsubscription.SubscriptionCategory(*body.Category)
		req.Category = &cat
	}
	if body.NextBillingAt != nil {
		t, err := time.Parse(time.RFC3339, *body.NextBillingAt)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "next_billing_at must be RFC3339")
		}
		req.NextBillingAt = &t
	}
	sub, err := h.svc.UpdateSubscription(c.Context(), userID, c.Params("id"), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"subscription": sub})
}

func (h *subscriptionHandler) delete(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.DeleteSubscription(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Subscription deleted"})
}

func (h *subscriptionHandler) archive(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.ArchiveSubscription(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Subscription archived"})
}
