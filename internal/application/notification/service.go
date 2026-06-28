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

	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"

	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var (
	ErrInvalidFormat = errors.New("unsupported export format: use \"csv\", \"excel\", or \"pdf\"")
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

// ExportNotifications generates a CSV, Excel, or PDF export of notifications for the
// given user and time range string. Supported formats: "csv", "excel", "pdf".
// Supported range strings: "7d", "30d", "90d", "this_month", "next_month",
// or "YYYY-MM" (calendar month).
func (s *Service) ExportNotifications(ctx context.Context, userID, format, rangeStr string) ([]byte, error) {
	if format != "csv" && format != "pdf" && format != "excel" {
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
	case "excel":
		return buildExcel(items, from, to)
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

func buildPDF(items []*domainnotification.Notification, from, to time.Time) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(0, 53, 39)
	pdf.CellFormat(0, 10, "Woles - Notification Report", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 6, fmt.Sprintf("Period: %s - %s", from.UTC().Format("2006-01-02"), to.UTC().Format("2006-01-02")), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Total: %d notification(s)", len(items)), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// Table header
	pdf.SetFillColor(0, 53, 39)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	cols := []struct {
		label string
		width float64
	}{
		{"Entity Type", 30},
		{"Channel", 26},
		{"Status", 26},
		{"Scheduled At", 42},
		{"Sent At", 42},
		{"Retries", 18},
		{"Failure Reason", 82},
	}
	for _, col := range cols {
		pdf.CellFormat(col.width, 8, col.label, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(30, 30, 30)
	for i, n := range items {
		fill := i%2 == 0
		if fill {
			pdf.SetFillColor(240, 248, 244)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		sentAt := ""
		if n.SentAt != nil {
			sentAt = n.SentAt.UTC().Format("2006-01-02 15:04")
		}
		failureReason := ""
		if n.FailureReason != nil {
			failureReason = *n.FailureReason
			if len(failureReason) > 40 {
				failureReason = failureReason[:40] + "..."
			}
		}
		row := []string{
			string(n.EntityType),
			string(n.Channel),
			string(n.Status),
			n.ScheduledAt.UTC().Format("2006-01-02 15:04"),
			sentAt,
			strconv.Itoa(n.RetryCount),
			failureReason,
		}
		for j, cell := range row {
			pdf.CellFormat(cols[j].width, 7, cell, "1", 0, "L", fill, 0, "")
		}
		pdf.Ln(-1)
	}

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 10, fmt.Sprintf("Generated by Woles on %s", time.Now().UTC().Format("2006-01-02 15:04 UTC")), "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// ─── Excel builder ────────────────────────────────────────────────────────────

func buildExcel(items []*domainnotification.Notification, from, to time.Time) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close() //nolint:errcheck

	sheet := "Notifications"
	f.SetSheetName("Sheet1", sheet)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"003527"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	evenStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F0F8F4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "003527"},
	})

	// Title
	f.MergeCell(sheet, "A1", "G1") //nolint:errcheck
	f.SetCellValue(sheet, "A1", "Woles - Notification Report")
	f.SetCellStyle(sheet, "A1", "A1", titleStyle) //nolint:errcheck
	f.MergeCell(sheet, "A2", "G2")                //nolint:errcheck
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Period: %s - %s  |  Total: %d notification(s)",
		from.UTC().Format("2006-01-02"), to.UTC().Format("2006-01-02"), len(items)))

	// Headers
	headers := []string{"Entity Type", "Channel", "Status", "Scheduled At", "Sent At", "Retries", "Failure Reason"}
	colWidths := []float64{16, 14, 14, 20, 20, 10, 35}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle) //nolint:errcheck
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, colWidths[i]) //nolint:errcheck
	}
	f.SetRowHeight(sheet, 4, 20) //nolint:errcheck

	// Data rows
	for i, n := range items {
		row := i + 5
		sentAt := ""
		if n.SentAt != nil {
			sentAt = n.SentAt.UTC().Format("2006-01-02 15:04")
		}
		failureReason := ""
		if n.FailureReason != nil {
			failureReason = *n.FailureReason
		}
		vals := []interface{}{
			string(n.EntityType),
			string(n.Channel),
			string(n.Status),
			n.ScheduledAt.UTC().Format("2006-01-02 15:04"),
			sentAt,
			n.RetryCount,
			failureReason,
		}
		for j, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
			if i%2 == 0 {
				f.SetCellStyle(sheet, cell, cell, evenStyle) //nolint:errcheck
			}
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("write excel: %w", err)
	}
	return buf.Bytes(), nil
}

// ─── Range parsing ────────────────────────────────────────────────────────────
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
