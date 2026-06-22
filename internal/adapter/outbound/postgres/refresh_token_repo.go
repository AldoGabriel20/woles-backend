package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// RefreshTokenRepo implements database.RefreshTokenRepository.
type RefreshTokenRepo struct {
	pool *pgxpool.Pool
}

// NewRefreshTokenRepo creates a new RefreshTokenRepo.
func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool}
}

const refreshTokenSelectCols = `id, user_id, token_hash, family_id, expires_at, revoked_at, created_at`

func scanRefreshToken(row interface {
	Scan(dest ...any) error
}) (*database.RefreshToken, error) {
	t := &database.RefreshToken{}
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.FamilyID, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Create inserts a new refresh token row.
func (r *RefreshTokenRepo) Create(ctx context.Context, t *database.RefreshToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, family_id, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		t.ID, t.UserID, t.TokenHash, t.FamilyID, t.ExpiresAt, t.RevokedAt, t.CreatedAt,
	)
	return err
}

// FindByID returns the refresh token with the given ID, or ErrNotFound.
func (r *RefreshTokenRepo) FindByID(ctx context.Context, id string) (*database.RefreshToken, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM refresh_tokens WHERE id = $1`, refreshTokenSelectCols),
		id,
	)
	t, err := scanRefreshToken(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return t, err
}

// FindByFamilyID returns all refresh tokens belonging to the given family.
func (r *RefreshTokenRepo) FindByFamilyID(ctx context.Context, familyID string) ([]*database.RefreshToken, error) {
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT %s FROM refresh_tokens WHERE family_id = $1`, refreshTokenSelectCols),
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tokens []*database.RefreshToken
	for rows.Next() {
		t, err := scanRefreshToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// Revoke marks a single refresh token as revoked.
func (r *RefreshTokenRepo) Revoke(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2`,
		now, id,
	)
	return err
}

// RevokeAllForUser revokes all refresh tokens for a user.
func (r *RefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	return err
}

// RevokeFamily revokes all tokens in a refresh token family.
func (r *RefreshTokenRepo) RevokeFamily(ctx context.Context, familyID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE family_id = $1 AND revoked_at IS NULL`,
		familyID,
	)
	return err
}
