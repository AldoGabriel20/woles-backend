package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterSubscriptionRoutes mounts all /api/v1/subscriptions routes.
func RegisterSubscriptionRoutes(router fiber.Router) {
	s := router.Group("/subscriptions")

	s.Post("/", handleCreateSubscription)
	s.Get("/", handleListSubscriptions)
	s.Get("/:id", handleGetSubscription)
	s.Patch("/:id", handleUpdateSubscription)
	s.Delete("/:id", handleDeleteSubscription)
	s.Post("/:id/archive", handleArchiveSubscription)
}

func handleCreateSubscription(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleListSubscriptions(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleGetSubscription(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleUpdateSubscription(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleDeleteSubscription(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleArchiveSubscription(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
