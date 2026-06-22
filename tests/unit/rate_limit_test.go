package unit_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ─── In-memory rate limiter (mirrors sliding-window semantics) ────────────────

type memRateLimiter struct {
	counts map[string][]time.Time
}

func newMemRateLimiter() *memRateLimiter {
	return &memRateLimiter{counts: map[string][]time.Time{}}
}

// allow records one request. Returns (allowed, remaining).
func (m *memRateLimiter) allow(_ context.Context, key string, limit int, window time.Duration) (bool, int) {
	now := time.Now()
	cutoff := now.Add(-window)
	var valid []time.Time
	for _, ts := range m.counts[key] {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	if len(valid) >= limit {
		m.counts[key] = valid
		return false, 0
	}
	valid = append(valid, now)
	m.counts[key] = valid
	remaining := limit - len(valid)
	return true, remaining
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestRateLimit_10AllowedThen11thBlocked(t *testing.T) {
	limiter := newMemRateLimiter()
	key := "user-123:/auth/login"
	limit := 10
	window := 15 * time.Minute

	for i := 1; i <= 10; i++ {
		allowed, _ := limiter.allow(context.Background(), key, limit, window)
		if !allowed {
			t.Errorf("request %d should be allowed (got blocked)", i)
		}
	}
	allowed, _ := limiter.allow(context.Background(), key, limit, window)
	if allowed {
		t.Error("11th request within window should be blocked")
	}
}

func TestRateLimit_HTTPMiddleware_Returns429OnExceeded(t *testing.T) {
	limiter := newMemRateLimiter()
	limit := 3
	window := time.Minute

	app := fiber.New()
	app.Post("/auth/login", func(c *fiber.Ctx) error {
		key := fmt.Sprintf("ip:%s:/auth/login", c.IP())
		allowed, remaining := limiter.allow(c.Context(), key, limit, window)
		if !allowed {
			c.Set("X-RateLimit-Remaining", "0")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate_limit_exceeded",
			})
		}
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		return c.SendStatus(fiber.StatusOK)
	})

	for i := 1; i <= limit; i++ {
		req := httptest.NewRequest("POST", "/auth/login", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("request %d: want 200, got %d", i, resp.StatusCode)
		}
	}
	// limit+1 th request
	req := httptest.NewRequest("POST", "/auth/login", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("request %d: want 429, got %d", limit+1, resp.StatusCode)
	}
}
