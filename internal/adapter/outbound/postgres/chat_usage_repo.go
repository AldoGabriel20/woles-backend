package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/chat"
)

// ChatUsageRepo implements database.ChatUsageRepository.
type ChatUsageRepo struct {
	pool *pgxpool.Pool
}

// NewChatUsageRepo creates a new ChatUsageRepo.
func NewChatUsageRepo(pool *pgxpool.Pool) *ChatUsageRepo {
	return &ChatUsageRepo{pool: pool}
}

func scanChatUsage(row interface {
	Scan(dest ...any) error
}) (*chat.ChatUsage, error) {
	u := &chat.ChatUsage{}
	err := row.Scan(&u.ID, &u.UserID, &u.Month, &u.MessagesUsed, &u.Quota, &u.UpdatedAt)
	return u, err
}

// GetOrCreate returns the chat_usage row for the given user+month, creating it if absent.
// Uses INSERT ... ON CONFLICT DO NOTHING to avoid races.
func (r *ChatUsageRepo) GetOrCreate(ctx context.Context, userID string, month time.Time) (*chat.ChatUsage, error) {
	m := firstOfMonth(month)
	// Ensure the row exists (upsert with no change on conflict).
	_, err := r.pool.Exec(ctx, `
		INSERT INTO chat_usage (id, user_id, month, messages_used, quota, updated_at)
		VALUES (gen_random_uuid(), $1, $2, 0, 10, NOW())
		ON CONFLICT (user_id, month) DO NOTHING`,
		userID, m,
	)
	if err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, month, messages_used, quota, updated_at FROM chat_usage WHERE user_id = $1 AND month = $2`,
		userID, m,
	)
	u, scanErr := scanChatUsage(row)
	if isNotFound(scanErr) {
		return nil, ErrNotFound
	}
	return u, scanErr
}

// Increment adds 1 to messages_used for the given user+month row.
func (r *ChatUsageRepo) Increment(ctx context.Context, userID string, month time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE chat_usage
		SET messages_used = messages_used + 1, updated_at = NOW()
		WHERE user_id = $1 AND month = $2`,
		userID, firstOfMonth(month),
	)
	return err
}

// GetQuota returns the chat_usage row for the given user+month, or ErrNotFound.
func (r *ChatUsageRepo) GetQuota(ctx context.Context, userID string, month time.Time) (*chat.ChatUsage, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, month, messages_used, quota, updated_at FROM chat_usage WHERE user_id = $1 AND month = $2`,
		userID, firstOfMonth(month),
	)
	u, err := scanChatUsage(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return u, err
}

// firstOfMonth returns the first day of the month containing t, in UTC.
func firstOfMonth(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
