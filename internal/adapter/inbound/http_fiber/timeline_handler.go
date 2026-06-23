package http_fiber

import (
	"time"

	"github.com/gofiber/fiber/v2"
	apptimeline "github.com/woles/woles-backend/internal/application/timeline"
)

type timelineHandler struct{ svc *apptimeline.Service }

// RegisterTimelineRoutes mounts all /api/v1/timeline routes.
func RegisterTimelineRoutes(router fiber.Router, svc *Services) {
	h := &timelineHandler{svc: svc.Timeline}
	router.Get("/timeline", h.get)
}

func (h *timelineHandler) get(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 50, 200)

	// Support ?range=30d or ?from=...&to=...
	if rangeStr := c.Query("range"); rangeStr != "" {
		result, err := h.svc.GetTimelineByRange(c.Context(), userID, rangeStr, page, perPage)
		if err != nil {
			return mapServiceError(c, err)
		}
		return c.JSON(fiber.Map{
			"items": result.Items,
			"meta":  paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
		})
	}

	fromStr := c.Query("from", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	toStr := c.Query("to", time.Now().Format("2006-01-02"))
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "from must be YYYY-MM-DD")
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "to must be YYYY-MM-DD")
	}
	result, err := h.svc.GetTimeline(c.Context(), userID, from, to, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{
		"items": result.Items,
		"meta":  paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}
