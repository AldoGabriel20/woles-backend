// Package payment defines outbound port interfaces for payment gateway providers.
package payment

import "context"

// PaymentEventType classifies the event received from a payment webhook.
type PaymentEventType string

const (
	PaymentEventSuccess PaymentEventType = "payment.success"
	PaymentEventFailed  PaymentEventType = "payment.failed"
	PaymentEventPending PaymentEventType = "payment.pending"
)

// PaymentEvent is the parsed payload from a payment webhook notification.
type PaymentEvent struct {
	Type      PaymentEventType
	Plan      string
	UserID    string
	OrderID   string
	AmountIDR float64
}

// PaymentProvider abstracts the payment gateway (e.g. Midtrans).
type PaymentProvider interface {
	// CreateCheckout creates a checkout session for the given plan and user.
	// Returns the URL the user should be redirected to.
	CreateCheckout(ctx context.Context, plan, userID string) (checkoutURL string, err error)

	// VerifyWebhook validates the provider signature and parses the event payload.
	VerifyWebhook(ctx context.Context, payload []byte, signature string) (*PaymentEvent, error)
}
