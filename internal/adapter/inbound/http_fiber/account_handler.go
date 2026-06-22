package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterAccountRoutes mounts all /api/v1/account routes.
func RegisterAccountRoutes(router fiber.Router) {
	a := router.Group("/account")

	a.Get("/profile", handleGetProfile)
	a.Patch("/profile", handleUpdateProfile)
	a.Post("/avatar", handleUploadAvatar)
	a.Get("/export", handleExportAccountData)
	a.Delete("/", handleDeleteAccount)
}

// handleGetProfile handles GET /api/v1/account/profile
func handleGetProfile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleUpdateProfile handles PATCH /api/v1/account/profile
func handleUpdateProfile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleUploadAvatar handles POST /api/v1/account/avatar
func handleUploadAvatar(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleExportAccountData handles GET /api/v1/account/export
func handleExportAccountData(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleDeleteAccount handles DELETE /api/v1/account
func handleDeleteAccount(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
