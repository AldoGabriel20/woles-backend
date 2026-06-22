// Package family implements the Family application service.
package family

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	domainfamily "github.com/woles/woles-backend/internal/domain/family"
	domainidentity "github.com/woles/woles-backend/internal/domain/identity"
	domainreminder "github.com/woles/woles-backend/internal/domain/reminder"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrNotFound    = errors.New("family member not found")
	ErrForbidden   = errors.New("access denied")
	ErrPlanGate    = errors.New("advanced plan required for family management")
	ErrMemberLimit = errors.New("family member limit reached (max 10)")
	ErrInvalidRole = errors.New("invalid family member role")
)

// allowedRoles is the set of valid MemberRole values.
var allowedRoles = map[domainfamily.MemberRole]bool{
	domainfamily.RolePrimary: true,
	domainfamily.RoleSpouse:  true,
	domainfamily.RoleParent:  true,
	domainfamily.RoleChild:   true,
	domainfamily.RoleOther:   true,
}

// ─── Request / Response types ─────────────────────────────────────────────────

// CreateMemberRequest carries input for creating a family member.
type CreateMemberRequest struct {
	Name          string                  `json:"name"`
	Role          domainfamily.MemberRole `json:"role"`
	RelationLabel *string                 `json:"relation_label,omitempty"`
	AvatarURL     *string                 `json:"avatar_url,omitempty"`
}

// UpdateMemberRequest carries fields that may be changed.
type UpdateMemberRequest struct {
	Name          *string                  `json:"name,omitempty"`
	Role          *domainfamily.MemberRole `json:"role,omitempty"`
	RelationLabel *string                  `json:"relation_label,omitempty"`
	AvatarURL     *string                  `json:"avatar_url,omitempty"`
}

// MemberWithCount extends FamilyMember with an active reminder count.
type MemberWithCount struct {
	*domainfamily.FamilyMember
	ActiveReminderCount int `json:"active_reminder_count"`
}

