package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

// CSRFMiddleware implements double-submit CSRF protection for all environments.
// GET/HEAD/OPTIONS: issue a new CSRF token cookie.
// POST/PATCH/DELETE: verify X-CSRF-Token header matches csrf_token cookie.
// Skip for /webhooks/ paths (server-to-server).
//
// Cookie Secure flag: true in production (APP_ENV != "development") so it
// works over plain HTTP on localhost.
func CSRFMiddleware() fiber.Handler {
	isProduction := os.Getenv("APP_ENV") != "development"

	return func(c *fiber.Ctx) error {
		// Skip CSRF for webhook paths (server-to-server calls).
		if strings.HasPrefix(c.Path(), "/webhooks/") {
			return c.Next()
		}

		method := c.Method()

		if method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions {
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
				Secure:   isProduction,
				HTTPOnly: false,
				SameSite: "Strict",
			})
			return c.Next()
		}

		// Validate CSRF token on mutating methods.
		headerToken := c.Get(csrfHeaderName)
		cookieToken := c.Cookies(csrfCookieName)

		if headerToken == "" || cookieToken == "" {
			return csrfForbidden(c)
		}

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
