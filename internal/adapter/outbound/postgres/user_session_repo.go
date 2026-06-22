package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// UserSessionRepo implements database.UserSessionRepository.
type UserSessionRepo struct {
	pool *pgxpool.Pool
}

// NewUserSessionRepo creates a new UserSessionRepo.
func NewUserSessionRepo(pool *pgxpool.Pool) *UserSessionRepo {
	return &UserSessionRepo{pool: pool}
}

const userSessionSelectCols = `id, user_id, refresh_token_id, device_name, ip_address, user_agent, last_active_at, created_at`

func scanUserSession(row interface {
	Scan(dest ...any) error
}) (*database.UserSession, error) {
	s := &database.UserSession{}
	err := row.Scan(
		&s.ID, &s.UserID, &s.RefreshTokenID,
		&s.DeviceName, &s.IPAddress, &s.UserAgent,
		&s.LastActiveAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Create inserts a new user session.
func (r *UserSessionRepo) Create(ctx context.Context, s *database.UserSession) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_sessions
		  (id, user_id, refresh_token_id, device_name, ip_address, user_agent, last_active_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		s.ID, s.UserID, s.RefreshTokenID,
		s.DeviceName, s.IPAddress, s.UserAgent,
		s.LastActiveAt, s.CreatedAt,
	)
	return err
}

// FindAllByUser returns all sessions for the given user.
func (r *UserSessionRepo) FindAllByUser(ctx context.Context, userID string) ([]*database.UserSession, error) {
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT %s FROM user_sessions WHERE user_id = $1 ORDER BY last_active_at DESC`, userSessionSelectCols),
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []*database.UserSession
	for rows.Next() {
		s, err := scanUserSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// FindByID returns the session with the given ID, or ErrNotFound.
func (r *UserSessionRepo) FindByID(ctx context.Context, id string) (*database.UserSession, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM user_sessions WHERE id = $1`, userSessionSelectCols),
		id,
	)
	s, err := scanUserSession(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return s, err
}

// UpdateLastActive updates the last_active_at timestamp for the session.
func (r *UserSessionRepo) UpdateLastActive(ctx context.Context, id string, at time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_sessions SET last_active_at = $1 WHERE id = $2`,
		at, id,
	)
	return err
}

// Delete hard-deletes a single session row.
func (r *UserSessionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_sessions WHERE id = $1`, id)
	return err
}

// DeleteAllForUser removes all sessions for the user.
func (r *UserSessionRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}
