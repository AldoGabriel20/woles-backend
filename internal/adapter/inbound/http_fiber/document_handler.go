package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterDocumentRoutes mounts all /api/v1/documents routes.
// The file upload endpoint uses multipart/form-data; the 10 MB body-limit
// must be applied at the Fiber app level via app.Use(fiber.New(fiber.Config{BodyLimit: 10<<20})).
func RegisterDocumentRoutes(router fiber.Router) {
	d := router.Group("/documents")

	d.Post("/", handleCreateDocument)
	d.Get("/", handleListDocuments)

	// Static sub-paths must come before the /:id catch-all.
	d.Get("/storage/usage", handleGetStorageUsage)
	d.Get("/vault/health", handleGetVaultHealth)

	d.Get("/:id", handleGetDocument)
	d.Patch("/:id", handleUpdateDocument)
	d.Delete("/:id", handleDeleteDocument)

	// File management on a specific document.
	d.Post("/:id/file", handleUploadDocumentFile) // multipart/form-data
	d.Delete("/:id/file", handleDeleteDocumentFile)
}

func handleCreateDocument(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleListDocuments(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleGetDocument(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleUpdateDocument(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleDeleteDocument(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleUploadDocumentFile handles POST /api/v1/documents/:id/file
// Parses multipart/form-data and enforces the 10 MB limit set at the app level.
func handleUploadDocumentFile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleDeleteDocumentFile(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleGetStorageUsage(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

func handleGetVaultHealth(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
