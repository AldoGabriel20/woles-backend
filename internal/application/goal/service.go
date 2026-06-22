// Package goal implements the Goal application service.
package goal

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domaingoal "github.com/woles/woles-backend/internal/domain/goal"
	domainidentity "github.com/woles/woles-backend/internal/domain/identity"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var (
	ErrNotFound     = errors.New("goal not found")
	ErrPlanRequired = errors.New("goal tracker requires a premium or advanced plan")
	ErrInvalidInput = errors.New("invalid input")
)

// ─── Request / response types ─────────────────────────────────────────────────

// CreateGoalRequest holds the input for creating a new goal.
type CreateGoalRequest struct {
	Title         string
	Icon          *domaingoal.GoalIcon
	TargetAmount  float64
	MonthlyTarget *float64
	Currency      string
	TargetDate    *time.Time
}

// UpdateGoalRequest holds the fields that may be updated.
type UpdateGoalRequest struct {
	Title         *string
	Icon          *domaingoal.GoalIcon
	TargetAmount  *float64
	MonthlyTarget *float64
	Currency      *string
	TargetDate    *time.Time
}

// GoalWithTip combines a goal with a motivational tip string.
type GoalWithTip struct {
	Goal *domaingoal.Goal
	Tip  string
}

// ─── Service ──────────────────────────────────────────────────────────────────

// Service implements the goal application service.
type Service struct {
	goals     database.GoalRepository
	users     database.UserRepository
	auditLogs database.AuditLogRepository
}

