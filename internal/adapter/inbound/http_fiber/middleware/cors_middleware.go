package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const defaultAllowedOrigins = "https://woles.id,https://www.woles.id"

// CORSMiddleware sets CORS headers based on the CORS_ALLOWED_ORIGINS env var.
// Never sets a wildcard origin. Credentials are always allowed.
func CORSMiddleware() fiber.Handler {
	allowedOrigins := buildAllowedOriginsSet()

	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		if origin != "" && allowedOrigins[origin] {
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Access-Control-Allow-Credentials", "true")
			c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-CSRF-Token")
			c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Set("Access-Control-Max-Age", "86400")
			c.Set("Vary", "Origin")
		}

		// Respond immediately to preflight OPTIONS requests.
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

// buildAllowedOriginsSet parses CORS_ALLOWED_ORIGINS into a set for O(1) lookup.
func buildAllowedOriginsSet() map[string]bool {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = defaultAllowedOrigins
	}
	set := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			set[trimmed] = true
		}
	}
	return set
}
