package http_fiber

import (
	"github.com/gofiber/fiber/v2"
	appchat "github.com/woles/woles-backend/internal/application/chat"
)

type chatHandler struct{ svc *appchat.Service }

// RegisterChatRoutes mounts all /api/v1/chat routes.
func RegisterChatRoutes(router fiber.Router, svc *Services) {
	h := &chatHandler{svc: svc.Chat}
	ch := router.Group("/chat")
	ch.Get("/messages", h.messages)
	ch.Post("/messages", h.send)
	ch.Delete("/messages", h.deleteAll)
	ch.Get("/usage", h.usage)
	ch.Get("/intents", h.intents)
}

func (h *chatHandler) messages(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 30, 100)
	result, err := h.svc.GetMessages(c.Context(), userID, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{
		"messages": result.Items,
		"meta":     paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}

type sendMessageBody struct {
	Content string `json:"content"`
}

func (h *chatHandler) send(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body sendMessageBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Content == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "content is required")
	}
	result, err := h.svc.SendMessage(c.Context(), userID, body.Content)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"result": result})
}

func (h *chatHandler) deleteAll(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.DeleteAllMessages(c.Context(), userID); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "All messages deleted"})
}

func (h *chatHandler) usage(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	u, err := h.svc.GetUsage(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"usage": u})
}

func (h *chatHandler) intents(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	intents, err := h.svc.GetDetectedIntents(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"intents": intents})
}
