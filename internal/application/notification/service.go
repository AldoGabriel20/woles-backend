// Package notification implements the Notification application service.
package notification

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"time"

	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var (
	ErrInvalidFormat = errors.New("unsupported export format: use \"csv\" or \"pdf\"")
	ErrInvalidRange  = errors.New("invalid range string")
)

// ─── Type aliases ─────────────────────────────────────────────────────────────

// NotificationStats is an alias for the database type.
type NotificationStats = database.NotificationStats

// ─── Service ──────────────────────────────────────────────────────────────────

// Service implements the notification application service.
type Service struct {
	notifications database.NotificationRepository
}

// NewService constructs the notification service.
func NewService(notifications database.NotificationRepository) *Service {
	return &Service{notifications: notifications}
}

// ─── GetNotifications ─────────────────────────────────────────────────────────

// GetNotifications returns a paginated list of notifications for a user,
// optionally filtered by entity_type, date range, and status.
func (s *Service) GetNotifications(ctx context.Context, userID string, filter database.NotificationFilter, page, perPage int) (*database.PaginatedResult[*domainnotification.Notification], error) {
	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "scheduled_at",
		Order:   "desc",
	}
	result, err := s.notifications.FindAllByUser(ctx, userID, filter, p)
	if err != nil {
		return nil, fmt.Errorf("get notifications: %w", err)
	}
	return result, nil
}

// ─── GetStats ─────────────────────────────────────────────────────────────────

// GetStats returns delivery statistics for a user's notifications.
func (s *Service) GetStats(ctx context.Context, userID string) (*NotificationStats, error) {
	stats, err := s.notifications.GetStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get notification stats: %w", err)
	}
	return stats, nil
}

// ─── ExportNotifications ─────────────────────────────────────────────────────

// ExportNotifications generates a CSV or PDF export of notifications for the
// given user and time range string. Supported formats: "csv", "pdf".
// Supported range strings: "7d", "30d", "90d", "this_month", "next_month",
// or "YYYY-MM" (calendar month).
func (s *Service) ExportNotifications(ctx context.Context, userID, format, rangeStr string) ([]byte, error) {
	if format != "csv" && format != "pdf" {
		return nil, ErrInvalidFormat
	}

	from, to, err := parseRange(rangeStr)
	if err != nil {
		return nil, err
	}

	items, err := s.notifications.ExportRange(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("export notifications: %w", err)
	}

	switch format {
	case "pdf":
		return buildPDF(items, from, to)
	default:
		return buildCSV(items)
	}
}

// ─── CSV builder ─────────────────────────────────────────────────────────────

func buildCSV(items []*domainnotification.Notification) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header row.
	if err := w.Write([]string{
		"ID", "EntityType", "EntityID", "Channel",
		"Status", "ScheduledAt", "SentAt", "RetryCount", "FailureReason",
	}); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, n := range items {
		sentAt := ""
		if n.SentAt != nil {
			sentAt = n.SentAt.UTC().Format(time.RFC3339)
		}
		failureReason := ""
		if n.FailureReason != nil {
			failureReason = *n.FailureReason
		}
		row := []string{
			n.ID,
			string(n.EntityType),
			n.EntityID,
			string(n.Channel),
			string(n.Status),
			n.ScheduledAt.UTC().Format(time.RFC3339),
			sentAt,
			strconv.Itoa(n.RetryCount),
			failureReason,
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}

// ─── PDF builder ──────────────────────────────────────────────────────────────

// buildPDF returns a minimal plain-text representation of the notification list
// formatted as if it were a PDF report. A full PDF implementation requires an
// external library (e.g. gofpdf); this scaffold produces a UTF-8 text document
// that can be replaced when the dependency is available.
func buildPDF(items []*domainnotification.Notification, from, to time.Time) ([]byte, error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Woles Notification Report\n")
	fmt.Fprintf(&buf, "Period: %s — %s\n", from.UTC().Format("2006-01-02"), to.UTC().Format("2006-01-02"))
	fmt.Fprintf(&buf, "Total: %d notification(s)\n\n", len(items))
	fmt.Fprintf(&buf, "%-36s  %-12s  %-12s  %-10s  %s\n",
		"ID", "EntityType", "Channel", "Status", "ScheduledAt")
	fmt.Fprintf(&buf, "%s\n", repeatChar('-', 100))

	for _, n := range items {
		fmt.Fprintf(&buf, "%-36s  %-12s  %-12s  %-10s  %s\n",
			n.ID,
			string(n.EntityType),
			string(n.Channel),
			string(n.Status),
			n.ScheduledAt.UTC().Format("2006-01-02 15:04"),
		)
	}

	return buf.Bytes(), nil
}

// ─── Range parsing ────────────────────────────────────────────────────────────

// parseRange converts a range string into a [from, to) time window in UTC.
// Accepts "7d", "30d", "90d", "this_month", "next_month", or "YYYY-MM".
func parseRange(rangeStr string) (from, to time.Time, err error) {
	now := time.Now().UTC()

	switch rangeStr {
	case "7d":
		return now.AddDate(0, 0, -7), now, nil
	case "30d":
		return now.AddDate(0, 0, -30), now, nil
	case "90d":
		return now.AddDate(0, 0, -90), now, nil
	case "this_month":
		first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return first, first.AddDate(0, 1, 0), nil
	case "next_month":
		first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
		return first, first.AddDate(0, 1, 0), nil
	}

	// Try YYYY-MM.
	t, parseErr := time.Parse("2006-01", rangeStr)
	if parseErr != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: %q", ErrInvalidRange, rangeStr)
	}
	first := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return first, first.AddDate(0, 1, 0), nil
}

func repeatChar(c rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}
