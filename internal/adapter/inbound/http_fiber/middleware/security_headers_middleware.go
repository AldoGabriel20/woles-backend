package middleware

import "github.com/gofiber/fiber/v2"

// SecurityHeadersMiddleware sets security-related HTTP response headers on every response.
func SecurityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; object-src 'none'; frame-ancestors 'none'")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		return c.Next()
	}
}
