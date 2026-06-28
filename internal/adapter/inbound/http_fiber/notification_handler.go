package http_fiber

import (
	"time"

	"github.com/gofiber/fiber/v2"
	appnotification "github.com/woles/woles-backend/internal/application/notification"
	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

type notificationHandler struct{ svc *appnotification.Service }

// RegisterNotificationRoutes mounts all /api/v1/notifications routes.
func RegisterNotificationRoutes(router fiber.Router, svc *Services) {
	h := &notificationHandler{svc: svc.Notification}
	n := router.Group("/notifications")
	n.Get("/stats", h.stats)
	n.Get("/export", h.export)
	n.Get("/", h.list)
}

func (h *notificationHandler) list(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 20, 100)
	filter := database.NotificationFilter{}
	if et := c.Query("entity_type"); et != "" {
		t := domainnotification.NotificationEntityType(et)
		filter.EntityType = &t
	}
	if st := c.Query("status"); st != "" {
		s := domainnotification.NotificationStatus(st)
		filter.Status = &s
	}
	if from := c.Query("from"); from != "" {
		t, err := time.Parse("2006-01-02", from)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "from must be YYYY-MM-DD")
		}
		filter.From = &t
	}
	if to := c.Query("to"); to != "" {
		t, err := time.Parse("2006-01-02", to)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "to must be YYYY-MM-DD")
		}
		filter.To = &t
	}
	result, err := h.svc.GetNotifications(c.Context(), userID, filter, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{
		"notifications": result.Items,
		"meta":          paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}

func (h *notificationHandler) stats(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	stats, err := h.svc.GetStats(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"stats": stats})
}

func (h *notificationHandler) export(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	format := c.Query("format", "csv")
	rangeStr := c.Query("range", "30d")
	data, err := h.svc.ExportNotifications(c.Context(), userID, format, rangeStr)
	if err != nil {
		return mapServiceError(c, err)
	}
	switch format {
	case "pdf":
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", `attachment; filename="notifications.pdf"`)
		return c.Send(data)
	case "excel":
		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", `attachment; filename="notifications.xlsx"`)
		return c.Send(data)
	default: // csv
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", `attachment; filename="notifications.csv"`)
		return c.Send(data)
	}
}
