// Package reminder implements the Reminder application service.
package reminder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
	domainreminder "github.com/woles/woles-backend/internal/domain/reminder"
	"github.com/woles/woles-backend/internal/port/outbound/database"
	"github.com/woles/woles-backend/internal/port/outbound/message"
)

// ─── Errors ──────────────────────────────────────────────────────────────────

var (
	ErrNotFound     = errors.New("reminder not found")
	ErrForbidden    = errors.New("forbidden")
	ErrPlanLimit    = errors.New("reminder limit reached for your plan")
	ErrInvalidInput = errors.New("invalid input")
)

// ─── Request / response types ─────────────────────────────────────────────────

// CreateReminderRequest holds the input for creating a new reminder.
type CreateReminderRequest struct {
	Title          string
	Category       domainreminder.ReminderCategory
	RecurrenceType domainreminder.RecurrenceType
	RecurrenceRule []byte // raw JSONB; required for custom_interval
	NextRunAt      time.Time
	Timezone       string
	Source         domainreminder.ReminderSource
	OriginalText   *string
}

// UpdateReminderRequest holds the fields that may be updated.
type UpdateReminderRequest struct {
	Title          *string
	Category       *domainreminder.ReminderCategory
	RecurrenceType *domainreminder.RecurrenceType
	RecurrenceRule []byte
	NextRunAt      *time.Time
	Timezone       *string
}

// ─── Service ──────────────────────────────────────────────────────────────────

// Service implements the reminder application service.
type Service struct {
	reminders     database.ReminderRepository
	occurrences   database.ReminderOccurrenceRepository
	notifications database.NotificationRepository
	usageLimits   database.UsageLimitRepository
	auditLogs     database.AuditLogRepository
	publisher     message.EventPublisher
}

