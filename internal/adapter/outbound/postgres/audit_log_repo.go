package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// AuditLogRepo implements database.AuditLogRepository.
type AuditLogRepo struct {
	pool *pgxpool.Pool
}

// NewAuditLogRepo creates a new AuditLogRepo.
func NewAuditLogRepo(pool *pgxpool.Pool) *AuditLogRepo {
	return &AuditLogRepo{pool: pool}
}

var auditLogSortAllowlist = map[string]struct{}{
	"id": {}, "created_at": {}, "action": {},
}

const auditLogSelectCols = `id, user_id, actor_type, action, entity_type, entity_id, ip_address, user_agent, created_at`

func scanAuditLog(row interface {
	Scan(dest ...any) error
}) (*database.AuditLog, error) {
	l := &database.AuditLog{}
	err := row.Scan(
		&l.ID, &l.UserID, &l.ActorType, &l.Action,
		&l.EntityType, &l.EntityID, &l.IPAddress, &l.UserAgent, &l.CreatedAt,
	)
	return l, err
}

// Create inserts a new audit log entry.
func (r *AuditLogRepo) Create(ctx context.Context, l *database.AuditLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs
		  (id, user_id, actor_type, action, entity_type, entity_id, ip_address, user_agent, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		l.ID, l.UserID, l.ActorType, l.Action,
		l.EntityType, l.EntityID, l.IPAddress, l.UserAgent, l.CreatedAt,
	)
	return err
}

// FindAllByUser returns a paginated list of audit log entries for the user.
func (r *AuditLogRepo) FindAllByUser(
	ctx context.Context, userID string, p database.PaginationParams,
) (*database.PaginatedResult[*database.AuditLog], error) {
	order := safeOrderBy(p.Sort, p.Order, auditLogSortAllowlist, "created_at")

	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM audit_logs WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, err
	}

	offset := pageOffset(p.Page, p.PerPage)
	rows, err := r.pool.Query(ctx,
		"SELECT "+auditLogSelectCols+" FROM audit_logs WHERE user_id = $1 ORDER BY "+order+" LIMIT $2 OFFSET $3",
		userID, p.PerPage, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*database.AuditLog
	for rows.Next() {
		l, err := scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &database.PaginatedResult[*database.AuditLog]{
		Items: items, Total: total, Page: p.Page,
		PerPage: p.PerPage, TotalPages: totalPages(total, p.PerPage),
	}, nil
}
