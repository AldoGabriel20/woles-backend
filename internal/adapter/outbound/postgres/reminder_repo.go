package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/reminder"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ReminderRepo implements database.ReminderRepository.
type ReminderRepo struct {
	pool *pgxpool.Pool
}

// NewReminderRepo creates a new ReminderRepo.
func NewReminderRepo(pool *pgxpool.Pool) *ReminderRepo {
	return &ReminderRepo{pool: pool}
}

var reminderSortAllowlist = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "next_run_at": {}, "title": {}, "status": {},
}

const reminderSelectCols = `id, user_id, title, category, recurrence_type, recurrence_rule,
       next_run_at, timezone, status, source, original_text, created_at, updated_at`

func scanReminder(row interface {
	Scan(dest ...any) error
}) (*reminder.Reminder, error) {
	r := &reminder.Reminder{}
	var cat, rt, status, source string
	var recurrenceRule []byte
	err := row.Scan(
		&r.ID, &r.UserID, &r.Title, &cat, &rt, &recurrenceRule,
		&r.NextRunAt, &r.Timezone, &status, &source, &r.OriginalText,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	r.Category = reminder.ReminderCategory(cat)
	r.RecurrenceType = reminder.RecurrenceType(rt)
	r.Status = reminder.ReminderStatus(status)
	r.Source = reminder.ReminderSource(source)
	if len(recurrenceRule) > 0 {
		r.RecurrenceRule = json.RawMessage(recurrenceRule)
	}
	return r, nil
}

// Create inserts a new reminder.
func (r *ReminderRepo) Create(ctx context.Context, rem *reminder.Reminder) error {
	var ruleJSON []byte
	if len(rem.RecurrenceRule) > 0 {
		ruleJSON = rem.RecurrenceRule
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO reminders
		  (id, user_id, title, category, recurrence_type, recurrence_rule,
		   next_run_at, timezone, status, source, original_text, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		rem.ID, rem.UserID, rem.Title,
		string(rem.Category), string(rem.RecurrenceType), ruleJSON,
		rem.NextRunAt, rem.Timezone, string(rem.Status), string(rem.Source),
		rem.OriginalText, rem.CreatedAt, rem.UpdatedAt,
	)
	return err
}

// FindByID returns the reminder with the given ID, or ErrNotFound.
func (r *ReminderRepo) FindByID(ctx context.Context, id string) (*reminder.Reminder, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM reminders WHERE id = $1 AND status != 'archived'`, reminderSelectCols),
		id,
	)
	rem, err := scanReminder(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return rem, err
}

// FindAllByUser returns a paginated list of reminders for the user.
func (r *ReminderRepo) FindAllByUser(
	ctx context.Context, userID string,
	filter database.ReminderFilter, p database.PaginationParams,
) (*database.PaginatedResult[*reminder.Reminder], error) {
	order := safeOrderBy(p.Sort, p.Order, reminderSortAllowlist, "next_run_at")
	args := []any{userID}
	where := "user_id = $1 AND status != 'archived'"
	idx := 2

	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(*filter.Status))
		idx++
	}
	if filter.Category != nil {
		where += fmt.Sprintf(" AND category = $%d", idx)
		args = append(args, string(*filter.Category))
		idx++
	}
	if filter.Search != nil && *filter.Search != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", idx)
		args = append(args, "%"+*filter.Search+"%")
		idx++
	}

	var total int
	countRow := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM reminders WHERE %s`, where),
		args...,
	)
	if err := countRow.Scan(&total); err != nil {
		return nil, err
	}

	offset := pageOffset(p.Page, p.PerPage)
	args = append(args, p.PerPage, offset)
	query := fmt.Sprintf(`SELECT %s FROM reminders WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		reminderSelectCols, where, order, idx, idx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*reminder.Reminder
	for rows.Next() {
		rem, err := scanReminder(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, rem)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &database.PaginatedResult[*reminder.Reminder]{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalPages: totalPages(total, p.PerPage),
	}, nil
}

// Update overwrites a reminder row with the current state.
func (r *ReminderRepo) Update(ctx context.Context, rem *reminder.Reminder) error {
	var ruleJSON []byte
	if len(rem.RecurrenceRule) > 0 {
		ruleJSON = rem.RecurrenceRule
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE reminders SET
		  title = $1, category = $2, recurrence_type = $3, recurrence_rule = $4,
		  next_run_at = $5, timezone = $6, status = $7, source = $8,
		  original_text = $9, updated_at = NOW()
		WHERE id = $10`,
		rem.Title, string(rem.Category), string(rem.RecurrenceType), ruleJSON,
		rem.NextRunAt, rem.Timezone, string(rem.Status), string(rem.Source),
		rem.OriginalText, rem.ID,
	)
	return err
}

// SoftDelete archives a reminder by setting status to 'archived'.
func (r *ReminderRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE reminders SET status = 'archived', updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

// FindDueOccurrences returns scheduled occurrences with scheduled_at <= before.
func (r *ReminderRepo) FindDueOccurrences(ctx context.Context, before time.Time, limit int) ([]*reminder.ReminderOccurrence, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, reminder_id, user_id, scheduled_at, completed_at, status, created_at
		FROM reminder_occurrences
		WHERE status = 'scheduled' AND scheduled_at <= $1
		ORDER BY scheduled_at ASC
		LIMIT $2`,
		before, limit,
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
