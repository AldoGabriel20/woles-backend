package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/notification"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// NotificationRepo implements database.NotificationRepository.
// The ClaimDue method is in claim_for_sending.go.
type NotificationRepo struct {
	pool *pgxpool.Pool
}

// NewNotificationRepo creates a new NotificationRepo.
func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

var notificationSortAllowlist = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "scheduled_at": {}, "status": {},
}

const notificationSelectCols = `id, user_id, entity_type, entity_id, occurrence_id,
       channel, scheduled_at, sent_at, status, idempotency_key,
       provider_message_id, failure_reason, retry_count, created_at, updated_at`

func scanNotification(row interface {
	Scan(dest ...any) error
}) (*notification.Notification, error) {
	n := &notification.Notification{}
	var entityType, channel, status string
	err := row.Scan(
		&n.ID, &n.UserID, &entityType, &n.EntityID, &n.OccurrenceID,
		&channel, &n.ScheduledAt, &n.SentAt, &status, &n.IdempotencyKey,
		&n.ProviderMessageID, &n.FailureReason, &n.RetryCount,
		&n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	n.EntityType = notification.NotificationEntityType(entityType)
	n.Channel = notification.NotificationChannel(channel)
	n.Status = notification.NotificationStatus(status)
	return n, nil
}

// Create inserts a new notification.
func (r *NotificationRepo) Create(ctx context.Context, n *notification.Notification) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications
		  (id, user_id, entity_type, entity_id, occurrence_id,
		   channel, scheduled_at, sent_at, status, idempotency_key,
		   provider_message_id, failure_reason, retry_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		n.ID, n.UserID, string(n.EntityType), n.EntityID, n.OccurrenceID,
		string(n.Channel), n.ScheduledAt, n.SentAt, string(n.Status), n.IdempotencyKey,
		n.ProviderMessageID, n.FailureReason, n.RetryCount, n.CreatedAt, n.UpdatedAt,
	)
	return err
}

// FindByID returns the notification with the given ID, or ErrNotFound.
func (r *NotificationRepo) FindByID(ctx context.Context, id string) (*notification.Notification, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM notifications WHERE id = $1`, notificationSelectCols),
		id,
	)
	n, err := scanNotification(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return n, err
}

// FindAllByUser returns a paginated list of notifications for the user.
func (r *NotificationRepo) FindAllByUser(
	ctx context.Context, userID string,
	filter database.NotificationFilter, p database.PaginationParams,
) (*database.PaginatedResult[*notification.Notification], error) {
	order := safeOrderBy(p.Sort, p.Order, notificationSortAllowlist, "scheduled_at")
	args := []any{userID}
	where := "user_id = $1"
	idx := 2

	if filter.EntityType != nil {
		where += fmt.Sprintf(" AND entity_type = $%d", idx)
		args = append(args, string(*filter.EntityType))
		idx++
	}
	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(*filter.Status))
		idx++
	}
	if filter.From != nil {
		where += fmt.Sprintf(" AND scheduled_at >= $%d", idx)
		args = append(args, *filter.From)
		idx++
	}
	if filter.To != nil {
		where += fmt.Sprintf(" AND scheduled_at <= $%d", idx)
		args = append(args, *filter.To)
		idx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM notifications WHERE %s`, where), args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := pageOffset(p.Page, p.PerPage)
	args = append(args, p.PerPage, offset)
	query := fmt.Sprintf(`SELECT %s FROM notifications WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		notificationSelectCols, where, order, idx, idx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*notification.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &database.PaginatedResult[*notification.Notification]{
		Items: items, Total: total, Page: p.Page,
		PerPage: p.PerPage, TotalPages: totalPages(total, p.PerPage),
	}, nil
}

// UpdateStatus changes the status of a notification.
func (r *NotificationRepo) UpdateStatus(ctx context.Context, id string, status notification.NotificationStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET status = $1, updated_at = NOW() WHERE id = $2`,
		string(status), id,
	)
	return err
}

// IncrementRetry increments the retry_count for a notification.
func (r *NotificationRepo) IncrementRetry(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET retry_count = retry_count + 1, updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

// GetStats returns delivery statistics for the given user.
func (r *NotificationRepo) GetStats(ctx context.Context, userID string) (*database.NotificationStats, error) {
	stats := &database.NotificationStats{}
	err := r.pool.QueryRow(ctx, `
		SELECT
		  COUNT(*) FILTER (WHERE status = 'sent') AS total_sent,
		  COUNT(*) FILTER (WHERE status = 'failed') AS total_failed
		FROM notifications
		WHERE user_id = $1`,
		userID,
	).Scan(&stats.TotalSent, &stats.TotalFailed)
	if err != nil {
		return nil, err
	}
	total := stats.TotalSent + stats.TotalFailed
	if total > 0 {
		stats.DeliveryRatePct = float64(stats.TotalSent) / float64(total) * 100
	}

	// Top category by number of sent notifications.
	row := r.pool.QueryRow(ctx, `
		SELECT entity_type
		FROM notifications
		WHERE user_id = $1 AND status = 'sent'
		GROUP BY entity_type
		ORDER BY COUNT(*) DESC
		LIMIT 1`,
		userID,
	)
	var topCat string
	if err := row.Scan(&topCat); err == nil {
		stats.TopCategory = topCat
	}
	return stats, nil
}

// ExportRange returns all notifications for the user within the time range.
func (r *NotificationRepo) ExportRange(ctx context.Context, userID string, from, to time.Time) ([]*notification.Notification, error) {
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT %s FROM notifications
		WHERE user_id = $1 AND scheduled_at >= $2 AND scheduled_at <= $3
		ORDER BY scheduled_at ASC`, notificationSelectCols),
		userID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*notification.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}
