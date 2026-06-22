package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/subscription"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// SubscriptionRepo implements database.SubscriptionRepository.
type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

// NewSubscriptionRepo creates a new SubscriptionRepo.
func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{pool: pool}
}

var subscriptionSortAllowlist = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "next_billing_at": {}, "name": {}, "amount": {},
}

const subscriptionSelectCols = `id, user_id, name, amount, currency,
       billing_cycle, next_billing_at, category, status, created_at, updated_at`

func scanSubscription(row interface {
	Scan(dest ...any) error
}) (*subscription.Subscription, error) {
	s := &subscription.Subscription{}
	var cycle, cat, status string
	err := row.Scan(
		&s.ID, &s.UserID, &s.Name, &s.Amount, &s.Currency,
		&cycle, &s.NextBillingAt, &cat, &status,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.BillingCycle = subscription.BillingCycle(cycle)
	s.Category = subscription.SubscriptionCategory(cat)
	s.Status = subscription.SubscriptionStatus(status)
	return s, nil
}

// Create inserts a new subscription.
func (r *SubscriptionRepo) Create(ctx context.Context, s *subscription.Subscription) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscriptions
		  (id, user_id, name, amount, currency, billing_cycle, next_billing_at, category, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.UserID, s.Name, s.Amount, s.Currency,
		string(s.BillingCycle), s.NextBillingAt,
		string(s.Category), string(s.Status),
		s.CreatedAt, s.UpdatedAt,
	)
	return err
}

// FindByID returns the subscription with the given ID, or ErrNotFound.
func (r *SubscriptionRepo) FindByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM subscriptions WHERE id = $1 AND status != 'canceled'`, subscriptionSelectCols),
		id,
	)
	s, err := scanSubscription(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return s, err
}

// FindAllByUser returns a paginated list of subscriptions for the user.
func (r *SubscriptionRepo) FindAllByUser(
	ctx context.Context, userID string,
	filter database.SubscriptionFilter, p database.PaginationParams,
) (*database.PaginatedResult[*subscription.Subscription], error) {
	order := safeOrderBy(p.Sort, p.Order, subscriptionSortAllowlist, "next_billing_at")
	args := []any{userID}
	where := "user_id = $1 AND status != 'canceled'"
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

	var total int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM subscriptions WHERE %s`, where), args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := pageOffset(p.Page, p.PerPage)
	args = append(args, p.PerPage, offset)
	query := fmt.Sprintf(`SELECT %s FROM subscriptions WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		subscriptionSelectCols, where, order, idx, idx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*subscription.Subscription
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &database.PaginatedResult[*subscription.Subscription]{
		Items: items, Total: total, Page: p.Page,
		PerPage: p.PerPage, TotalPages: totalPages(total, p.PerPage),
	}, nil
}

// Update overwrites a subscription row.
func (r *SubscriptionRepo) Update(ctx context.Context, s *subscription.Subscription) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE subscriptions SET
		  name = $1, amount = $2, currency = $3, billing_cycle = $4,
		  next_billing_at = $5, category = $6, status = $7, updated_at = NOW()
		WHERE id = $8`,
		s.Name, s.Amount, s.Currency, string(s.BillingCycle),
		s.NextBillingAt, string(s.Category), string(s.Status), s.ID,
	)
	return err
}

// SoftDelete marks a subscription as canceled.
func (r *SubscriptionRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE subscriptions SET status = 'canceled', updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

// GetMonthlyCostSummary sums active subscription amounts grouped by currency.
func (r *SubscriptionRepo) GetMonthlyCostSummary(ctx context.Context, userID string) ([]*database.MonthlyCostItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT currency, SUM(amount) AS total, COUNT(*) AS cnt
		FROM subscriptions
		WHERE user_id = $1 AND status = 'active'
		GROUP BY currency
		ORDER BY total DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*database.MonthlyCostItem
	for rows.Next() {
		item := &database.MonthlyCostItem{}
		if err := rows.Scan(&item.Currency, &item.TotalAmount, &item.SubscriptionCount); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