// NewService constructs the reminder service.
func NewService(
	reminders database.ReminderRepository,
	occurrences database.ReminderOccurrenceRepository,
	notifications database.NotificationRepository,
	usageLimits database.UsageLimitRepository,
	auditLogs database.AuditLogRepository,
	publisher message.EventPublisher,
) *Service {
	return &Service{
		reminders:     reminders,
		occurrences:   occurrences,
		notifications: notifications,
		usageLimits:   usageLimits,
		auditLogs:     auditLogs,
		publisher:     publisher,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

// CreateReminder validates the request, enforces plan limits, and inserts a new
// reminder with its first occurrence and a scheduled notification.
func (s *Service) CreateReminder(ctx context.Context, userID string, req CreateReminderRequest) (*domainreminder.Reminder, error) {
	// Plan limit check.
	within, err := s.usageLimits.IsWithinLimit(ctx, userID, "reminders")
	if err != nil {
		return nil, fmt.Errorf("check usage limit: %w", err)
	}
	if !within {
		return nil, ErrPlanLimit
	}

	// Validate recurrence type.
	if !validRecurrenceType(req.RecurrenceType) {
		return nil, fmt.Errorf("%w: unknown recurrence_type %q", ErrInvalidInput, req.RecurrenceType)
	}

	// Validate next_run_at is in the future.
	if !req.NextRunAt.After(time.Now()) {
		return nil, fmt.Errorf("%w: next_run_at must be in the future", ErrInvalidInput)
	}

	// Sanitize strings.
	title := sanitize(req.Title, 200)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	var origText *string
	if req.OriginalText != nil {
		t := sanitize(*req.OriginalText, 4000)
		origText = &t
	}

	tz := req.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}

	category := req.Category
	if category == "" {
		category = domainreminder.CategoryCustom
	}

	now := time.Now().UTC()
	rem := &domainreminder.Reminder{
		ID:             uuid.NewString(),
		UserID:         userID,
		Title:          title,
		Category:       category,
		RecurrenceType: req.RecurrenceType,
		RecurrenceRule: req.RecurrenceRule,
		NextRunAt:      req.NextRunAt.UTC(),
		Timezone:       tz,
		Status:         domainreminder.ReminderStatusActive,
		Source:         req.Source,
		OriginalText:   origText,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.reminders.Create(ctx, rem); err != nil {
		return nil, fmt.Errorf("create reminder: %w", err)
	}

	// Increment usage counter.
	_ = s.usageLimits.Increment(ctx, userID, "reminders")

	// Create first occurrence.
	occ := &domainreminder.ReminderOccurrence{
		ID:          uuid.NewString(),
		ReminderID:  rem.ID,
		UserID:      userID,
		ScheduledAt: rem.NextRunAt,
		Status:      domainreminder.OccurrenceScheduled,
		CreatedAt:   now,
	}
	if err := s.occurrences.Create(ctx, occ); err != nil {
		return nil, fmt.Errorf("create occurrence: %w", err)
	}

	// Create notification tied to the first occurrence.
	notif := buildNotification(userID, rem.ID, occ.ID, occ.ScheduledAt)
	if err := s.notifications.Create(ctx, notif); err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	// Publish domain event.
	_ = s.publisher.Publish(ctx, "reminder.created", map[string]string{
		"reminder_id": rem.ID,
		"user_id":     userID,
	})

	s.writeAudit(ctx, userID, "reminder", "reminder_created", rem.ID)

	return rem, nil
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetReminders returns a paginated list of reminders for a user, applying
// optional filters and sorting by next_run_at ASC.
func (s *Service) GetReminders(ctx context.Context, userID string, filter database.ReminderFilter, page, perPage int) (*database.PaginatedResult[*domainreminder.Reminder], error) {
	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "next_run_at",
		Order:   "asc",
	}
	return s.reminders.FindAllByUser(ctx, userID, filter, p)
}

// GetReminderByID returns a single reminder, enforcing ownership.
func (s *Service) GetReminderByID(ctx context.Context, userID, reminderID string) (*domainreminder.Reminder, error) {
	rem, err := s.reminders.FindByID(ctx, reminderID)
	if err != nil || rem == nil {
		return nil, ErrNotFound
	}
	if rem.UserID != userID {
		return nil, ErrNotFound // intentionally opaque to prevent enumeration
	}
	return rem, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

// UpdateReminder applies partial updates to a reminder. If recurrence or
// next_run_at change, the first pending occurrence and its notification are
// recreated.
func (s *Service) UpdateReminder(ctx context.Context, userID, reminderID string, req UpdateReminderRequest) (*domainreminder.Reminder, error) {
	rem, err := s.GetReminderByID(ctx, userID, reminderID)
	if err != nil {
		return nil, err
	}

	recurrenceChanged := false

	if req.Title != nil {
		rem.Title = sanitize(*req.Title, 200)
	}
	if req.Category != nil {
		rem.Category = *req.Category
	}
	if req.RecurrenceType != nil {
		if !validRecurrenceType(*req.RecurrenceType) {
			return nil, fmt.Errorf("%w: unknown recurrence_type", ErrInvalidInput)
		}
		if rem.RecurrenceType != *req.RecurrenceType {
			recurrenceChanged = true
		}
		rem.RecurrenceType = *req.RecurrenceType
	}
	if req.RecurrenceRule != nil {
		rem.RecurrenceRule = req.RecurrenceRule
		recurrenceChanged = true
	}
	if req.NextRunAt != nil {
		rem.NextRunAt = req.NextRunAt.UTC()
		recurrenceChanged = true
	}
	if req.Timezone != nil {
		rem.Timezone = *req.Timezone
	}
	rem.UpdatedAt = time.Now().UTC()

	if err := s.reminders.Update(ctx, rem); err != nil {
		return nil, fmt.Errorf("update reminder: %w", err)
	}

	if recurrenceChanged {
		// Cancel pending occurrences and recreate.
		s.cancelPendingNotifications(ctx, rem.ID)
		now := time.Now().UTC()
		occ := &domainreminder.ReminderOccurrence{
			ID:          uuid.NewString(),
			ReminderID:  rem.ID,
			UserID:      userID,
			ScheduledAt: rem.NextRunAt,
			Status:      domainreminder.OccurrenceScheduled,
			CreatedAt:   now,
		}
		_ = s.occurrences.Create(ctx, occ)
		notif := buildNotification(userID, rem.ID, occ.ID, occ.ScheduledAt)
		_ = s.notifications.Create(ctx, notif)
	}

	_ = s.publisher.Publish(ctx, "reminder.updated", map[string]string{
		"reminder_id": rem.ID,
		"user_id":     userID,
	})

	return rem, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

// DeleteReminder soft-deletes a reminder, cancels its pending notifications,
// and decrements the usage counter.
func (s *Service) DeleteReminder(ctx context.Context, userID, reminderID string) error {
	rem, err := s.GetReminderByID(ctx, userID, reminderID)
	if err != nil {
		return err
	}

	if err := s.reminders.SoftDelete(ctx, rem.ID); err != nil {
		return fmt.Errorf("soft delete reminder: %w", err)
	}

	s.cancelPendingNotifications(ctx, rem.ID)
	_ = s.usageLimits.Decrement(ctx, userID, "reminders")
	s.writeAudit(ctx, userID, "reminder", "reminder_deleted", rem.ID)

	return nil
}

// ─── Pause / Resume ───────────────────────────────────────────────────────────

// PauseReminder sets status=paused and cancels pending notifications.
func (s *Service) PauseReminder(ctx context.Context, userID, reminderID string) error {
	rem, err := s.GetReminderByID(ctx, userID, reminderID)
	if err != nil {
		return err
	}

	rem.Status = domainreminder.ReminderStatusPaused
	rem.UpdatedAt = time.Now().UTC()

	if err := s.reminders.Update(ctx, rem); err != nil {
		return fmt.Errorf("pause reminder: %w", err)
	}

	s.cancelPendingNotifications(ctx, rem.ID)
	return nil
}

// ResumeReminder sets status=active, recalculates the next occurrence, and
// creates a new notification.
func (s *Service) ResumeReminder(ctx context.Context, userID, reminderID string) error {
	rem, err := s.GetReminderByID(ctx, userID, reminderID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	next := rem.CalculateNextRunAt(now)
	rem.Status = domainreminder.ReminderStatusActive
	rem.NextRunAt = next
	rem.UpdatedAt = now

	if err := s.reminders.Update(ctx, rem); err != nil {
		return fmt.Errorf("resume reminder: %w", err)
	}

	occ := &domainreminder.ReminderOccurrence{
		ID:          uuid.NewString(),
		ReminderID:  rem.ID,
		UserID:      userID,
		ScheduledAt: next,
		Status:      domainreminder.OccurrenceScheduled,
		CreatedAt:   now,
	}
	if err := s.occurrences.Create(ctx, occ); err != nil {
		return fmt.Errorf("create occurrence on resume: %w", err)
	}

	notif := buildNotification(userID, rem.ID, occ.ID, occ.ScheduledAt)
	_ = s.notifications.Create(ctx, notif)

	return nil
}

// ─── Complete occurrence ──────────────────────────────────────────────────────

// CompleteOccurrence marks the most recent scheduled occurrence as done. For
// recurring reminders it calculates the next run time, creates the next
// occurrence, and schedules a new notification.
func (s *Service) CompleteOccurrence(ctx context.Context, userID, reminderID string) error {
	rem, err := s.GetReminderByID(ctx, userID, reminderID)
	if err != nil {
		return err
	}

	// Find the latest scheduled occurrence for this reminder.
	occs, err := s.occurrences.FindByReminderID(ctx, rem.ID)
	if err != nil {
		return fmt.Errorf("find occurrences: %w", err)
	}

	var latest *domainreminder.ReminderOccurrence
	for _, o := range occs {
		if o.Status == domainreminder.OccurrenceScheduled {
			if latest == nil || o.ScheduledAt.After(latest.ScheduledAt) {
				latest = o
			}
		}
	}
	if latest == nil {
		return fmt.Errorf("%w: no scheduled occurrence found", ErrNotFound)
	}

	// Mark as done.
	if err := s.occurrences.UpdateStatus(ctx, latest.ID, domainreminder.OccurrenceDone); err != nil {
		return fmt.Errorf("mark occurrence done: %w", err)
	}

	// For recurring reminders, schedule the next occurrence.
	if rem.RecurrenceType != domainreminder.RecurrenceOneTime {
		now := time.Now().UTC()
		next := rem.CalculateNextRunAt(latest.ScheduledAt)
		rem.NextRunAt = next
		rem.UpdatedAt = now
		_ = s.reminders.Update(ctx, rem)

		nextOcc := &domainreminder.ReminderOccurrence{
			ID:          uuid.NewString(),
			ReminderID:  rem.ID,
			UserID:      userID,
			ScheduledAt: next,
			Status:      domainreminder.OccurrenceScheduled,
			CreatedAt:   now,
		}
		if err := s.occurrences.Create(ctx, nextOcc); err != nil {
			return fmt.Errorf("create next occurrence: %w", err)
		}

		notif := buildNotification(userID, rem.ID, nextOcc.ID, nextOcc.ScheduledAt)
		_ = s.notifications.Create(ctx, notif)
	}

	s.writeAudit(ctx, userID, "reminder", "occurrence_completed", rem.ID)
	return nil
}

// ─── CalculateNextRunAt (pure function wrapper) ───────────────────────────────

// CalculateNextRunAt is a stateless helper that delegates to the domain method.
// Handles all recurrence types; for custom_interval it reads interval_days from
// the JSONB RecurrenceRule field.
func CalculateNextRunAt(rem *domainreminder.Reminder, from time.Time) time.Time {
	return rem.CalculateNextRunAt(from)
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// cancelPendingNotifications sets all scheduled notifications for a reminder to
// "canceled". Errors are best-effort.
func (s *Service) cancelPendingNotifications(ctx context.Context, reminderID string) {
	// The NotificationRepository does not expose a bulk cancel-by-entity method
	// in TASK-005. We use a filtered query and cancel individually.
	// This will be optimised in TASK-018 with a direct SQL UPDATE.
	entityType := domainnotification.EntityReminder
	status := domainnotification.StatusScheduled
	filter := database.NotificationFilter{
		EntityType: &entityType,
		Status:     &status,
	}
	p := database.PaginationParams{Page: 1, PerPage: 100, Sort: "scheduled_at", Order: "asc"}
	result, err := s.notifications.FindAllByUser(ctx, "", filter, p)
	if err != nil {
		return
	}
	for _, n := range result.Items {
		if n.EntityID == reminderID {
			_ = s.notifications.UpdateStatus(ctx, n.ID, domainnotification.StatusCanceled)
		}
	}
}

// buildNotification constructs a Notification record for a reminder occurrence.
func buildNotification(userID, reminderID, occurrenceID string, scheduledAt time.Time) *domainnotification.Notification {
	now := time.Now().UTC()
	notifID := uuid.NewString()
	n := &domainnotification.Notification{
		ID:             notifID,
		UserID:         userID,
		EntityType:     domainnotification.EntityReminder,
		EntityID:       reminderID,
		OccurrenceID:   &occurrenceID,
		Channel:        domainnotification.ChannelWhatsApp,
		ScheduledAt:    scheduledAt,
		Status:         domainnotification.StatusScheduled,
		IdempotencyKey: fmt.Sprintf("notification:%s:channel:whatsapp", notifID),
		RetryCount:     0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return n
}

// validRecurrenceType returns true when t is one of the defined enum values.
func validRecurrenceType(t domainreminder.RecurrenceType) bool {
	switch t {
	case domainreminder.RecurrenceOneTime,
		domainreminder.RecurrenceDaily,
		domainreminder.RecurrenceWeekly,
		domainreminder.RecurrenceMonthly,
		domainreminder.RecurrenceYearly,
		domainreminder.RecurrenceCustomInterval:
		return true
	}
	return false
}

// sanitize strips leading/trailing whitespace, strips HTML-like tags, and
// truncates to maxRunes runes.
func sanitize(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	// Strip simple HTML tags (< ... >).
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
