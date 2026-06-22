package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// TimelineRepo implements database.TimelineRepository.
// It aggregates reminder_occurrences, documents, subscriptions, and goals into
// a single sorted view.
type TimelineRepo struct {
	pool *pgxpool.Pool
}

// NewTimelineRepo creates a new TimelineRepo.
func NewTimelineRepo(pool *pgxpool.Pool) *TimelineRepo {
	return &TimelineRepo{pool: pool}
}

// GetTimelineItems returns a paginated list of timeline items across entity types.
func (r *TimelineRepo) GetTimelineItems(
	ctx context.Context,
	userID string,
	from, to time.Time,
	p database.PaginationParams,
) (*database.PaginatedResult[*database.TimelineItem], error) {
	// Union query: each branch emits (id, type, title, due_at, status, entity_id).
	// All parameters are positional to avoid SQL injection.
	unionSQL := `
		SELECT
		  ro.id            AS id,
		  'reminder'       AS type,
		  r.title          AS title,
		  ro.scheduled_at  AS due_at,
		  ro.status        AS status,
		  r.id             AS entity_id
		FROM reminder_occurrences ro
		JOIN reminders r ON r.id = ro.reminder_id
		WHERE ro.user_id = $1
		  AND ro.status = 'scheduled'
		  AND ro.scheduled_at BETWEEN $2 AND $3

		UNION ALL

		SELECT
		  d.id             AS id,
		  'document'       AS type,
		  d.title          AS title,
		  d.expiry_date    AS due_at,
		  d.status         AS status,
		  d.id             AS entity_id
		FROM documents d
		WHERE d.user_id = $1
		  AND d.status = 'active'
		  AND d.expiry_date IS NOT NULL
		  AND d.expiry_date BETWEEN $2 AND $3

		UNION ALL

		SELECT
		  s.id             AS id,
		  'subscription'   AS type,
		  s.name           AS title,
		  s.next_billing_at AS due_at,
		  s.status         AS status,
		  s.id             AS entity_id
		FROM subscriptions s
		WHERE s.user_id = $1
		  AND s.status = 'active'
		  AND s.next_billing_at BETWEEN $2 AND $3

		UNION ALL

		SELECT
		  g.id             AS id,
		  'goal'           AS type,
		  g.title          AS title,
		  g.target_date    AS due_at,
		  g.status         AS status,
		  g.id             AS entity_id
		FROM goals g
		WHERE g.user_id = $1
		  AND g.status = 'active'
		  AND g.target_date IS NOT NULL
		  AND g.target_date BETWEEN $2 AND $3`

	// Count total items.
	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM ("+unionSQL+") t",
		userID, from, to,
	).Scan(&total); err != nil {
		return nil, err
	}

	offset := pageOffset(p.Page, p.PerPage)
	rows, err := r.pool.Query(ctx,
		unionSQL+" ORDER BY due_at ASC LIMIT $4 OFFSET $5",
		userID, from, to, p.PerPage, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*database.TimelineItem
	for rows.Next() {
		item := &database.TimelineItem{}
		if err := rows.Scan(
			&item.ID, &item.Type, &item.Title, &item.DueAt, &item.Status, &item.EntityID,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &database.PaginatedResult[*database.TimelineItem]{
		Items: items, Total: total, Page: p.Page,
		PerPage: p.PerPage, TotalPages: totalPages(total, p.PerPage),
	}, nil
}
