// Package whatsapp defines outbound port interfaces for the WhatsApp provider.
package whatsapp

import "context"

// WhatsAppSender sends templated messages through a WhatsApp Business API provider.
type WhatsAppSender interface {
	// SendMessage sends a template message to the given phone number.
	// Returns the provider-assigned message ID on success.
	SendMessage(ctx context.Context, to, templateName string, params map[string]string) (providerMsgID string, err error)
}

// WhatsAppVerifier validates HMAC-SHA256 webhook signatures from the provider.
type WhatsAppVerifier interface {
	// VerifySignature returns true when the signature is valid for the given payload.
	VerifySignature(payload []byte, signature string) bool
}
