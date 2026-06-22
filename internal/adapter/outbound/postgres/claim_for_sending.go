package postgres

import (
	"context"
	"fmt"

	"github.com/woles/woles-backend/internal/domain/notification"
	"github.com/woles/woles-backend/internal/domain/reminder"
)

// ClaimDue atomically claims up to batchSize notifications that are ready to
// send (status='scheduled', scheduled_at <= NOW()) using FOR UPDATE SKIP LOCKED.
// Claimed rows have their status set to 'sending'.
func (r *NotificationRepo) ClaimDue(ctx context.Context, batchSize int) ([]*notification.Notification, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ClaimDue: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT `+notificationSelectCols+`
		FROM notifications
		WHERE status = 'scheduled' AND scheduled_at <= NOW()
		ORDER BY scheduled_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`,
		batchSize,
	)
	if err != nil {
		return nil, fmt.Errorf("ClaimDue: select: %w", err)
	}

	var claimed []*notification.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("ClaimDue: scan: %w", err)
		}
		claimed = append(claimed, n)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ClaimDue: rows: %w", err)
	}

	if len(claimed) == 0 {
		return nil, tx.Commit(ctx)
	}

	// Collect IDs, then UPDATE them.
	ids := make([]string, len(claimed))
	for i, n := range claimed {
		ids[i] = n.ID
	}
	_, err = tx.Exec(ctx, `
		UPDATE notifications
		SET status = 'sending', updated_at = NOW()
		WHERE id = ANY($1::uuid[])`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("ClaimDue: update: %w", err)
	}
	// Reflect new status in returned structs.
	for _, n := range claimed {
		n.Status = notification.StatusSending
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ClaimDue: commit: %w", err)
	}
	return claimed, nil
}

// ClaimForSending atomically claims up to batchSize scheduled reminder
// occurrences using FOR UPDATE SKIP LOCKED. Claimed rows have their status
// set to 'sent' (processing hand-off to the notification worker).
func (r *ReminderOccurrenceRepo) ClaimForSending(ctx context.Context, batchSize int) ([]*reminder.ReminderOccurrence, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ClaimForSending: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id, reminder_id, user_id, scheduled_at, completed_at, status, created_at
		FROM reminder_occurrences
		WHERE status = 'scheduled' AND scheduled_at <= NOW()
		ORDER BY scheduled_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`,
		batchSize,
	)
	if err != nil {
		return nil, fmt.Errorf("ClaimForSending: select: %w", err)
	}

	var claimed []*reminder.ReminderOccurrence
	for rows.Next() {
		o, err := scanOccurrence(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("ClaimForSending: scan: %w", err)
		}
		claimed = append(claimed, o)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ClaimForSending: rows: %w", err)
	}

	if len(claimed) == 0 {
		return nil, tx.Commit(ctx)
	}

	ids := make([]string, len(claimed))
	for i, o := range claimed {
		ids[i] = o.ID
	}
	_, err = tx.Exec(ctx, `
		UPDATE reminder_occurrences
		SET status = 'sent'
		WHERE id = ANY($1::uuid[])`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("ClaimForSending: update: %w", err)
	}
	for _, o := range claimed {
		o.Status = reminder.OccurrenceSent
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ClaimForSending: commit: %w", err)
	}
	return claimed, nil
}
