package integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/woles/woles-backend/internal/adapter/outbound/postgres"
	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
)

// TestPostgres_NotificationClaim_ForUpdateSkipLocked verifies that concurrent
// ClaimDue calls do not double-claim the same notification row.
func TestPostgres_NotificationClaim_ForUpdateSkipLocked(t *testing.T) {
	db := skipIfNoDB(t)
	repo := postgres.NewNotificationRepo(db.Pool)
	ctx := context.Background()

	userID := uuid.NewString()

	// Insert a batch of scheduled notifications.
	const batchSize = 10
	for i := 0; i < batchSize; i++ {
		n := &domainnotification.Notification{
			ID:             uuid.NewString(),
			UserID:         userID,
			EntityType:     domainnotification.EntityReminder,
			EntityID:       uuid.NewString(),
			Channel:        domainnotification.ChannelWhatsApp,
			ScheduledAt:    time.Now().Add(-1 * time.Minute), // due now
			Status:         domainnotification.StatusScheduled,
			IdempotencyKey: "test:" + uuid.NewString(),
		}
		if err := repo.Create(ctx, n); err != nil {
			t.Fatalf("Create notification %d: %v", i, err)
		}
	}

	// Run two concurrent ClaimDue calls.
	var (
		mu      sync.Mutex
		claimed []string
	)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			notifications, err := repo.ClaimDue(ctx, batchSize)
			if err != nil {
				return
			}
			mu.Lock()
			for _, n := range notifications {
				claimed = append(claimed, n.ID)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Check for duplicates.
	seen := map[string]int{}
	for _, id := range claimed {
		seen[id]++
	}
	for id, count := range seen {
		if count > 1 {
			t.Errorf("notification %s was claimed %d times (expected 1)", id, count)
		}
	}
	if len(claimed) > batchSize {
		t.Errorf("claimed %d notifications from a batch of %d — expect at most %d", len(claimed), batchSize, batchSize)
	}
}
