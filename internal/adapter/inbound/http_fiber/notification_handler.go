package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterNotificationRoutes mounts all /api/v1/notifications routes.
func RegisterNotificationRoutes(router fiber.Router) {
	n := router.Group("/notifications")

	// Static sub-paths must come before the /:id catch-all.
	n.Get("/stats", handleGetNotificationStats)
	n.Get("/export", handleExportNotifications)

	n.Get("/", handleListNotifications)
	n.Get("/:id", handleGetNotification)
}

// handleListNotifications handles GET /api/v1/notifications?page=1&per_page=20&category=vehicle&range=30d&status=sent
func handleListNotifications(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetNotification handles GET /api/v1/notifications/:id
func handleGetNotification(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetNotificationStats handles GET /api/v1/notifications/stats
func handleGetNotificationStats(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleExportNotifications handles GET /api/v1/notifications/export?format=pdf&range=30d
func handleExportNotifications(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
