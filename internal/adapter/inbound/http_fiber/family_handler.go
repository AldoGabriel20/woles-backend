package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterFamilyRoutes mounts all /api/v1/family routes.
func RegisterFamilyRoutes(router fiber.Router) {
	f := router.Group("/family")

	f.Get("/members", handleListFamilyMembers)
	f.Post("/members", handleCreateFamilyMember)
	f.Get("/members/:id", handleGetFamilyMember)
	f.Patch("/members/:id", handleUpdateFamilyMember)
	f.Delete("/members/:id", handleDeleteFamilyMember)
	f.Get("/reminders", handleGetSharedReminders)
}

func handleListFamilyMembers(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleCreateFamilyMember(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleGetFamilyMember(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleUpdateFamilyMember(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleDeleteFamilyMember(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetSharedReminders handles GET /api/v1/family/reminders?page=1&per_page=20
func handleGetSharedReminders(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