// SharedReminderItem is a reminder visible across family members.
type SharedReminderItem struct {
	ReminderID  string    `json:"reminder_id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	NextRunAt   time.Time `json:"next_run_at"`
	Status      string    `json:"status"`
	OwnerName   string    `json:"owner_name"`
	OwnerAvatar *string   `json:"owner_avatar,omitempty"`
}

// ─── Service ──────────────────────────────────────────────────────────────────

// Service implements the family application service.
type Service struct {
	members   database.FamilyMemberRepository
	users     database.UserRepository
	reminders database.ReminderRepository
	auditLogs database.AuditLogRepository
}

// NewService constructs the family service.
func NewService(
	members database.FamilyMemberRepository,
	users database.UserRepository,
	reminders database.ReminderRepository,
	auditLogs database.AuditLogRepository,
) *Service {
	return &Service{
		members:   members,
		users:     users,
		reminders: reminders,
		auditLogs: auditLogs,
	}
}

// ─── Plan gate ────────────────────────────────────────────────────────────────

// checkPlan returns ErrPlanGate if the user's plan is not Advanced.
func (s *Service) checkPlan(ctx context.Context, userID string) error {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("check plan: %w", err)
	}
	if u.Plan != domainidentity.PlanAdvanced {
		return ErrPlanGate
	}
	return nil
}

// ─── CreateMember ─────────────────────────────────────────────────────────────

// CreateMember adds a new family member profile under the owner's account.
// Requires Advanced plan. Max 10 members per user.
func (s *Service) CreateMember(ctx context.Context, ownerUserID string, req CreateMemberRequest) (*domainfamily.FamilyMember, error) {
	if err := s.checkPlan(ctx, ownerUserID); err != nil {
		return nil, err
	}

	// Validate role.
	if !allowedRoles[req.Role] {
		return nil, ErrInvalidRole
	}

	// Enforce member limit.
	existing, err := s.members.FindAllByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("create member: %w", err)
	}
	if len(existing) >= 10 {
		return nil, ErrMemberLimit
	}

	// Sanitize inputs.
	name := sanitize(req.Name, 200)
	if name == "" {
		name = req.Name
	}
	var relLabel *string
	if req.RelationLabel != nil {
		v := sanitize(*req.RelationLabel, 100)
		relLabel = &v
	}

	member := &domainfamily.FamilyMember{
		ID:            uuid.NewString(),
		OwnerUserID:   ownerUserID,
		Name:          name,
		Role:          req.Role,
		RelationLabel: relLabel,
		AvatarURL:     req.AvatarURL,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.members.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("create member: %w", err)
	}

	entityType := "family_member"
	entityID := member.ID
	s.writeAudit(ctx, ownerUserID, "family_member.created", &entityType, &entityID)

	return member, nil
}

// ─── GetMembers ───────────────────────────────────────────────────────────────

// GetMembers returns all family members for the owner, each including an
// active reminder count derived from the reminder repository.
func (s *Service) GetMembers(ctx context.Context, ownerUserID string) ([]*MemberWithCount, error) {
	if err := s.checkPlan(ctx, ownerUserID); err != nil {
		return nil, err
	}

	mems, err := s.members.FindAllByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}

	// Count active reminders for the owner (all reminders are under the owner's userID).
	activeStatus := domainreminder.ReminderStatusActive
	result, err := s.reminders.FindAllByUser(ctx, ownerUserID, database.ReminderFilter{Status: &activeStatus}, database.PaginationParams{
		Page:    1,
		PerPage: 10000,
		Sort:    "next_run_at",
		Order:   "asc",
	})
	activeCount := 0
	if err == nil && result != nil {
		activeCount = len(result.Items)
	}

	out := make([]*MemberWithCount, len(mems))
	for i, m := range mems {
		out[i] = &MemberWithCount{
			FamilyMember:        m,
			ActiveReminderCount: activeCount,
		}
	}
	return out, nil
}

// ─── GetMemberByID ────────────────────────────────────────────────────────────

// GetMemberByID returns a single family member, verifying ownership.
func (s *Service) GetMemberByID(ctx context.Context, ownerUserID, memberID string) (*domainfamily.FamilyMember, error) {
	if err := s.checkPlan(ctx, ownerUserID); err != nil {
		return nil, err
	}

	m, err := s.members.FindByID(ctx, memberID)
	if err != nil {
		return nil, ErrNotFound
	}
	if m.OwnerUserID != ownerUserID {
		return nil, ErrForbidden
	}
	return m, nil
}

// ─── UpdateMember ─────────────────────────────────────────────────────────────

// UpdateMember applies a partial update to a family member.
func (s *Service) UpdateMember(ctx context.Context, ownerUserID, memberID string, req UpdateMemberRequest) (*domainfamily.FamilyMember, error) {
	if err := s.checkPlan(ctx, ownerUserID); err != nil {
		return nil, err
	}

	m, err := s.members.FindByID(ctx, memberID)
	if err != nil {
		return nil, ErrNotFound
	}
	if m.OwnerUserID != ownerUserID {
		return nil, ErrForbidden
	}

	if req.Name != nil {
		m.Name = sanitize(*req.Name, 200)
	}
	if req.Role != nil {
		if !allowedRoles[*req.Role] {
			return nil, ErrInvalidRole
		}
		m.Role = *req.Role
	}
	if req.RelationLabel != nil {
		v := sanitize(*req.RelationLabel, 100)
		m.RelationLabel = &v
	}
	if req.AvatarURL != nil {
		m.AvatarURL = req.AvatarURL
	}
	m.UpdatedAt = time.Now().UTC()

	if err := s.members.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("update member: %w", err)
	}
	return m, nil
}

// ─── DeleteMember ─────────────────────────────────────────────────────────────

// DeleteMember soft-deletes a family member. Because FamilyMemberRepository
// only exposes a hard Delete, the record is removed. Reminders owned by the
// account remain under the owner's userID.
func (s *Service) DeleteMember(ctx context.Context, ownerUserID, memberID string) error {
	if err := s.checkPlan(ctx, ownerUserID); err != nil {
		return err
	}

	m, err := s.members.FindByID(ctx, memberID)
	if err != nil {
		return ErrNotFound
	}
	if m.OwnerUserID != ownerUserID {
		return ErrForbidden
	}

	if err := s.members.Delete(ctx, memberID); err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

	entityType := "family_member"
	s.writeAudit(ctx, ownerUserID, "family_member.deleted", &entityType, &memberID)
	return nil
}

// ─── GetSharedReminders ───────────────────────────────────────────────────────

// GetSharedReminders returns paginated reminders for the owner account,
// including owner_name and owner_avatar from the primary owner's profile.
func (s *Service) GetSharedReminders(ctx context.Context, ownerUserID string, page, perPage int) (*database.PaginatedResult[*SharedReminderItem], error) {
	if err := s.checkPlan(ctx, ownerUserID); err != nil {
		return nil, err
	}

	// Look up the owner user for name/avatar.
	owner, err := s.users.FindByID(ctx, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("get shared reminders: %w", err)
	}

	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "next_run_at",
		Order:   "asc",
	}

	result, err := s.reminders.FindAllByUser(ctx, ownerUserID, database.ReminderFilter{}, p)
	if err != nil {
		return nil, fmt.Errorf("get shared reminders: %w", err)
	}

	items := make([]*SharedReminderItem, 0, len(result.Items))
	for _, r := range result.Items {
		ownerName := ""
		if owner.Name != nil {
			ownerName = *owner.Name
		}
		items = append(items, &SharedReminderItem{
			ReminderID:  r.ID,
			Title:       r.Title,
			Category:    string(r.Category),
			NextRunAt:   r.NextRunAt,
			Status:      string(r.Status),
			OwnerName:   ownerName,
			OwnerAvatar: owner.AvatarURL,
		})
	}

	return &database.PaginatedResult[*SharedReminderItem]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *Service) writeAudit(ctx context.Context, userID, action string, entityType, entityID *string) {
	id := uuid.NewString()
	now := time.Now().UTC()
	_ = s.auditLogs.Create(ctx, &database.AuditLog{
		ID:         id,
		UserID:     &userID,
		ActorType:  "user",
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		CreatedAt:  now,
	})
}

// sanitize strips leading/trailing space and truncates to maxRunes runes.
func sanitize(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > maxRunes {
		runes := []rune(s)
		s = string(runes[:maxRunes])
	}
	return s
}

// calcFamilyTotalPages computes total pages for offset pagination.
func calcFamilyTotalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(perPage)))
}
