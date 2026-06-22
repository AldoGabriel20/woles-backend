package unit_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/woles/woles-backend/internal/adapter/inbound/http_fiber/middleware"
)

func newFiberApp() *fiber.App {
	return fiber.New(fiber.Config{
		// Disable the default error handler to avoid noise in tests.
	})
}

// ─── CSRF missing header → 403 ───────────────────────────────────────────────

func TestCSRF_MissingHeader_Returns403(t *testing.T) {
	app := newFiberApp()
	app.Use(middleware.CSRFMiddleware())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	// Provide the cookie but omit the header.
	req.Header.Set("Cookie", "csrf_token=some-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("missing X-CSRF-Token header: want 403, got %d", resp.StatusCode)
	}
}

// ─── CSRF mismatched header → 403 ────────────────────────────────────────────

func TestCSRF_MismatchedHeader_Returns403(t *testing.T) {
	app := newFiberApp()
	app.Use(middleware.CSRFMiddleware())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Cookie", "csrf_token=cookie-token")
	req.Header.Set("X-CSRF-Token", "different-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("mismatched tokens: want 403, got %d", resp.StatusCode)
	}
}

// ─── CSRF matching tokens → 200 ──────────────────────────────────────────────

func TestCSRF_MatchingTokens_Returns200(t *testing.T) {
	app := newFiberApp()
	app.Use(middleware.CSRFMiddleware())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Cookie", "csrf_token=matching-token")
	req.Header.Set("X-CSRF-Token", "matching-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("matching tokens: want 200, got %d", resp.StatusCode)
	}
}

// ─── CSRF skipped for /webhooks/ paths ───────────────────────────────────────

func TestCSRF_WebhookPath_Skipped(t *testing.T) {
	app := newFiberApp()
	app.Use(middleware.CSRFMiddleware())
	app.Post("/webhooks/whatsapp/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/webhooks/whatsapp/test", nil)
	// No CSRF header or cookie.

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("webhook path: CSRF should be skipped, want 200, got %d", resp.StatusCode)
	}
}

// ─── CSRF GET request issues cookie ──────────────────────────────────────────

func TestCSRF_GetRequest_IssuesCookie(t *testing.T) {
	app := newFiberApp()
	app.Use(middleware.CSRFMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	found := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrf_token" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GET request should issue csrf_token cookie")
	}
}
