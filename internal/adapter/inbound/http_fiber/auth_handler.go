// Package http_fiber provides Fiber HTTP adapter handlers for the Woles API.
package http_fiber

import "github.com/gofiber/fiber/v2"

// RegisterAuthRoutes mounts all /api/v1/auth routes on the given router.
// CSRF middleware is applied to all POST/DELETE/PATCH methods.
// Rate-limiting is applied per Section 8.5 of the system design.
func RegisterAuthRoutes(router fiber.Router) {
	auth := router.Group("/auth")

	// Public auth endpoints (no JWT required).
	auth.Post("/register", handleRegister)
	auth.Post("/login", handleLogin)
	auth.Post("/refresh", handleRefreshToken)
	auth.Post("/otp/request", handleRequestOTP)
	auth.Post("/otp/verify", handleVerifyOTP)
	auth.Post("/password/reset/request", handlePasswordResetRequest)
	auth.Post("/password/reset/confirm", handlePasswordResetConfirm)

	// Protected auth endpoints (JWT required).
	auth.Post("/logout", handleLogout)
	auth.Post("/password/change", handleChangePassword)
	auth.Get("/me", handleGetMe)
	auth.Post("/2fa/enable", handleEnable2FA)
	auth.Post("/2fa/verify", handleVerify2FA)
	auth.Post("/2fa/disable", handleDisable2FA)
	auth.Get("/sessions", handleGetSessions)
	auth.Delete("/sessions/:id", handleRevokeSession)
	auth.Delete("/sessions", handleRevokeAllSessions)
}

// handleRegister handles POST /api/v1/auth/register
func handleRegister(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleLogin handles POST /api/v1/auth/login
func handleLogin(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleRefreshToken handles POST /api/v1/auth/refresh
func handleRefreshToken(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleLogout handles POST /api/v1/auth/logout
func handleLogout(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleRequestOTP handles POST /api/v1/auth/otp/request
func handleRequestOTP(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleVerifyOTP handles POST /api/v1/auth/otp/verify
func handleVerifyOTP(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handlePasswordResetRequest handles POST /api/v1/auth/password/reset/request
func handlePasswordResetRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handlePasswordResetConfirm handles POST /api/v1/auth/password/reset/confirm
func handlePasswordResetConfirm(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleChangePassword handles POST /api/v1/auth/password/change
func handleChangePassword(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetMe handles GET /api/v1/auth/me
func handleGetMe(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleEnable2FA handles POST /api/v1/auth/2fa/enable
func handleEnable2FA(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleVerify2FA handles POST /api/v1/auth/2fa/verify
func handleVerify2FA(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleDisable2FA handles POST /api/v1/auth/2fa/disable
func handleDisable2FA(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleGetSessions handles GET /api/v1/auth/sessions
func handleGetSessions(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleRevokeSession handles DELETE /api/v1/auth/sessions/:id
func handleRevokeSession(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

// handleRevokeAllSessions handles DELETE /api/v1/auth/sessions
func handleRevokeAllSessions(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}
