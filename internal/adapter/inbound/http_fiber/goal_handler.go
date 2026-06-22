package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterGoalRoutes mounts all /api/v1/goals routes.
func RegisterGoalRoutes(router fiber.Router) {
	g := router.Group("/goals")

	// Static sub-paths must come before the /:id catch-all.
	g.Get("/history", handleGetGoalHistory)

	g.Post("/", handleCreateGoal)
	g.Get("/", handleListGoals)
	g.Get("/:id", handleGetGoal)
	g.Patch("/:id", handleUpdateGoal)
	g.Delete("/:id", handleDeleteGoal)
	g.Post("/:id/progress", handleUpdateGoalProgress)
}

func handleCreateGoal(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleListGoals(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleGetGoal(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleUpdateGoal(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleDeleteGoal(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleUpdateGoalProgress(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleGetGoalHistory(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
