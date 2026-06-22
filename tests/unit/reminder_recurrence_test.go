package unit_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/woles/woles-backend/internal/domain/reminder"
)

// jakartaLoc loads Asia/Jakarta. Tests that need it call this helper.
func jakartaLoc(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatalf("load Asia/Jakarta: %v", err)
	}
	return loc
}

func newReminder(rt reminder.RecurrenceType, nextRunAt time.Time) *reminder.Reminder {
	return &reminder.Reminder{
		ID:             "test-id",
		RecurrenceType: rt,
		NextRunAt:      nextRunAt,
	}
}

func newCustomReminder(intervalDays int, nextRunAt time.Time) *reminder.Reminder {
	rule, _ := json.Marshal(map[string]int{"interval_days": intervalDays})
	return &reminder.Reminder{
		ID:             "test-id",
		RecurrenceType: reminder.RecurrenceCustomInterval,
		RecurrenceRule: rule,
		NextRunAt:      nextRunAt,
	}
}

// ─── one_time ─────────────────────────────────────────────────────────────────

func TestCalculateNextRunAt_OneTime(t *testing.T) {
	base := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	r := newReminder(reminder.RecurrenceOneTime, base)
	got := r.CalculateNextRunAt(base)
	if !got.Equal(base) {
		t.Errorf("one_time: expected %v, got %v", base, got)
	}
}

// ─── daily ────────────────────────────────────────────────────────────────────

func TestCalculateNextRunAt_Daily(t *testing.T) {
	base := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	r := newReminder(reminder.RecurrenceDaily, base)
	got := r.CalculateNextRunAt(base)
	want := base.AddDate(0, 0, 1)
	if !got.Equal(want) {
		t.Errorf("daily: expected %v, got %v", want, got)
	}
}

// ─── weekly ───────────────────────────────────────────────────────────────────

func TestCalculateNextRunAt_Weekly(t *testing.T) {
	base := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	r := newReminder(reminder.RecurrenceWeekly, base)
	got := r.CalculateNextRunAt(base)
	want := base.AddDate(0, 0, 7)
	if !got.Equal(want) {
		t.Errorf("weekly: expected %v, got %v", want, got)
	}
}

// ─── monthly ──────────────────────────────────────────────────────────────────

func TestCalculateNextRunAt_Monthly(t *testing.T) {
	base := time.Date(2025, 1, 31, 9, 0, 0, 0, time.UTC)
	r := newReminder(reminder.RecurrenceMonthly, base)
	got := r.CalculateNextRunAt(base)
	// time.AddDate(0,1,0) on Jan 31 → March 3 (Go normalises overflow)
	want := base.AddDate(0, 1, 0)
	if !got.Equal(want) {
		t.Errorf("monthly (month-end): expected %v, got %v", want, got)
	}
}

func TestCalculateNextRunAt_Monthly_Normal(t *testing.T) {
	base := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	r := newReminder(reminder.RecurrenceMonthly, base)
	got := r.CalculateNextRunAt(base)
	want := time.Date(2025, 7, 15, 9, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("monthly (normal): expected %v, got %v", want, got)
	}
}

// ─── yearly ───────────────────────────────────────────────────────────────────

func TestCalculateNextRunAt_Yearly(t *testing.T) {
	base := time.Date(2024, 2, 29, 9, 0, 0, 0, time.UTC) // leap day
	r := newReminder(reminder.RecurrenceYearly, base)
	got := r.CalculateNextRunAt(base)
	// AddDate(1,0,0) on 2024-02-29 → 2025-03-01 (Go normalises overflow)
	want := base.AddDate(1, 0, 0)
	if !got.Equal(want) {
		t.Errorf("yearly (leap): expected %v, got %v", want, got)
	}
}

// ─── custom_interval ─────────────────────────────────────────────────────────

func TestCalculateNextRunAt_CustomInterval_14Days(t *testing.T) {
	base := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	r := newCustomReminder(14, base)
	got := r.CalculateNextRunAt(base)
	want := base.AddDate(0, 0, 14)
	if !got.Equal(want) {
		t.Errorf("custom 14d: expected %v, got %v", want, got)
	}
}

func TestCalculateNextRunAt_CustomInterval_InvalidFallback(t *testing.T) {
	// Missing interval_days → fall back to NextRunAt
	base := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	rule, _ := json.Marshal(map[string]string{"unexpected": "data"})
	r := &reminder.Reminder{
		ID:             "test",
		RecurrenceType: reminder.RecurrenceCustomInterval,
		RecurrenceRule: rule,
		NextRunAt:      base,
	}
	got := r.CalculateNextRunAt(base)
	if !got.Equal(base) {
		t.Errorf("custom fallback: expected %v, got %v", base, got)
	}
}

// ─── DST boundary (Asia/Jakarta, WIB UTC+7, no DST changes) ──────────────────

func TestCalculateNextRunAt_Daily_Jakarta(t *testing.T) {
	loc := jakartaLoc(t)
	base := time.Date(2025, 10, 26, 7, 0, 0, 0, loc) // midnight run time WIB
	r := newReminder(reminder.RecurrenceDaily, base)
	got := r.CalculateNextRunAt(base)
	want := base.AddDate(0, 0, 1)
	if !got.Equal(want) {
		t.Errorf("daily Jakarta: expected %v, got %v", want, got)
	}
}
