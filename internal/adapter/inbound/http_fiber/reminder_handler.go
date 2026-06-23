package http_fiber

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	appreminder "github.com/woles/woles-backend/internal/application/reminder"
	domainreminder "github.com/woles/woles-backend/internal/domain/reminder"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

type reminderHandler struct{ svc *appreminder.Service }

// RegisterReminderRoutes mounts all /api/v1/reminders routes.
func RegisterReminderRoutes(router fiber.Router, svc *Services) {
	h := &reminderHandler{svc: svc.Reminder}
	r := router.Group("/reminders")
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/:id", h.get)
	r.Patch("/:id", h.update)
	r.Delete("/:id", h.delete)
	r.Post("/:id/pause", h.pause)
	r.Post("/:id/resume", h.resume)
	r.Post("/:id/complete", h.complete)
}

type createReminderBody struct {
	Title          string `json:"title"`
	Category       string `json:"category"`
	RecurrenceType string `json:"recurrence_type"`
	RecurrenceRule []byte `json:"recurrence_rule"`
	NextRunAt      string `json:"next_run_at"`
	Timezone       string `json:"timezone"`
}

func (h *reminderHandler) create(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body createReminderBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Title == "" || body.RecurrenceType == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "title and recurrence_type are required")
	}
	var nextRunAt time.Time
	if body.NextRunAt != "" {
		t, err := time.Parse(time.RFC3339, body.NextRunAt)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "next_run_at must be RFC3339")
		}
		nextRunAt = t
	} else {
		nextRunAt = time.Now().Add(24 * time.Hour)
	}
	tz := body.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	req := appreminder.CreateReminderRequest{
		Title:          body.Title,
		Category:       domainreminder.ReminderCategory(body.Category),
		RecurrenceType: domainreminder.RecurrenceType(body.RecurrenceType),
		RecurrenceRule: body.RecurrenceRule,
		NextRunAt:      nextRunAt,
		Timezone:       tz,
		Source:         domainreminder.SourceWeb,
	}
	reminder, err := h.svc.CreateReminder(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"reminder": reminder})
}

func (h *reminderHandler) list(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 20, 100)

	filter := database.ReminderFilter{}
	if s := c.Query("status"); s != "" {
		st := domainreminder.ReminderStatus(s)
		filter.Status = &st
	}
	if cat := c.Query("category"); cat != "" {
		c2 := domainreminder.ReminderCategory(cat)
		filter.Category = &c2
	}

	allowedSort := map[string]bool{"next_run_at": true, "created_at": true}
	sort := c.Query("sort", "next_run_at")
	if !allowedSort[sort] {
		sort = "next_run_at"
	}
	order := c.Query("order", "asc")
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	result, err := h.svc.GetReminders(c.Context(), userID, filter, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	_ = sort // passed via PaginationParams in service
	return c.JSON(fiber.Map{
		"reminders": result.Items,
		"meta":      paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}

func (h *reminderHandler) get(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	reminder, err := h.svc.GetReminderByID(c.Context(), userID, c.Params("id"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"reminder": reminder})
}

type updateReminderBody struct {
	Title          *string `json:"title"`
	Category       *string `json:"category"`
	RecurrenceType *string `json:"recurrence_type"`
	RecurrenceRule []byte  `json:"recurrence_rule"`
	NextRunAt      *string `json:"next_run_at"`
	Timezone       *string `json:"timezone"`
}

func (h *reminderHandler) update(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body updateReminderBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	req := appreminder.UpdateReminderRequest{
		RecurrenceRule: body.RecurrenceRule,
	}
	if body.Title != nil {
		req.Title = body.Title
	}
	if body.Category != nil {
		cat := domainreminder.ReminderCategory(*body.Category)
		req.Category = &cat
	}
	if body.RecurrenceType != nil {
		rt := domainreminder.RecurrenceType(*body.RecurrenceType)
		req.RecurrenceType = &rt
	}
	if body.NextRunAt != nil {
		t, err := time.Parse(time.RFC3339, *body.NextRunAt)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "next_run_at must be RFC3339")
		}
		req.NextRunAt = &t
	}
	if body.Timezone != nil {
		req.Timezone = body.Timezone
	}
	reminder, err := h.svc.UpdateReminder(c.Context(), userID, c.Params("id"), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"reminder": reminder})
}

func (h *reminderHandler) delete(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.DeleteReminder(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Reminder deleted"})
}

func (h *reminderHandler) pause(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.PauseReminder(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Reminder paused"})
}

func (h *reminderHandler) resume(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.ResumeReminder(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Reminder resumed"})
}

func (h *reminderHandler) complete(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.CompleteOccurrence(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Occurrence completed"})
}

// Ensure json import is used.
var _ = json.Marshal
