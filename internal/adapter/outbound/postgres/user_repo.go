package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/identity"
)

// UserRepo implements database.UserRepository using pgx/v5.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

const userSelectCols = `id, email, phone, password_hash, name, avatar_url,
       timezone, plan, account_status, failed_login_count, locked_until,
       totp_secret, totp_enabled, created_at, updated_at`

func scanUser(row interface {
	Scan(dest ...any) error
}) (*identity.User, error) {
	u := &identity.User{}
	var plan, accountStatus string
	err := row.Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.Name, &u.AvatarURL,
		&u.Timezone, &plan, &accountStatus,
		&u.FailedLoginCount, &u.LockedUntil,
		&u.TOTPSecret, &u.TOTPEnabled,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Plan = identity.Plan(plan)
	u.AccountStatus = identity.AccountStatus(accountStatus)
	return u, nil
}

// Create inserts a new user row.
func (r *UserRepo) Create(ctx context.Context, u *identity.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users
		  (id, email, phone, password_hash, name, avatar_url, timezone,
		   plan, account_status, failed_login_count, locked_until,
		   totp_secret, totp_enabled, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		u.ID, u.Email, u.Phone, u.PasswordHash, u.Name, u.AvatarURL, u.Timezone,
		string(u.Plan), string(u.AccountStatus), u.FailedLoginCount, u.LockedUntil,
		u.TOTPSecret, u.TOTPEnabled, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

// FindByID returns the user with the given ID, or ErrNotFound.
func (r *UserRepo) FindByID(ctx context.Context, id string) (*identity.User, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM users WHERE id = $1 AND account_status != 'deleted'`, userSelectCols),
		id,
	)
	u, err := scanUser(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return u, err
}

// FindByEmail returns the user with the given email, or ErrNotFound.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM users WHERE email = $1 AND account_status != 'deleted'`, userSelectCols),
		email,
	)
	u, err := scanUser(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return u, err
}

// FindByPhone returns the user with the given phone, or ErrNotFound.
func (r *UserRepo) FindByPhone(ctx context.Context, phone string) (*identity.User, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM users WHERE phone = $1 AND account_status != 'deleted'`, userSelectCols),
		phone,
	)
	u, err := scanUser(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return u, err
}

// UpdateFailedLoginCount sets the failed_login_count for the user.
func (r *UserRepo) UpdateFailedLoginCount(ctx context.Context, id string, count int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET failed_login_count = $1, updated_at = NOW() WHERE id = $2`,
		count, id,
	)
	return err
}

// UpdateLockedUntil sets (or clears) the locked_until timestamp.
func (r *UserRepo) UpdateLockedUntil(ctx context.Context, id string, until *time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET locked_until = $1, updated_at = NOW() WHERE id = $2`,
		until, id,
	)
	return err
}

// UpdatePlan changes the user's subscription plan.
func (r *UserRepo) UpdatePlan(ctx context.Context, id string, plan identity.Plan) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET plan = $1, updated_at = NOW() WHERE id = $2`,
		string(plan), id,
	)
	return err
}

// Delete performs a hard delete of the user row.
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// SoftDelete marks the user account as deleted without removing the row.
func (r *UserRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET account_status = 'deleted', updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

// Update persists name and timezone changes on the user record.
func (r *UserRepo) Update(ctx context.Context, u *identity.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET name = $1, timezone = $2, updated_at = NOW() WHERE id = $3`,
		u.Name, u.Timezone, u.ID,
	)
	return err
}

// UpdateAvatarURL sets the avatar_url field for the user.
func (r *UserRepo) UpdateAvatarURL(ctx context.Context, id string, avatarURL string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2`,
		avatarURL, id,
	)
	return err
}
