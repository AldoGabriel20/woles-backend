package unit_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ─── Stub repository ─────────────────────────────────────────────────────────

// stubReminderStore is a simple in-memory map to test ownership logic without
// a real database.
type stubReminderEntry struct {
	id     string
	userID string
}

type stubReminderStore struct {
	data []stubReminderEntry
}

var errNotFoundOwnership = errors.New("not found")

func (s *stubReminderStore) findForUser(_ context.Context, userID, reminderID string) error {
	for _, e := range s.data {
		if e.id == reminderID {
			if e.userID != userID {
				// Item exists but belongs to a different user → return 404
				// to avoid confirming resource existence.
				return errNotFoundOwnership
			}
			return nil
		}
	}
	return errNotFoundOwnership
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestOwnership_UserACannotAccessUserBReminder(t *testing.T) {
	store := &stubReminderStore{
		data: []stubReminderEntry{
			{id: "reminder-B", userID: "user-B"},
		},
	}

	app := fiber.New()
	app.Get("/reminders/:id", func(c *fiber.Ctx) error {
		callerUserID := "user-A"
		reminderID := c.Params("id")
		if err := store.findForUser(c.Context(), callerUserID, reminderID); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "not_found",
			})
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/reminders/reminder-B", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("user A accessing user B's reminder: want 404, got %d", resp.StatusCode)
	}
}

func TestOwnership_UserCanAccessOwnReminder(t *testing.T) {
	store := &stubReminderStore{
		data: []stubReminderEntry{
			{id: "reminder-A", userID: "user-A"},
		},
	}

	app := fiber.New()
	app.Get("/reminders/:id", func(c *fiber.Ctx) error {
		callerUserID := "user-A"
		reminderID := c.Params("id")
		if err := store.findForUser(c.Context(), callerUserID, reminderID); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not_found"})
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/reminders/reminder-A", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("user A accessing own reminder: want 200, got %d", resp.StatusCode)
	}
}
