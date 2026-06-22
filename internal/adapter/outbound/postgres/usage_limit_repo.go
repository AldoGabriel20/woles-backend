package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/billing"
)

// UsageLimitRepo implements database.UsageLimitRepository.
type UsageLimitRepo struct {
	pool *pgxpool.Pool
}

// NewUsageLimitRepo creates a new UsageLimitRepo.
func NewUsageLimitRepo(pool *pgxpool.Pool) *UsageLimitRepo {
	return &UsageLimitRepo{pool: pool}
}

// Get returns the usage_limits row for the given user, or ErrNotFound.
func (r *UsageLimitRepo) Get(ctx context.Context, userID string) (*billing.UsageLimit, error) {
	ul := &billing.UsageLimit{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, reminders_used, documents_used, subscriptions_used, updated_at
		FROM usage_limits
		WHERE user_id = $1`,
		userID,
	).Scan(&ul.ID, &ul.UserID, &ul.RemindersUsed, &ul.DocumentsUsed, &ul.SubscriptionsUsed, &ul.UpdatedAt)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return ul, err
}

// validResource checks the resource name against the safe allowlist.
func validResource(resource string) (string, error) {
	switch resource {
	case "reminders":
		return "reminders_used", nil
	case "documents":
		return "documents_used", nil
	case "subscriptions":
		return "subscriptions_used", nil
	default:
		return "", errors.New("unknown resource: " + resource)
	}
}

// Increment adds 1 to the named resource counter for the user.
func (r *UsageLimitRepo) Increment(ctx context.Context, userID string, resource string) error {
	col, err := validResource(resource)
	if err != nil {
		return err
	}
	// col is from an internal allowlist — safe to embed.
	_, err = r.pool.Exec(ctx,
		fmt.Sprintf(`UPDATE usage_limits SET %s = %s + 1, updated_at = NOW() WHERE user_id = $1`, col, col),
		userID,
	)
	return err
}

// Decrement subtracts 1 from the named resource counter (floor 0) for the user.
func (r *UsageLimitRepo) Decrement(ctx context.Context, userID string, resource string) error {
	col, err := validResource(resource)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		fmt.Sprintf(`UPDATE usage_limits SET %s = GREATEST(%s - 1, 0), updated_at = NOW() WHERE user_id = $1`, col, col),
		userID,
	)
	return err
}

// IsWithinLimit returns true when the user's current usage for resource is below their plan limit.
// It joins usage_limits with users and plans to resolve the limit.
func (r *UsageLimitRepo) IsWithinLimit(ctx context.Context, userID string, resource string) (bool, error) {
	col, err := validResource(resource)
	if err != nil {
		return false, err
	}
	// Map resource counter column to the plans limit column.
	planCol := map[string]string{
		"reminders_used":     "reminder_limit",
		"documents_used":     "document_limit",
		"subscriptions_used": "subscription_limit",
	}[col]

	// Safely embed the allowlisted column names.
	query := fmt.Sprintf(`
		SELECT
		  CASE
		    WHEN p.%s < 0 THEN TRUE
		    ELSE ul.%s < p.%s
		  END AS within_limit
		FROM usage_limits ul
		JOIN users u ON u.id = ul.user_id
		JOIN plans p ON p.name = u.plan
		WHERE ul.user_id = $1`, planCol, col, planCol)

	var within bool
	err = r.pool.QueryRow(ctx, query, userID).Scan(&within)
	if isNotFound(err) {
		// No usage_limits row — treat as within limit (row will be created on first increment).
		return true, nil
	}
	return within, err
}
