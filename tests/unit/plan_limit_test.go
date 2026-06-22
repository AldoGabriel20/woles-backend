package unit_test

import (
	"testing"

	"github.com/woles/woles-backend/internal/domain/billing"
)

func TestIsWithinLimit_Free_UnderLimit(t *testing.T) {
	p := &billing.Plan{ReminderLimit: 20, DocumentLimit: 5, SubscriptionLimit: 5}
	if !p.IsWithinLimit("reminders", 19) {
		t.Error("19/20 reminders: should be within limit")
	}
}

func TestIsWithinLimit_Free_AtLimit(t *testing.T) {
	p := &billing.Plan{ReminderLimit: 20}
	if p.IsWithinLimit("reminders", 20) {
		t.Error("20/20 reminders: should be at limit (not allowed)")
	}
}

func TestIsWithinLimit_Free_OverLimit(t *testing.T) {
	p := &billing.Plan{ReminderLimit: 20}
	if p.IsWithinLimit("reminders", 21) {
		t.Error("21/20 reminders: should be over limit")
	}
}

func TestIsWithinLimit_Premium_Unlimited(t *testing.T) {
	p := &billing.Plan{ReminderLimit: -1, DocumentLimit: -1, SubscriptionLimit: -1}
	if !p.IsWithinLimit("reminders", 9999) {
		t.Error("unlimited plan: should always be within limit")
	}
	if !p.IsWithinLimit("documents", 9999) {
		t.Error("unlimited plan: documents should always be within limit")
	}
}

func TestIsWithinLimit_Documents(t *testing.T) {
	p := &billing.Plan{DocumentLimit: 5}
	if !p.IsWithinLimit("documents", 4) {
		t.Error("4/5 docs: should be within limit")
	}
	if p.IsWithinLimit("documents", 5) {
		t.Error("5/5 docs: should be at limit")
	}
}

func TestIsWithinLimit_Subscriptions(t *testing.T) {
	p := &billing.Plan{SubscriptionLimit: 5}
	if !p.IsWithinLimit("subscriptions", 0) {
		t.Error("0/5: should be within limit")
	}
	if p.IsWithinLimit("subscriptions", 5) {
		t.Error("5/5: should be at limit")
	}
}

func TestIsWithinLimit_UnknownResource(t *testing.T) {
	p := &billing.Plan{ReminderLimit: 1}
	// Unknown resource defaults to allowed.
	if !p.IsWithinLimit("unknown_resource", 999) {
		t.Error("unknown resource: should default to allowed")
	}
}
