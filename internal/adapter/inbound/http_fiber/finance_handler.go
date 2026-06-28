package http_fiber

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v2"
	appsubscription "github.com/woles/woles-backend/internal/application/subscription"
	domainsubscription "github.com/woles/woles-backend/internal/domain/subscription"
	database "github.com/woles/woles-backend/internal/port/outbound/database"
	"github.com/xuri/excelize/v2"
)

type financeHandler struct{ svc *appsubscription.Service }

// RegisterFinanceRoutes mounts all /api/v1/finances routes.
func RegisterFinanceRoutes(router fiber.Router, svc *Services) {
	h := &financeHandler{svc: svc.Subscription}
	f := router.Group("/finances")
	f.Get("/summary", h.summary)
	f.Get("/monthly-costs", h.monthlyCosts)
	f.Get("/export", h.export)
}

func (h *financeHandler) summary(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	items, err := h.svc.GetMonthlyCostSummary(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}

	// Build a FinancialSummary object the frontend expects.
	totalMonthly := 0.0
	currency := "IDR"
	subCount := 0
	for _, item := range items {
		if item.Currency == "IDR" {
			totalMonthly = item.TotalAmount
			currency = item.Currency
			subCount = item.SubscriptionCount
		}
	}
	// Fall back to first item if no IDR entry exists.
	if subCount == 0 && len(items) > 0 {
		totalMonthly = items[0].TotalAmount
		currency = items[0].Currency
		subCount = items[0].SubscriptionCount
	}

	return c.JSON(fiber.Map{"summary": fiber.Map{
		"total_monthly_cost": totalMonthly,
		"currency":           currency,
		"subscription_count": subCount,
		"active_goals":       0,
		"upcoming_bills":     subCount,
	}})
}

func (h *financeHandler) monthlyCosts(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	costs, err := h.svc.GetMonthlyCostSummary(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"monthly_costs": costs})
}

func (h *financeHandler) export(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	format := c.Query("format", "csv")

	result, err := h.svc.GetSubscriptions(c.Context(), userID, database.SubscriptionFilter{}, 1, 500)
	if err != nil {
		return mapServiceError(c, err)
	}
	subs := result.Items

	switch format {
	case "pdf":
		data, buildErr := financeExportPDF(subs)
		if buildErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate PDF"})
		}
		c.Set("Content-Type", "application/pdf")
		c.Set("Content-Disposition", `attachment; filename="finances.pdf"`)
		return c.Send(data)
	case "excel":
		data, buildErr := financeExportExcel(subs)
		if buildErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate Excel"})
		}
		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", `attachment; filename="finances.xlsx"`)
		return c.Send(data)
	default: // csv
		var buf bytes.Buffer
		buf.WriteString("Name,Amount,Currency,Billing Cycle,Category,Status,Next Billing Date\n")
		for _, s := range subs {
			buf.WriteString(fmt.Sprintf("%q,%g,%s,%s,%s,%s,%s\n",
				s.Name, s.Amount, s.Currency,
				string(s.BillingCycle), string(s.Category), string(s.Status),
				s.NextBillingAt.UTC().Format("2006-01-02"),
			))
		}
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", `attachment; filename="finances.csv"`)
		return c.Send(buf.Bytes())
	}
}

func financeExportPDF(subs []*domainsubscription.Subscription) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(0, 53, 39)
	pdf.CellFormat(0, 10, "Financial Overview - Subscriptions", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 6, "Generated: "+time.Now().UTC().Format("2 Jan 2006, 15:04 UTC"), "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Header row
	pdf.SetFillColor(0, 53, 39)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 9)
	cols := []struct {
		w float64
		h string
	}{
		{70, "Name"}, {30, "Amount"}, {20, "Currency"},
		{30, "Billing"}, {35, "Category"}, {25, "Status"}, {37, "Next Billing"},
	}
	for _, col := range cols {
		pdf.CellFormat(col.w, 8, col.h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Data rows
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(30, 30, 30)
	for i, s := range subs {
		if i%2 == 0 {
			pdf.SetFillColor(240, 248, 244)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		fill := true
		pdf.CellFormat(70, 7, s.Name, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%.0f", s.Amount), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(20, 7, s.Currency, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(30, 7, string(s.BillingCycle), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(35, 7, string(s.Category), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(25, 7, string(s.Status), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(37, 7, s.NextBillingAt.UTC().Format("2 Jan 2006"), "1", 0, "C", fill, 0, "")
		pdf.Ln(-1)
	}

	// Footer
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total: %d subscriptions", len(subs)), "", 0, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func financeExportExcel(subs []*domainsubscription.Subscription) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Subscriptions"
	f.SetSheetName("Sheet1", sheet)

	// Title row
	f.SetCellValue(sheet, "A1", "Financial Overview - Subscriptions")
	f.SetCellValue(sheet, "A2", "Generated: "+time.Now().UTC().Format("2 Jan 2006, 15:04 UTC"))

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "003527"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheet, "A1", "G1", titleStyle)
	f.MergeCell(sheet, "A1", "G1")
	f.MergeCell(sheet, "A2", "G2")

	// Header row
	headers := []string{"Name", "Amount", "Currency", "Billing Cycle", "Category", "Status", "Next Billing Date"}
	hStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"003527"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, hStyle)
	}

	// Data rows
	evenStyle, _ := f.NewStyle(&excelize.Style{Fill: excelize.Fill{Type: "pattern", Color: []string{"F0F8F4"}, Pattern: 1}})
	for ri, s := range subs {
		row := ri + 5
		vals := []interface{}{
			s.Name,
			s.Amount,
			s.Currency,
			string(s.BillingCycle),
			string(s.Category),
			string(s.Status),
			s.NextBillingAt.UTC().Format("2006-01-02"),
		}
		for ci, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(ci+1, row)
			f.SetCellValue(sheet, cell, v)
			if ri%2 == 0 {
				f.SetCellStyle(sheet, cell, cell, evenStyle)
			}
		}
	}

	// Column widths
	widths := []float64{30, 12, 10, 14, 15, 12, 18}
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
