package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/reminder"
)

// ReminderOccurrenceRepo implements database.ReminderOccurrenceRepository.
// The ClaimForSending method is in claim_for_sending.go.
type ReminderOccurrenceRepo struct {
	pool *pgxpool.Pool
}

// NewReminderOccurrenceRepo creates a new ReminderOccurrenceRepo.
func NewReminderOccurrenceRepo(pool *pgxpool.Pool) *ReminderOccurrenceRepo {
	return &ReminderOccurrenceRepo{pool: pool}
}

func scanOccurrence(row interface {
	Scan(dest ...any) error
}) (*reminder.ReminderOccurrence, error) {
	o := &reminder.ReminderOccurrence{}
	var status string
	err := row.Scan(
		&o.ID, &o.ReminderID, &o.UserID,
		&o.ScheduledAt, &o.CompletedAt, &status, &o.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	o.Status = reminder.OccurrenceStatus(status)
	return o, nil
}

// Create inserts a new reminder occurrence.
func (r *ReminderOccurrenceRepo) Create(ctx context.Context, o *reminder.ReminderOccurrence) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO reminder_occurrences
		  (id, reminder_id, user_id, scheduled_at, completed_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		o.ID, o.ReminderID, o.UserID,
		o.ScheduledAt, o.CompletedAt, string(o.Status), o.CreatedAt,
	)
	return err
}

// FindByReminderID returns all occurrences for the given reminder.
func (r *ReminderOccurrenceRepo) FindByReminderID(ctx context.Context, reminderID string) ([]*reminder.ReminderOccurrence, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, reminder_id, user_id, scheduled_at, completed_at, status, created_at
		FROM reminder_occurrences
		WHERE reminder_id = $1
		ORDER BY scheduled_at DESC`,
		reminderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*reminder.ReminderOccurrence
	for rows.Next() {
		o, err := scanOccurrence(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

// UpdateStatus changes the status of a single occurrence.
func (r *ReminderOccurrenceRepo) UpdateStatus(ctx context.Context, id string, status reminder.OccurrenceStatus) error {
	_, err := r.pool.Exec(ctx,
		fmt.Sprintf(`UPDATE reminder_occurrences SET status = $1 WHERE id = $2`),
		string(status), id,
	)
	return err
}
