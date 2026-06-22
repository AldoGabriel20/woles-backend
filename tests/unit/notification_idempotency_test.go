package unit_test

import (
	"strings"
	"testing"

	"github.com/woles/woles-backend/internal/domain/notification"
)

func TestBuildIdempotencyKey_Format(t *testing.T) {
	n := &notification.Notification{
		ID:      "abc-123",
		Channel: notification.ChannelWhatsApp,
	}
	got := n.BuildIdempotencyKey()
	want := "notification:abc-123:channel:whatsapp"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestBuildIdempotencyKey_Email(t *testing.T) {
	n := &notification.Notification{
		ID:      "xyz-456",
		Channel: notification.ChannelEmail,
	}
	got := n.BuildIdempotencyKey()
	if !strings.Contains(got, "channel:email") {
		t.Errorf("expected channel:email in key, got %s", got)
	}
}

func TestBuildIdempotencyKey_Uniqueness(t *testing.T) {
	n1 := &notification.Notification{ID: "id-1", Channel: notification.ChannelWhatsApp}
	n2 := &notification.Notification{ID: "id-2", Channel: notification.ChannelWhatsApp}
	k1 := n1.BuildIdempotencyKey()
	k2 := n2.BuildIdempotencyKey()
	if k1 == k2 {
		t.Errorf("different IDs should produce different keys: %s == %s", k1, k2)
	}
}

func TestBuildIdempotencyKey_SameIDDifferentChannel_Unique(t *testing.T) {
	n1 := &notification.Notification{ID: "same-id", Channel: notification.ChannelWhatsApp}
	n2 := &notification.Notification{ID: "same-id", Channel: notification.ChannelEmail}
	k1 := n1.BuildIdempotencyKey()
	k2 := n2.BuildIdempotencyKey()
	if k1 == k2 {
		t.Errorf("same ID + different channel should produce different keys: %s == %s", k1, k2)
	}
}
