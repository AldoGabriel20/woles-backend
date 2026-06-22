package unit_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

const testWebhookSecret = "test-webhook-secret-key"

// verifyWebhookSignature mirrors the webhook handler's HMAC-SHA256 check.
func verifyWebhookSignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func newWebhookApp() *fiber.App {
	app := fiber.New()
	app.Post("/webhooks/whatsapp/provider", func(c *fiber.Ctx) error {
		sig := c.Get("X-Webhook-Signature")
		body := c.Body()
		if !verifyWebhookSignature(body, sig, testWebhookSecret) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid_signature",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestWebhookSignature_ValidSignature_Returns200(t *testing.T) {
	app := newWebhookApp()
	payload := []byte(`{"event":"message_received"}`)

	mac := hmac.New(sha256.New, []byte(testWebhookSecret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest("POST", "/webhooks/whatsapp/provider", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", sig)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("valid signature: want 200, got %d", resp.StatusCode)
	}
}

func TestWebhookSignature_InvalidSignature_Returns401(t *testing.T) {
	app := newWebhookApp()
	payload := []byte(`{"event":"message_received"}`)

	req := httptest.NewRequest("POST", "/webhooks/whatsapp/provider", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "invalidsignature")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("invalid signature: want 401, got %d", resp.StatusCode)
	}
}

func TestWebhookSignature_MissingSignature_Returns401(t *testing.T) {
	app := newWebhookApp()
	payload := []byte(`{"event":"message_received"}`)

	req := httptest.NewRequest("POST", "/webhooks/whatsapp/provider", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	// No X-Webhook-Signature header.

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("missing signature: want 401, got %d", resp.StatusCode)
	}
}

func TestWebhookSignature_TamperedPayload_Returns401(t *testing.T) {
	app := newWebhookApp()
	original := []byte(`{"event":"message_received"}`)

	mac := hmac.New(sha256.New, []byte(testWebhookSecret))
	mac.Write(original)
	sig := hex.EncodeToString(mac.Sum(nil))

	// Send a different payload with the original signature.
	tampered := []byte(`{"event":"malicious_event"}`)
	req := httptest.NewRequest("POST", "/webhooks/whatsapp/provider", strings.NewReader(string(tampered)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", sig)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("tampered payload: want 401, got %d", resp.StatusCode)
	}
}
