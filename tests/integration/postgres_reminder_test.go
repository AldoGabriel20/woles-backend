package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/woles/woles-backend/internal/adapter/outbound/postgres"
	domainreminder "github.com/woles/woles-backend/internal/domain/reminder"
)

func skipIfNoDB(t *testing.T) *postgres.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}
	os.Setenv("DATABASE_URL", url)
	db, err := postgres.New(context.Background())
	if err != nil {
		t.Fatalf("connect to test DB: %v", err)
	}
	return db
}

func TestPostgres_Reminder_CreateFetchUpdateSoftDelete(t *testing.T) {
	db := skipIfNoDB(t)
	repo := postgres.NewReminderRepo(db.Pool)
	ctx := context.Background()

	userID := uuid.NewString()
	r := &domainreminder.Reminder{
		ID:             uuid.NewString(),
		UserID:         userID,
		Title:          "Test reminder",
		Category:       domainreminder.CategoryCustom,
		RecurrenceType: domainreminder.RecurrenceOneTime,
		NextRunAt:      time.Now().Add(24 * time.Hour).UTC(),
		Timezone:       "Asia/Jakarta",
		Status:         domainreminder.ReminderStatusActive,
		Source:         domainreminder.SourceWeb,
	}

	// Create
	if err := repo.Create(ctx, r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Fetch
	fetched, err := repo.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if fetched.Title != r.Title {
		t.Errorf("Title mismatch: want %q, got %q", r.Title, fetched.Title)
	}

	// Update
	fetched.Title = "Updated title"
	if err := repo.Update(ctx, fetched); err != nil {
		t.Fatalf("Update: %v", err)
	}
	updated, _ := repo.FindByID(ctx, r.ID)
	if updated.Title != "Updated title" {
		t.Errorf("Update: title not persisted")
	}

	// Soft delete
	if err := repo.SoftDelete(ctx, r.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	// After soft delete the row should not be returned.
	_, err = repo.FindByID(ctx, r.ID)
	if err == nil {
		t.Error("FindByID after SoftDelete: expected not-found error")
	}
}
