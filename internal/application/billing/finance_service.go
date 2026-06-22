// Package billing includes the financial overview service for the dashboard.
package billing

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	domainsubscription "github.com/woles/woles-backend/internal/domain/subscription"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrInvalidPeriod     = errors.New("invalid period: use \"monthly\" or \"YYYY-MM\"")
	ErrUnsupportedFormat = errors.New("unsupported format: use \"csv\"")
)

// ─── Response types ───────────────────────────────────────────────────────────

// FinancialSummary is the top-level financial overview for a period.
type FinancialSummary struct {
	Period                string  `json:"period"`
	TotalExpenses         float64 `json:"total_expenses"`
	Income                float64 `json:"income"`
	Savings               float64 `json:"savings"`
	ChangeVsLastPeriodPct float64 `json:"change_vs_last_period_pct"`
}

// SpendingCategory is a single category row inside SpendingBreakdown.
type SpendingCategory struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
	SubCount   int     `json:"sub_count"`
}

// SpendingBreakdown groups subscription spending by display category.
type SpendingBreakdown struct {
	Period     string              `json:"period"`
	Total      float64             `json:"total"`
	Categories []*SpendingCategory `json:"categories"`
}

// WeeklyAmount represents spending for one week inside SpendingTrend.
type WeeklyAmount struct {
	Week   string    `json:"week"`
	From   time.Time `json:"from"`
	To     time.Time `json:"to"`
	Amount float64   `json:"amount"`
}

// SpendingTrend holds per-week spending totals for a given period.
type SpendingTrend struct {
	Period      string          `json:"period"`
	Granularity string          `json:"granularity"`
	Weeks       []*WeeklyAmount `json:"weeks"`
}

// UpcomingBill is an active subscription due within the next 30 days.
type UpcomingBill struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	NextBillingAt time.Time `json:"next_billing_at"`
	Category      string    `json:"category"`
	UrgencyStatus string    `json:"urgency_status"` // URGENT, PENDING, SCHEDULED
}

// ─── FinancialService ─────────────────────────────────────────────────────────

// FinancialService implements the financial overview application service.
type FinancialService struct {
	subscriptions database.SubscriptionRepository
}

// NewFinancialService constructs the financial service.
func NewFinancialService(subscriptions database.SubscriptionRepository) *FinancialService {
	return &FinancialService{subscriptions: subscriptions}
}

// ─── GetFinancialSummary ──────────────────────────────────────────────────────

// GetFinancialSummary returns a period-level financial overview.
// For MVP, income and savings are 0 (bank sync is V2). Expenses are derived
// from active subscriptions whose next_billing_at falls in the period.
func (s *FinancialService) GetFinancialSummary(ctx context.Context, userID, period string) (*FinancialSummary, error) {
	from, to, label, err := parsePeriod(period)
	if err != nil {
		return nil, err
	}

	subs, err := s.fetchAllActive(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get financial summary: %w", err)
	}

	totalExpenses := sumInPeriod(subs, from, to)

	prevFrom, prevTo := prevPeriod(from)
	prevTotal := sumInPeriod(subs, prevFrom, prevTo)

	changePct := 0.0
	if prevTotal > 0 {
		changePct = math.Round(((totalExpenses-prevTotal)/prevTotal*100)*100) / 100
	}

	return &FinancialSummary{
		Period:                label,
		TotalExpenses:         totalExpenses,
		Income:                0, // V2: bank sync
		Savings:               0,
		ChangeVsLastPeriodPct: changePct,
	}, nil
}

// ─── GetSpendingByCategory ────────────────────────────────────────────────────

// GetSpendingByCategory groups active subscriptions by display category for the
// given period and returns the amount and percentage for each group.
//
// Category mapping:
//
//	entertainment → Household
//	productivity, bill → Utilities
//	other → Others
func (s *FinancialService) GetSpendingByCategory(ctx context.Context, userID, period string) (*SpendingBreakdown, error) {
	from, to, label, err := parsePeriod(period)
	if err != nil {
		return nil, err
	}

	subs, err := s.fetchAllActive(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get spending by category: %w", err)
	}

	totals := map[string]float64{}
	counts := map[string]int{}

	for _, sub := range subs {
		if !inPeriod(sub.NextBillingAt, from, to) {
			continue
		}
		cat := mapDisplayCategory(sub.Category)
		totals[cat] += sub.Amount
		counts[cat]++
	}

	var grandTotal float64
	for _, v := range totals {
		grandTotal += v
	}

	cats := make([]*SpendingCategory, 0, len(totals))
	for cat, amount := range totals {
		pct := 0.0
		if grandTotal > 0 {
			pct = math.Round(amount/grandTotal*100*100) / 100
		}
		cats = append(cats, &SpendingCategory{
			Category:   cat,
			Amount:     amount,
			Percentage: pct,
			SubCount:   counts[cat],
		})
	}

	sortCategories(cats)

	return &SpendingBreakdown{
		Period:     label,
		Total:      grandTotal,
		Categories: cats,
	}, nil
}

