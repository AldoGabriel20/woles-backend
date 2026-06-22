package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterReminderRoutes mounts all /api/v1/reminders routes.
func RegisterReminderRoutes(router fiber.Router) {
	r := router.Group("/reminders")

	r.Post("/", handleCreateReminder)
	r.Get("/", handleListReminders)
	r.Get("/:id", handleGetReminder)
	r.Patch("/:id", handleUpdateReminder)
	r.Delete("/:id", handleDeleteReminder)
	r.Post("/:id/pause", handlePauseReminder)
	r.Post("/:id/resume", handleResumeReminder)
	r.Post("/:id/complete", handleCompleteOccurrence)
}

func handleCreateReminder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleListReminders(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleGetReminder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleUpdateReminder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleDeleteReminder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handlePauseReminder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleResumeReminder(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleCompleteOccurrence(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
