package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

// CSRFMiddleware implements double-submit CSRF protection.
// GET requests: issue a new CSRF token cookie (SameSite=Strict, HttpOnly=false).
// POST/PATCH/DELETE requests: verify X-CSRF-Token header matches csrf_token cookie.
// Skipped for paths starting with /webhooks/.
func CSRFMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip CSRF check for webhook paths (server-to-server).
		if strings.HasPrefix(c.Path(), "/webhooks/") {
			return c.Next()
		}

		method := c.Method()

		if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
			// Issue a new CSRF token on safe methods.
			token, err := generateCSRFToken()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "internal_error",
					"message": "Failed to generate CSRF token",
				})
			}
			c.Cookie(&fiber.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				MaxAge:   86400,
				Secure:   true,
				HTTPOnly: false,
				SameSite: "Strict",
			})
			return c.Next()
		}

		// For state-mutating methods: validate the CSRF token.
		headerToken := c.Get(csrfHeaderName)
		cookieToken := c.Cookies(csrfCookieName)

		if headerToken == "" || cookieToken == "" {
			return csrfForbidden(c)
		}

		// Constant-time comparison to prevent timing attacks.
		if subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookieToken)) != 1 {
			return csrfForbidden(c)
		}

		return c.Next()
	}
}

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func csrfForbidden(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"error":   "csrf_invalid",
		"message": "CSRF token mismatch or missing",
	})
}
