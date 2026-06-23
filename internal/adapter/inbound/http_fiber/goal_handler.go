package http_fiber

import (
	"time"

	"github.com/gofiber/fiber/v2"
	appgoal "github.com/woles/woles-backend/internal/application/goal"
	domaingoal "github.com/woles/woles-backend/internal/domain/goal"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

type goalHandler struct{ svc *appgoal.Service }

// RegisterGoalRoutes mounts all /api/v1/goals routes.
func RegisterGoalRoutes(router fiber.Router, svc *Services) {
	h := &goalHandler{svc: svc.Goal}
	g := router.Group("/goals")
	g.Get("/history", h.history)
	g.Post("/", h.create)
	g.Get("/", h.list)
	g.Get("/:id", h.get)
	g.Patch("/:id", h.update)
	g.Delete("/:id", h.delete)
	g.Post("/:id/progress", h.progress)
}

type createGoalBody struct {
	Title         string  `json:"title"`
	Icon          string  `json:"icon"`
	TargetAmount  float64 `json:"target_amount"`
	MonthlyTarget float64 `json:"monthly_target"`
	Currency      string  `json:"currency"`
	TargetDate    string  `json:"target_date"`
}

func (h *goalHandler) create(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body createGoalBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Title == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "title is required")
	}
	req := appgoal.CreateGoalRequest{
		Title:        body.Title,
		TargetAmount: body.TargetAmount,
		Currency:     body.Currency,
	}
	if body.Icon != "" {
		icon := domaingoal.GoalIcon(body.Icon)
		req.Icon = &icon
	}
	if body.MonthlyTarget > 0 {
		req.MonthlyTarget = &body.MonthlyTarget
	}
	if body.TargetDate != "" {
		t, err := time.Parse("2006-01-02", body.TargetDate)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "target_date must be YYYY-MM-DD")
		}
		req.TargetDate = &t
	}
	goal, err := h.svc.CreateGoal(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"goal": goal})
}

func (h *goalHandler) list(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 20, 100)
	filter := database.GoalFilter{}
	if s := c.Query("status"); s != "" {
		st := domaingoal.GoalStatus(s)
		filter.Status = &st
	}
	result, err := h.svc.GetGoals(c.Context(), userID, filter, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{
		"goals": result.Items,
		"meta":  paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}

func (h *goalHandler) get(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	goal, err := h.svc.GetGoalByID(c.Context(), userID, c.Params("id"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"goal": goal})
}

func (h *goalHandler) history(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 20, 100)
	result, err := h.svc.GetGoalHistory(c.Context(), userID, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{
		"goals": result.Items,
		"meta":  paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}

type updateGoalBody struct {
	Title         *string  `json:"title"`
	Icon          *string  `json:"icon"`
	TargetAmount  *float64 `json:"target_amount"`
	MonthlyTarget *float64 `json:"monthly_target"`
	Currency      *string  `json:"currency"`
	TargetDate    *string  `json:"target_date"`
}

func (h *goalHandler) update(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body updateGoalBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	req := appgoal.UpdateGoalRequest{
		Title:         body.Title,
		TargetAmount:  body.TargetAmount,
		MonthlyTarget: body.MonthlyTarget,
		Currency:      body.Currency,
	}
	if body.Icon != nil {
		icon := domaingoal.GoalIcon(*body.Icon)
		req.Icon = &icon
	}
	if body.TargetDate != nil {
		t, err := time.Parse("2006-01-02", *body.TargetDate)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "target_date must be YYYY-MM-DD")
		}
		req.TargetDate = &t
	}
	goal, err := h.svc.UpdateGoal(c.Context(), userID, c.Params("id"), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"goal": goal})
}

type progressBody struct {
	Amount float64 `json:"amount"`
}

func (h *goalHandler) progress(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body progressBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	goal, err := h.svc.UpdateProgress(c.Context(), userID, c.Params("id"), body.Amount)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"goal": goal})
}

func (h *goalHandler) delete(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.DeleteGoal(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Goal deleted"})
}
