package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterTimelineRoutes mounts all /api/v1/timeline routes.
func RegisterTimelineRoutes(router fiber.Router) {
	router.Get("/timeline", handleGetTimeline)
}

// handleGetTimeline handles:
//
//	GET /api/v1/timeline?from=2026-09-01&to=2026-09-30&page=1&per_page=50
//	GET /api/v1/timeline?range=90d&page=1&per_page=50
func handleGetTimeline(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
