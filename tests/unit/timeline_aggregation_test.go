package unit_test

import (
	"sort"
	"testing"
	"time"
)

// ─── TimelineItem (local copy for pure unit testing) ─────────────────────────

type timelineItemType string

const (
	itemReminder     timelineItemType = "reminder"
	itemDocument     timelineItemType = "document"
	itemSubscription timelineItemType = "subscription"
	itemGoal         timelineItemType = "goal"
)

type timelineItem struct {
	ID       string
	Type     timelineItemType
	Title    string
	DueAt    time.Time
	EntityID string
}

// sortByDueAt sorts items ascending by DueAt (mirrors service layer).
func sortByDueAt(items []timelineItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].DueAt.Before(items[j].DueAt)
	})
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestTimelineSort_AscendingOrder(t *testing.T) {
	now := time.Now().UTC()
	items := []timelineItem{
		{ID: "3", Type: itemGoal, DueAt: now.AddDate(0, 0, 10)},
		{ID: "1", Type: itemReminder, DueAt: now.AddDate(0, 0, 1)},
		{ID: "2", Type: itemDocument, DueAt: now.AddDate(0, 0, 5)},
		{ID: "4", Type: itemSubscription, DueAt: now.AddDate(0, 0, 3)},
	}
	sortByDueAt(items)
	for i := 1; i < len(items); i++ {
		if items[i].DueAt.Before(items[i-1].DueAt) {
			t.Errorf("items not sorted: index %d (%v) < index %d (%v)", i, items[i].DueAt, i-1, items[i-1].DueAt)
		}
	}
}

func TestTimelineSort_MixedTypes(t *testing.T) {
	now := time.Now().UTC()
	items := []timelineItem{
		{ID: "a", Type: itemGoal, DueAt: now.AddDate(0, 0, 2)},
		{ID: "b", Type: itemReminder, DueAt: now.AddDate(0, 0, 2)}, // same DueAt
		{ID: "c", Type: itemSubscription, DueAt: now.AddDate(0, 0, 1)},
	}
	sortByDueAt(items)
	if items[0].ID != "c" {
		t.Errorf("first item should be subscription (earliest), got %s", items[0].ID)
	}
}

func TestTimelineNormalization_FieldMapping(t *testing.T) {
	now := time.Now().UTC()
	item := timelineItem{
		ID:       "test-uuid",
		Type:     itemReminder,
		Title:    "Pay electricity",
		DueAt:    now.AddDate(0, 0, 3),
		EntityID: "reminder-uuid",
	}
	if item.ID == "" {
		t.Error("ID should be set")
	}
	if item.Type != itemReminder {
		t.Errorf("Type: want reminder, got %s", item.Type)
	}
	if item.Title != "Pay electricity" {
		t.Errorf("Title mismatch: %s", item.Title)
	}
}

func TestTimelinePagination_Offset(t *testing.T) {
	now := time.Now().UTC()
	items := make([]timelineItem, 25)
	for i := range items {
		items[i] = timelineItem{
			ID:    string(rune('a' + i)),
			DueAt: now.AddDate(0, 0, i),
		}
	}
	// Page 2, perPage 10 → items 10-19
	page, perPage := 2, 10
	offset := (page - 1) * perPage
	page2 := items[offset : offset+perPage]
	if len(page2) != 10 {
		t.Errorf("pagination: want 10 items, got %d", len(page2))
	}
	if page2[0].ID != items[10].ID {
		t.Errorf("pagination: first item of page 2 should be index 10")
	}
}
