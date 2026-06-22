// Package timeline implements the Timeline application service.
package timeline

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var ErrInvalidRange = errors.New("invalid range string")

// ─── Type aliases ─────────────────────────────────────────────────────────────

// TimelineItem is the normalised timeline entry returned to callers.
type TimelineItem = database.TimelineItem

// ─── Service ──────────────────────────────────────────────────────────────────

// Service implements the timeline application service.
type Service struct {
	timeline database.TimelineRepository
}

// NewService constructs the timeline service.
func NewService(timeline database.TimelineRepository) *Service {
	return &Service{timeline: timeline}
}

// ─── GetTimeline ──────────────────────────────────────────────────────────────

// GetTimeline returns a paginated, chronologically sorted list of timeline
// items for the given user in the [from, to) window.
//
// The repository layer is responsible for querying reminder_occurrences,
// document expiry alert dates, subscriptions' next_billing_at, and goals'
// target_date, normalising them into TimelineItem rows and sorting by DueAt ASC.
func (s *Service) GetTimeline(ctx context.Context, userID string, from, to time.Time, page, perPage int) (*database.PaginatedResult[*TimelineItem], error) {
	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "due_at",
		Order:   "asc",
	}
	result, err := s.timeline.GetTimelineItems(ctx, userID, from, to, p)
	if err != nil {
		return nil, fmt.Errorf("get timeline: %w", err)
	}
	return result, nil
}

// ─── GetTimelineByRange ───────────────────────────────────────────────────────

// GetTimelineByRange parses a range string and delegates to GetTimeline.
//
// Supported range strings:
//   - "7d"         — next 7 days from now
//   - "30d"        — next 30 days from now
//   - "90d"        — next 90 days from now
//   - "this_month" — first day to last day of the current calendar month
//   - "next_month" — first day to last day of the next calendar month
func (s *Service) GetTimelineByRange(ctx context.Context, userID, rangeStr string, page, perPage int) (*database.PaginatedResult[*TimelineItem], error) {
	from, to, err := parseRange(rangeStr)
	if err != nil {
		return nil, err
	}
	return s.GetTimeline(ctx, userID, from, to, page, perPage)
}

// ─── Range parsing ────────────────────────────────────────────────────────────

// parseRange converts a range string into a [from, to) time window in UTC.
func parseRange(rangeStr string) (from, to time.Time, err error) {
	now := time.Now().UTC()

	switch rangeStr {
	case "7d":
		return now, now.AddDate(0, 0, 7), nil
	case "30d":
		return now, now.AddDate(0, 0, 30), nil
	case "90d":
		return now, now.AddDate(0, 0, 90), nil
	case "this_month":
		firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastDay := firstDay.AddDate(0, 1, 0) // exclusive upper bound (first of next month)
		return firstDay, lastDay, nil
	case "next_month":
		firstDayNext := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
		lastDayNext := firstDayNext.AddDate(0, 1, 0)
		return firstDayNext, lastDayNext, nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("%w: %q", ErrInvalidRange, rangeStr)
	}
}
