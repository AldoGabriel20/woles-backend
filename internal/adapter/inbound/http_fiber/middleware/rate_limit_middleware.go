package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	portcache "github.com/woles/woles-backend/internal/port/outbound/cache"
)

// RateLimitMiddleware enforces a sliding-window rate limit using Redis.
// identifier: "userID" for authenticated routes, or IP for unauthenticated ones.
// Set identifierKey to "userID" for authenticated routes; leave empty to use IP.
func RateLimitMiddleware(limiter portcache.RateLimiter, limit int, window time.Duration, identifierKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var identifier string
		if identifierKey != "" {
			if v, ok := c.Locals(identifierKey).(string); ok && v != "" {
				identifier = v
			}
		}
		if identifier == "" {
			identifier = c.IP()
		}

		endpointHash := endpointFingerprint(c.Method(), c.Path())
		key := fmt.Sprintf("ratelimit:%s:%s", identifier, endpointHash)

		allowed, remaining, err := limiter.Allow(context.Background(), key, limit, window)
		resetAt := time.Now().Add(window).Unix()

		// Always set rate limit headers.
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt))

		if err != nil {
			// On Redis error, fail open (let the request through).
			return c.Next()
		}

		if !allowed {
			retryAfter := int64(window.Seconds())
			c.Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please slow down.",
			})
		}

		return c.Next()
	}
}

// endpointFingerprint returns a short deterministic hash of method+path.
func endpointFingerprint(method, path string) string {
	sum := sha256.Sum256([]byte(method + ":" + path))
	return fmt.Sprintf("%x", sum[:8])
}