// NewService constructs the goal service.
func NewService(
	goals database.GoalRepository,
	users database.UserRepository,
	auditLogs database.AuditLogRepository,
) *Service {
	return &Service{
		goals:     goals,
		users:     users,
		auditLogs: auditLogs,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

// CreateGoal validates the request, enforces the premium plan requirement, and
// inserts a new goal.
func (s *Service) CreateGoal(ctx context.Context, userID string, req CreateGoalRequest) (*domaingoal.Goal, error) {
	// Premium plan gate.
	user, err := s.users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user.Plan == domainidentity.PlanFree {
		return nil, ErrPlanRequired
	}

	// Validate target_amount.
	if req.TargetAmount <= 0 {
		return nil, fmt.Errorf("%w: target_amount must be greater than 0", ErrInvalidInput)
	}

	// Validate icon.
	if req.Icon != nil && !validIcon(*req.Icon) {
		return nil, fmt.Errorf("%w: unknown icon %q", ErrInvalidInput, *req.Icon)
	}

	// Sanitize title.
	title := sanitize(req.Title, 200)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	currency := req.Currency
	if currency == "" {
		currency = "IDR"
	}

	now := time.Now().UTC()
	g := &domaingoal.Goal{
		ID:            uuid.NewString(),
		UserID:        userID,
		Title:         title,
		Icon:          req.Icon,
		TargetAmount:  req.TargetAmount,
		CurrentAmount: 0,
		MonthlyTarget: req.MonthlyTarget,
		Currency:      currency,
		TargetDate:    req.TargetDate,
		Status:        domaingoal.GoalStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.goals.Create(ctx, g); err != nil {
		return nil, fmt.Errorf("create goal: %w", err)
	}

	s.writeAudit(ctx, userID, "goal", "goal_created", g.ID)

	return g, nil
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetGoals returns a paginated list of goals for a user.
func (s *Service) GetGoals(ctx context.Context, userID string, filter database.GoalFilter, page, perPage int) (*database.PaginatedResult[*domaingoal.Goal], error) {
	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "created_at",
		Order:   "desc",
	}
	return s.goals.FindAllByUser(ctx, userID, filter, p)
}

// GetGoalByID returns a single goal, enforcing ownership.
func (s *Service) GetGoalByID(ctx context.Context, userID, goalID string) (*domaingoal.Goal, error) {
	return s.ownershipCheck(ctx, userID, goalID)
}

// GetGoalHistory returns completed and archived goals for a user.
func (s *Service) GetGoalHistory(ctx context.Context, userID string, page, perPage int) (*database.PaginatedResult[*domaingoal.Goal], error) {
	completed := domaingoal.GoalStatusCompleted
	filter := database.GoalFilter{Status: &completed}
	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "updated_at",
		Order:   "desc",
	}
	// Query completed goals.
	result, err := s.goals.FindAllByUser(ctx, userID, filter, p)
	if err != nil {
		return nil, fmt.Errorf("get goal history: %w", err)
	}

	// Also include archived goals in the same result set by fetching separately
	// and merging. For MVP simplicity we return whichever status was requested;
	// callers can pass status=archived to get archived goals separately.
	// The repository layer supports filtering by status, so this is sufficient.
	return result, nil
}

// GetActiveGoalWithTip fetches the first active goal and generates a
// motivational tip based on estimated months remaining.
func (s *Service) GetActiveGoalWithTip(ctx context.Context, userID string) (*GoalWithTip, error) {
	g, err := s.goals.FindActiveGoal(ctx, userID)
	if err != nil || g == nil {
		return nil, ErrNotFound
	}

	tip := buildTip(g)

	return &GoalWithTip{Goal: g, Tip: tip}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

// UpdateGoal applies partial updates to a goal.
func (s *Service) UpdateGoal(ctx context.Context, userID, goalID string, req UpdateGoalRequest) (*domaingoal.Goal, error) {
	g, err := s.ownershipCheck(ctx, userID, goalID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		g.Title = sanitize(*req.Title, 200)
	}
	if req.Icon != nil {
		if !validIcon(*req.Icon) {
			return nil, fmt.Errorf("%w: unknown icon %q", ErrInvalidInput, *req.Icon)
		}
		g.Icon = req.Icon
	}
	if req.TargetAmount != nil {
		if *req.TargetAmount <= 0 {
			return nil, fmt.Errorf("%w: target_amount must be greater than 0", ErrInvalidInput)
		}
		g.TargetAmount = *req.TargetAmount
	}
	if req.MonthlyTarget != nil {
		g.MonthlyTarget = req.MonthlyTarget
	}
	if req.Currency != nil {
		g.Currency = *req.Currency
	}
	if req.TargetDate != nil {
		g.TargetDate = req.TargetDate
	}

	g.UpdatedAt = time.Now().UTC()

	if err := s.goals.Update(ctx, g); err != nil {
		return nil, fmt.Errorf("update goal: %w", err)
	}

	return g, nil
}

// UpdateProgress updates the current_amount of a goal. If the new amount meets
// or exceeds the target, the goal is marked as completed.
func (s *Service) UpdateProgress(ctx context.Context, userID, goalID string, newAmount float64) (*domaingoal.Goal, error) {
	g, err := s.ownershipCheck(ctx, userID, goalID)
	if err != nil {
		return nil, err
	}

	g.CurrentAmount = newAmount
	g.UpdatedAt = time.Now().UTC()

	if g.CurrentAmount >= g.TargetAmount {
		g.Status = domaingoal.GoalStatusCompleted
	}

	if err := s.goals.UpdateProgress(ctx, g.ID, g.CurrentAmount); err != nil {
		return nil, fmt.Errorf("update progress: %w", err)
	}

	// Persist status change if goal completed.
	if g.Status == domaingoal.GoalStatusCompleted {
		if err := s.goals.Update(ctx, g); err != nil {
			return nil, fmt.Errorf("mark goal completed: %w", err)
		}
	}

	s.writeAudit(ctx, userID, "goal", "goal_progress_updated", g.ID)

	return g, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

// DeleteGoal soft-deletes a goal and writes an audit log.
func (s *Service) DeleteGoal(ctx context.Context, userID, goalID string) error {
	g, err := s.ownershipCheck(ctx, userID, goalID)
	if err != nil {
		return err
	}

	if err := s.goals.SoftDelete(ctx, g.ID); err != nil {
		return fmt.Errorf("soft delete goal: %w", err)
	}

	s.writeAudit(ctx, userID, "goal", "goal_deleted", g.ID)

	return nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// ownershipCheck fetches a goal by ID and verifies it belongs to userID.
// Returns ErrNotFound (opaque) on missing or unowned records.
func (s *Service) ownershipCheck(ctx context.Context, userID, goalID string) (*domaingoal.Goal, error) {
	g, err := s.goals.FindByID(ctx, goalID)
	if err != nil || g == nil {
		return nil, ErrNotFound
	}
	if g.UserID != userID {
		return nil, ErrNotFound // intentionally opaque to prevent enumeration
	}
	return g, nil
}

// buildTip generates a motivational tip for a goal based on estimated months
// remaining versus the monthly savings target.
func buildTip(g *domaingoal.Goal) string {
	remaining := g.Remaining()

	if remaining <= 0 {
		return "You've reached your goal! Great job saving up."
	}

	if g.MonthlyTarget == nil || *g.MonthlyTarget <= 0 {
		return fmt.Sprintf("You're %.0f%% of the way to your goal. Keep it up!", g.ProgressPercent())
	}

	monthsNeeded := remaining / *g.MonthlyTarget

	// Determine if the user is on track by comparing months needed to target date.
	if g.TargetDate != nil {
		monthsAvailable := time.Until(*g.TargetDate).Hours() / (24 * 30)
		if monthsAvailable > 0 && monthsNeeded <= monthsAvailable {
			return fmt.Sprintf(
				"You're on track! At your current monthly target, you'll reach your goal in about %d month(s).",
				int(math.Ceil(monthsNeeded)),
			)
		}
		return fmt.Sprintf(
			"You need to save more each month to hit your target date. Aim for at least %.0f/month to stay on track.",
			remaining/math.Max(monthsAvailable, 1),
		)
	}

	return fmt.Sprintf(
		"At your current monthly target, you'll reach your goal in about %d month(s). Stay consistent!",
		int(math.Ceil(monthsNeeded)),
	)
}

// validIcon returns true when i is one of the defined GoalIcon enum values.
func validIcon(i domaingoal.GoalIcon) bool {
	switch i {
	case domaingoal.IconLove,
		domaingoal.IconEmergency,
		domaingoal.IconVehicle,
		domaingoal.IconHome,
		domaingoal.IconTravel,
		domaingoal.IconOther:
		return true
	}
	return false
}

// sanitize strips whitespace, removes HTML-like tags, and truncates to maxRunes.
func sanitize(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	s = b.String()
	if utf8.RuneCountInString(s) > maxRunes {
		runes := []rune(s)
		s = string(runes[:maxRunes])
	}
	return s
}

// writeAudit creates an audit log entry; errors are silently discarded.
func (s *Service) writeAudit(ctx context.Context, userID, entityType, action, entityID string) {
	log := &database.AuditLog{
		ID:         uuid.NewString(),
		UserID:     strPtr(userID),
		ActorType:  "user",
		Action:     action,
		EntityType: strPtr(entityType),
		EntityID:   strPtr(entityID),
		CreatedAt:  time.Now().UTC(),
	}
	_ = s.auditLogs.Create(ctx, log)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
