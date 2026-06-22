package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterChatRoutes mounts all /api/v1/chat routes.
func RegisterChatRoutes(router fiber.Router) {
	ch := router.Group("/chat")

	ch.Get("/messages", handleGetChatMessages)
	ch.Post("/messages", handleSendChatMessage)
	ch.Delete("/messages", handleDeleteAllChatMessages)
	ch.Get("/usage", handleGetChatUsage)
	ch.Get("/intents", handleGetDetectedIntents)
}

// handleGetChatMessages handles GET /api/v1/chat/messages?page=1&per_page=30
func handleGetChatMessages(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleSendChatMessage handles POST /api/v1/chat/messages
func handleSendChatMessage(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleDeleteAllChatMessages handles DELETE /api/v1/chat/messages
func handleDeleteAllChatMessages(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetChatUsage handles GET /api/v1/chat/usage
func handleGetChatUsage(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetDetectedIntents handles GET /api/v1/chat/intents
func handleGetDetectedIntents(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
