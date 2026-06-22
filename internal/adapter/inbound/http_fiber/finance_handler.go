package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterFinanceRoutes mounts all /api/v1/finances routes.
func RegisterFinanceRoutes(router fiber.Router) {
	f := router.Group("/finances")

	f.Get("/summary", handleGetFinancialSummary)
	f.Get("/spending", handleGetSpendingByCategory)
	f.Get("/trend", handleGetSpendingTrend)
	f.Get("/upcoming-bills", handleGetUpcomingBills)
	f.Get("/export", handleExportFinances)
}

// handleGetFinancialSummary handles GET /api/v1/finances/summary?period=monthly
func handleGetFinancialSummary(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetSpendingByCategory handles GET /api/v1/finances/spending?period=monthly
func handleGetSpendingByCategory(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetSpendingTrend handles GET /api/v1/finances/trend?period=weekly
func handleGetSpendingTrend(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetUpcomingBills handles GET /api/v1/finances/upcoming-bills?page=1&per_page=20
func handleGetUpcomingBills(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleExportFinances handles GET /api/v1/finances/export?format=csv&period=monthly
func handleExportFinances(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
