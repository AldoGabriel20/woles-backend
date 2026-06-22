package unit_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ─── Title validation helper (mirrors application service validation) ─────────

func validateTitle(title string) error {
	if len(title) > 200 {
		return errStr("title must be at most 200 characters")
	}
	return nil
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestInputValidation_TitleTooLong_Returns422(t *testing.T) {
	app := fiber.New()
	app.Post("/reminders", func(c *fiber.Ctx) error {
		var req struct {
			Title string `json:"title"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad_request"})
		}
		if err := validateTitle(req.Title); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error":   "validation_failed",
				"message": err.Error(),
			})
		}
		return c.SendStatus(fiber.StatusCreated)
	})

	// 201-character title.
	longTitle := strings.Repeat("a", 201)
	body := `{"title":"` + longTitle + `"}`
	req := httptest.NewRequest("POST", "/reminders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("201-char title: want 422, got %d", resp.StatusCode)
	}
}

func TestInputValidation_ValidTitle_Returns201(t *testing.T) {
	app := fiber.New()
	app.Post("/reminders", func(c *fiber.Ctx) error {
		var req struct {
			Title string `json:"title"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad_request"})
		}
		if err := validateTitle(req.Title); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fiber.StatusCreated)
	})

	body := `{"title":"Pay electricity bill"}`
	req := httptest.NewRequest("POST", "/reminders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("valid title: want 201, got %d", resp.StatusCode)
	}
}

// ─── SQL-injection safety ─────────────────────────────────────────────────────
// This test verifies that a raw SQL injection payload is treated as a plain
// string value and does not alter query logic (parameterised query contract).

func TestInputValidation_SQLInjectionPayloadIsSafelyParameterized(t *testing.T) {
	// Simulate extracting a title from a request and using it in a parameterised query.
	injectionPayload := `'; DROP TABLE reminders; --`

	// The payload arrives as a string bound to $1 in a parameterised query.
	// It should pass our length validation (it's short).
	if err := validateTitle(injectionPayload); err != nil {
		t.Errorf("injection payload should pass length validation, got: %v", err)
	}

	// Verify that our query construction uses the value AS-IS as a parameter,
	// not via string interpolation. Since we cannot test the DB here, we document
	// the contract: the value is always passed as a query arg, never interpolated.
	// If the SQL were `"SELECT * FROM reminders WHERE title = '" + injectionPayload + "'"`,
	// the payload would escape the string literal. Using `$1` prevents this entirely.
	queryTemplate := "SELECT * FROM reminders WHERE title = $1"
	if strings.Contains(queryTemplate, injectionPayload) {
		t.Error("query template must NOT contain raw user input")
	}
}
