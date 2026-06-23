package http_fiber

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

// RegisterWebhookRoutes mounts the WhatsApp inbound webhook.
// This route must be registered WITHOUT CORS and WITHOUT CSRF middleware.
func RegisterWebhookRoutes(app *fiber.App, svc *Services) {
	app.Post("/webhooks/whatsapp/:provider", handleWhatsAppWebhook)
}

// handleWhatsAppWebhook handles POST /webhooks/whatsapp/:provider
//
// Responsibilities (per Section 8.3):
//  1. Verify HMAC-SHA256 provider signature.
//  2. Return HTTP 200 immediately after verification.
//  3. Publish whatsapp.message_received to RabbitMQ for async processing.
//
// No CORS headers are set on this route (server-to-server).
// No CSRF check is performed.
func handleWhatsAppWebhook(c *fiber.Ctx) error {
	signature := c.Get("X-Webhook-Signature")
	if signature == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing_signature",
		})
	}

	body := c.Body()
	if !verifyWebhookSignature(body, signature) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid_signature",
		})
	}

	// Acknowledge immediately; processing happens asynchronously via the message queue.
	// The publishing step will be wired in when the RabbitMQ publisher is injected.
	return c.SendStatus(fiber.StatusOK)
}

// verifyWebhookSignature validates an HMAC-SHA256 signature against the raw
// request body. The webhook secret is read from the WEBHOOK_SECRET env var.
func verifyWebhookSignature(body []byte, signature string) bool {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
	return hmac.Equal([]byte(expected), []byte(signature))
}
