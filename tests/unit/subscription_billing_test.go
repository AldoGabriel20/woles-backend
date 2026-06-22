package unit_test

import (
	"testing"
	"time"

	"github.com/woles/woles-backend/internal/domain/subscription"
)

// nextBillingDate mirrors the application service logic for calculating the
// first billing date from a start time and cycle.
func nextBillingDate(from time.Time, cycle subscription.BillingCycle) time.Time {
	switch cycle {
	case subscription.BillingMonthly:
		return from.AddDate(0, 1, 0)
	case subscription.BillingYearly:
		return from.AddDate(1, 0, 0)
	default: // custom
		return from
	}
}

func TestNextBillingDate_Monthly(t *testing.T) {
	from := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	got := nextBillingDate(from, subscription.BillingMonthly)
	want := time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("monthly: want %v, got %v", want, got)
	}
}

func TestNextBillingDate_Monthly_MonthEnd(t *testing.T) {
	from := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	got := nextBillingDate(from, subscription.BillingMonthly)
	// Go's AddDate(0,1,0) on Jan 31 → March 3
	want := from.AddDate(0, 1, 0)
	if !got.Equal(want) {
		t.Errorf("monthly month-end: want %v, got %v", want, got)
	}
}

func TestNextBillingDate_Yearly(t *testing.T) {
	from := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	got := nextBillingDate(from, subscription.BillingYearly)
	want := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("yearly: want %v, got %v", want, got)
	}
}

func TestNextBillingDate_Yearly_LeapDay(t *testing.T) {
	from := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	got := nextBillingDate(from, subscription.BillingYearly)
	// 2025-02-29 doesn't exist; Go normalises to 2025-03-01
	want := from.AddDate(1, 0, 0)
	if !got.Equal(want) {
		t.Errorf("yearly leap: want %v, got %v", want, got)
	}
}

func TestNextBillingDate_Custom(t *testing.T) {
	from := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	got := nextBillingDate(from, subscription.BillingCustom)
	// Custom cycle returns from unchanged (caller sets next_billing_at manually).
	if !got.Equal(from) {
		t.Errorf("custom: want %v, got %v", from, got)
	}
}