// ─── GetSpendingTrend ─────────────────────────────────────────────────────────

// GetSpendingTrend returns per-week spending totals for the given period.
// Each week spans exactly 7 days; the last week covers the remainder of the
// month. Amounts are derived from subscriptions whose next_billing_at falls
// within each week.
func (s *FinancialService) GetSpendingTrend(ctx context.Context, userID, period string) (*SpendingTrend, error) {
	from, to, label, err := parsePeriod(period)
	if err != nil {
		return nil, err
	}

	subs, err := s.fetchAllActive(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get spending trend: %w", err)
	}

	weeks := buildWeeks(from, to)

	for _, sub := range subs {
		for _, w := range weeks {
			if inPeriod(sub.NextBillingAt, w.From, w.To) {
				w.Amount += sub.Amount
				break
			}
		}
	}

	return &SpendingTrend{
		Period:      label,
		Granularity: "weekly",
		Weeks:       weeks,
	}, nil
}

// ─── GetUpcomingBills ─────────────────────────────────────────────────────────

// GetUpcomingBills returns active subscriptions due within the next 30 days,
// sorted by next_billing_at ASC, with a computed urgency label:
//
//	URGENT   — due within 3 days
//	PENDING  — due within 7 days
//	SCHEDULED — due later
func (s *FinancialService) GetUpcomingBills(ctx context.Context, userID string, page, perPage int) (*database.PaginatedResult[*UpcomingBill], error) {
	subs, err := s.fetchAllActive(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get upcoming bills: %w", err)
	}

	cutoff := time.Now().UTC().AddDate(0, 0, 30)
	var upcoming []*UpcomingBill

	for _, sub := range subs {
		if sub.NextBillingAt.After(cutoff) {
			continue
		}
		upcoming = append(upcoming, &UpcomingBill{
			ID:            sub.ID,
			Name:          sub.Name,
			Amount:        sub.Amount,
			Currency:      sub.Currency,
			NextBillingAt: sub.NextBillingAt,
			Category:      string(sub.Category),
			UrgencyStatus: billUrgency(sub.NextBillingAt),
		})
	}

	// Subs are already sorted by next_billing_at ASC from fetchAllActive.
	total := len(upcoming)
	offset := (page - 1) * perPage
	if page < 1 {
		offset = 0
	}
	if offset >= total {
		return &database.PaginatedResult[*UpcomingBill]{
			Items:      []*UpcomingBill{},
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: calcTotalPages(total, perPage),
		}, nil
	}
	end := offset + perPage
	if end > total {
		end = total
	}

	return &database.PaginatedResult[*UpcomingBill]{
		Items:      upcoming[offset:end],
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calcTotalPages(total, perPage),
	}, nil
}

// ─── ExportFinances ───────────────────────────────────────────────────────────

