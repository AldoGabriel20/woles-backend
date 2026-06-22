package unit_test

import (
	"testing"
	"time"

	"github.com/woles/woles-backend/internal/domain/document"
)

func ptr[T any](v T) *T { return &v }

func docWithExpiry(t *testing.T, daysFromNow int) *document.Document {
	t.Helper()
	expiry := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, daysFromNow)
	return &document.Document{ExpiryDate: &expiry}
}

// ─── DaysUntilExpiry ─────────────────────────────────────────────────────────

func TestDaysUntilExpiry_NoExpiry(t *testing.T) {
	d := &document.Document{}
	got := d.DaysUntilExpiry()
	if got <= 0 {
		t.Errorf("no expiry: expected large positive, got %d", got)
	}
}

func TestDaysUntilExpiry_Future(t *testing.T) {
	d := docWithExpiry(t, 30)
	got := d.DaysUntilExpiry()
	// Allow ±1 for clock rounding
	if got < 29 || got > 31 {
		t.Errorf("30d future: expected ~30, got %d", got)
	}
}

func TestDaysUntilExpiry_Today(t *testing.T) {
	d := docWithExpiry(t, 0)
	got := d.DaysUntilExpiry()
	if got != 0 {
		t.Errorf("today: expected 0, got %d", got)
	}
}

func TestDaysUntilExpiry_Expired(t *testing.T) {
	d := docWithExpiry(t, -5)
	got := d.DaysUntilExpiry()
	if got >= 0 {
		t.Errorf("expired: expected negative, got %d", got)
	}
}

// ─── ExpiryRisk ──────────────────────────────────────────────────────────────

func TestExpiryRisk_NoExpiry(t *testing.T) {
	d := &document.Document{}
	if got := d.ExpiryRisk(); got != "safe" {
		t.Errorf("no expiry: expected safe, got %s", got)
	}
}

func TestExpiryRisk_Safe(t *testing.T) {
	d := docWithExpiry(t, 60)
	if got := d.ExpiryRisk(); got != "safe" {
		t.Errorf("60d: expected safe, got %s", got)
	}
}

func TestExpiryRisk_UpcomingBoundary30(t *testing.T) {
	d := docWithExpiry(t, 30)
	if got := d.ExpiryRisk(); got != "upcoming" {
		t.Errorf("30d: expected upcoming, got %s", got)
	}
}

func TestExpiryRisk_Upcoming(t *testing.T) {
	d := docWithExpiry(t, 15)
	if got := d.ExpiryRisk(); got != "upcoming" {
		t.Errorf("15d: expected upcoming, got %s", got)
	}
}

func TestExpiryRisk_UrgentBoundary7(t *testing.T) {
	d := docWithExpiry(t, 7)
	if got := d.ExpiryRisk(); got != "urgent" {
		t.Errorf("7d: expected urgent, got %s", got)
	}
}

func TestExpiryRisk_Urgent(t *testing.T) {
	d := docWithExpiry(t, 3)
	if got := d.ExpiryRisk(); got != "urgent" {
		t.Errorf("3d: expected urgent, got %s", got)
	}
}

func TestExpiryRisk_Today(t *testing.T) {
	d := docWithExpiry(t, 0)
	if got := d.ExpiryRisk(); got != "urgent" {
		t.Errorf("today: expected urgent, got %s", got)
	}
}

func TestExpiryRisk_Expired(t *testing.T) {
	d := docWithExpiry(t, -1)
	if got := d.ExpiryRisk(); got != "expired" {
		t.Errorf("-1d: expected expired, got %s", got)
	}
}

// ─── Notification offset calculation ─────────────────────────────────────────

// Demonstrates that offsets [30, 7, 1] produce notifications in the future
// for a document expiring 60 days from now.
func TestNotificationOffsets_FutureDoc(t *testing.T) {
	expiryDays := 60
	expiry := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, expiryDays)
	offsets := []int{30, 7, 1}
	for _, offset := range offsets {
		alertAt := expiry.AddDate(0, 0, -offset)
		if !alertAt.After(time.Now().UTC()) {
			t.Errorf("offset %d: alert %v should be in the future for 60-day expiry", offset, alertAt)
		}
	}
}

// For an already-expired document none of the alerts should be scheduled.
func TestNotificationOffsets_ExpiredDoc(t *testing.T) {
	expiry := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -5)
	offsets := []int{30, 7, 1}
	for _, offset := range offsets {
		alertAt := expiry.AddDate(0, 0, -offset)
		if alertAt.After(time.Now().UTC()) {
			t.Errorf("offset %d: alert %v should NOT be in the future for expired doc", offset, alertAt)
		}
	}
}
