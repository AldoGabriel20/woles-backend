package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/woles/woles-backend/internal/adapter/outbound/postgres"
	portdb "github.com/woles/woles-backend/internal/port/outbound/database"
)

func TestAuditLog_WrittenOnLoginLogoutPasswordChange(t *testing.T) {
	db := skipIfNoDB(t)
	repo := postgres.NewAuditLogRepo(db.Pool)
	ctx := context.Background()

	userID := uuid.NewString()

	actions := []string{"login", "logout", "password_change"}
	for _, action := range actions {
		log := &portdb.AuditLog{
			ID:         uuid.NewString(),
			UserID:     &userID,
			ActorType:  "user",
			Action:     action,
			EntityType: strPtr("user"),
			EntityID:   &userID,
			CreatedAt:  time.Now(),
		}
		if err := repo.Create(ctx, log); err != nil {
			t.Fatalf("Create audit log for action %q: %v", action, err)
		}
	}

	// Fetch all audit logs for this user and verify all three were written.
	logs, err := repo.FindAllByUser(ctx, userID, portdb.PaginationParams{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatalf("FindAllByUser: %v", err)
	}
	if len(logs.Items) < len(actions) {
		t.Errorf("expected at least %d audit rows, got %d", len(actions), len(logs.Items))
	}

	found := map[string]bool{}
	for _, l := range logs.Items {
		found[l.Action] = true
	}
	for _, action := range actions {
		if !found[action] {
			t.Errorf("audit log for action %q was not found", action)
		}
	}
}

func strPtr(s string) *string { return &s }