// ExportFinances generates a CSV export of active subscriptions for the given
// period. Only "csv" is supported; "pdf" requires an external library (V2).
func (s *FinancialService) ExportFinances(ctx context.Context, userID, format, period string) ([]byte, error) {
	if format != "csv" {
		return nil, ErrUnsupportedFormat
	}

	from, to, label, err := parsePeriod(period)
	if err != nil {
		return nil, err
	}

	subs, err := s.fetchAllActive(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("export finances: %w", err)
	}

	var filtered []*domainsubscription.Subscription
	for _, sub := range subs {
		if inPeriod(sub.NextBillingAt, from, to) {
			filtered = append(filtered, sub)
		}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Report header meta.
	_ = w.Write([]string{"Period", label})
	_ = w.Write([]string{"Total Subscriptions", strconv.Itoa(len(filtered))})
	_ = w.Write([]string{}) // blank separator

	// Column headers.
	_ = w.Write([]string{
		"ID", "Name", "Amount", "Currency",
		"BillingCycle", "NextBillingAt", "Category", "Status",
	})

	var totalExpenses float64
	for _, sub := range filtered {
		totalExpenses += sub.Amount
		_ = w.Write([]string{
			sub.ID,
			sub.Name,
			strconv.FormatFloat(sub.Amount, 'f', 2, 64),
			sub.Currency,
			string(sub.BillingCycle),
			sub.NextBillingAt.UTC().Format("2006-01-02"),
			string(sub.Category),
			string(sub.Status),
		})
	}

	// Summary footer.
	_ = w.Write([]string{})
	_ = w.Write([]string{"Total Expenses", strconv.FormatFloat(totalExpenses, 'f', 2, 64)})

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return buf.Bytes(), nil
}

// ─── Private helpers ──────────────────────────────────────────────────────────

// fetchAllActive retrieves up to 1000 active subscriptions sorted by
// next_billing_at ASC. 1000 is sufficient for MVP usage patterns.
func (s *FinancialService) fetchAllActive(ctx context.Context, userID string) ([]*domainsubscription.Subscription, error) {
	status := domainsubscription.SubscriptionStatusActive
	filter := database.SubscriptionFilter{Status: &status}
	p := database.PaginationParams{
		Page:    1,
		PerPage: 1000,
		Sort:    "next_billing_at",
		Order:   "asc",
	}
	result, err := s.subscriptions.FindAllByUser(ctx, userID, filter, p)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// parsePeriod converts a period string to a [from, to) UTC time window.
// "monthly" → current calendar month; "YYYY-MM" → that specific month.
func parsePeriod(period string) (from, to time.Time, label string, err error) {
	now := time.Now().UTC()
	if period == "monthly" {
		first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return first, first.AddDate(0, 1, 0), now.Format("2006-01"), nil
	}
	t, parseErr := time.Parse("2006-01", period)
	if parseErr != nil {
		return time.Time{}, time.Time{}, "", ErrInvalidPeriod
	}
	first := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return first, first.AddDate(0, 1, 0), period, nil
}

// prevPeriod returns the [from, to) window for the month immediately before
// the given month start.
func prevPeriod(monthStart time.Time) (from, to time.Time) {
	prev := monthStart.AddDate(0, -1, 0)
	return prev, monthStart
}

// inPeriod returns true when t falls in the half-open interval [from, to).
func inPeriod(t, from, to time.Time) bool {
	return !t.Before(from) && t.Before(to)
}

// sumInPeriod sums subscription amounts whose next_billing_at falls in [from, to).
func sumInPeriod(subs []*domainsubscription.Subscription, from, to time.Time) float64 {
	var total float64
	for _, sub := range subs {
		if inPeriod(sub.NextBillingAt, from, to) {
			total += sub.Amount
		}
	}
	return total
}

// mapDisplayCategory maps a domain subscription category to a UI display label.
func mapDisplayCategory(cat domainsubscription.SubscriptionCategory) string {
	switch cat {
	case domainsubscription.CategoryEntertainment:
		return "Household"
	case domainsubscription.CategoryProductivity, domainsubscription.CategoryBill:
		return "Utilities"
	default:
		return "Others"
	}
}

// buildWeeks divides [from, to) into 7-day chunks; the last chunk covers the
// remainder of the month.
func buildWeeks(from, to time.Time) []*WeeklyAmount {
	var weeks []*WeeklyAmount
	start := from
	week := 1
	for start.Before(to) {
		end := start.AddDate(0, 0, 7)
		if end.After(to) {
			end = to
		}
		weeks = append(weeks, &WeeklyAmount{
			Week:   fmt.Sprintf("Week %d", week),
			From:   start,
			To:     end,
			Amount: 0,
		})
		start = end
		week++
	}
	return weeks
}

// billUrgency returns an urgency label based on days until billing.
func billUrgency(nextBillingAt time.Time) string {
	daysUntil := time.Until(nextBillingAt).Hours() / 24
	switch {
	case daysUntil <= 3:
		return "URGENT"
	case daysUntil <= 7:
		return "PENDING"
	default:
		return "SCHEDULED"
	}
}

// calcTotalPages computes the number of pages for offset pagination.
func calcTotalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(perPage)))
}

// sortCategories sorts a SpendingCategory slice by Amount descending.
func sortCategories(cats []*SpendingCategory) {
	for i := 1; i < len(cats); i++ {
		for j := i; j > 0 && cats[j].Amount > cats[j-1].Amount; j-- {
			cats[j], cats[j-1] = cats[j-1], cats[j]
		}
	}
}
