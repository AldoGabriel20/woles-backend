package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/goal"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// GoalRepo implements database.GoalRepository.
type GoalRepo struct {
	pool *pgxpool.Pool
}

// NewGoalRepo creates a new GoalRepo.
func NewGoalRepo(pool *pgxpool.Pool) *GoalRepo {
	return &GoalRepo{pool: pool}
}

var goalSortAllowlist = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "target_date": {}, "title": {},
}

const goalSelectCols = `id, user_id, title, icon, target_amount, current_amount,
       monthly_target, currency, target_date, status, created_at, updated_at`

func scanGoal(row interface {
	Scan(dest ...any) error
}) (*goal.Goal, error) {
	g := &goal.Goal{}
	var status, iconStr string
	var iconPtr *string
	err := row.Scan(
		&g.ID, &g.UserID, &g.Title, &iconPtr,
		&g.TargetAmount, &g.CurrentAmount, &g.MonthlyTarget,
		&g.Currency, &g.TargetDate, &status,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	g.Status = goal.GoalStatus(status)
	_ = iconStr
	if iconPtr != nil {
		ic := goal.GoalIcon(*iconPtr)
		g.Icon = &ic
	}
	return g, nil
}

// Create inserts a new goal.
func (r *GoalRepo) Create(ctx context.Context, g *goal.Goal) error {
	var icon *string
	if g.Icon != nil {
		s := string(*g.Icon)
		icon = &s
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO goals
		  (id, user_id, title, icon, target_amount, current_amount,
		   monthly_target, currency, target_date, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		g.ID, g.UserID, g.Title, icon,
		g.TargetAmount, g.CurrentAmount, g.MonthlyTarget,
		g.Currency, g.TargetDate, string(g.Status),
		g.CreatedAt, g.UpdatedAt,
	)
	return err
}

// FindByID returns the goal with the given ID, or ErrNotFound.
func (r *GoalRepo) FindByID(ctx context.Context, id string) (*goal.Goal, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM goals WHERE id = $1 AND status != 'archived'`, goalSelectCols),
		id,
	)
	g, err := scanGoal(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return g, err
}

// FindAllByUser returns a paginated list of goals.
func (r *GoalRepo) FindAllByUser(
	ctx context.Context, userID string,
	filter database.GoalFilter, p database.PaginationParams,
) (*database.PaginatedResult[*goal.Goal], error) {
	order := safeOrderBy(p.Sort, p.Order, goalSortAllowlist, "created_at")
	args := []any{userID}
	where := "user_id = $1 AND status != 'archived'"
	idx := 2

	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(*filter.Status))
		idx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM goals WHERE %s`, where), args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := pageOffset(p.Page, p.PerPage)
	args = append(args, p.PerPage, offset)
	query := fmt.Sprintf(`SELECT %s FROM goals WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		goalSelectCols, where, order, idx, idx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*goal.Goal
	for rows.Next() {
		g, err := scanGoal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &database.PaginatedResult[*goal.Goal]{
		Items: items, Total: total, Page: p.Page,
		PerPage: p.PerPage, TotalPages: totalPages(total, p.PerPage),
	}, nil
}

// FindActiveGoal returns the first active goal for the user, or ErrNotFound.
func (r *GoalRepo) FindActiveGoal(ctx context.Context, userID string) (*goal.Goal, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM goals WHERE user_id = $1 AND status = 'active' ORDER BY created_at ASC LIMIT 1`, goalSelectCols),
		userID,
	)
	g, err := scanGoal(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return g, err
}

// UpdateProgress sets the current_amount for a goal.
func (r *GoalRepo) UpdateProgress(ctx context.Context, id string, currentAmount float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE goals SET current_amount = $1, updated_at = NOW() WHERE id = $2`,
		currentAmount, id,
	)
	return err
}

// Update overwrites a goal row.
func (r *GoalRepo) Update(ctx context.Context, g *goal.Goal) error {
	var icon *string
	if g.Icon != nil {
		s := string(*g.Icon)
		icon = &s
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE goals SET
		  title = $1, icon = $2, target_amount = $3, current_amount = $4,
		  monthly_target = $5, currency = $6, target_date = $7, status = $8,
		  updated_at = NOW()
		WHERE id = $9`,
		g.Title, icon, g.TargetAmount, g.CurrentAmount,
		g.MonthlyTarget, g.Currency, g.TargetDate, string(g.Status), g.ID,
	)
	return err
}

// SoftDelete archives a goal.
func (r *GoalRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE goals SET status = 'archived', updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}
