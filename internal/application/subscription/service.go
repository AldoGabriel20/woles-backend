// Package subscription implements the Subscription application service.
package subscription

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
	domainsubscription "github.com/woles/woles-backend/internal/domain/subscription"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var (
	ErrNotFound     = errors.New("subscription not found")
	ErrPlanLimit    = errors.New("subscription limit reached for your plan")
	ErrInvalidInput = errors.New("invalid input")
)

// ─── Request / response types ─────────────────────────────────────────────────

// CreateSubscriptionRequest holds the input for creating a new subscription.
type CreateSubscriptionRequest struct {
	Name          string
	Amount        float64
	Currency      string
	BillingCycle  domainsubscription.BillingCycle
	NextBillingAt *time.Time // calculated from BillingCycle when nil
	Category      domainsubscription.SubscriptionCategory
}

// UpdateSubscriptionRequest holds the fields that may be updated.
type UpdateSubscriptionRequest struct {
	Name          *string
	Amount        *float64
	Currency      *string
	BillingCycle  *domainsubscription.BillingCycle
	NextBillingAt *time.Time
	Category      *domainsubscription.SubscriptionCategory
}

// MonthlyCostSummary is an alias for the database type returned by GetMonthlyCostSummary.
type MonthlyCostSummary = database.MonthlyCostItem

// ─── Service ──────────────────────────────────────────────────────────────────

// Service implements the subscription application service.
type Service struct {
	subscriptions database.SubscriptionRepository
	notifications database.NotificationRepository
	usageLimits   database.UsageLimitRepository
	auditLogs     database.AuditLogRepository
}

// NewService constructs the subscription service.
func NewService(
	subscriptions database.SubscriptionRepository,
	notifications database.NotificationRepository,
	usageLimits database.UsageLimitRepository,
	auditLogs database.AuditLogRepository,
) *Service {
	return &Service{
		subscriptions: subscriptions,
		notifications: notifications,
		usageLimits:   usageLimits,
		auditLogs:     auditLogs,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

// CreateSubscription validates the request, enforces plan limits, and inserts a
// new subscription together with a billing-date notification.
func (s *Service) CreateSubscription(ctx context.Context, userID string, req CreateSubscriptionRequest) (*domainsubscription.Subscription, error) {
	// Plan limit check.
	within, err := s.usageLimits.IsWithinLimit(ctx, userID, "subscriptions")
	if err != nil {
		return nil, fmt.Errorf("check usage limit: %w", err)
	}
	if !within {
		return nil, ErrPlanLimit
	}

	// Validate billing_cycle.
	if !validBillingCycle(req.BillingCycle) {
		return nil, fmt.Errorf("%w: unknown billing_cycle %q", ErrInvalidInput, req.BillingCycle)
	}

	// Sanitize name.
	name := sanitize(req.Name, 200)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	// Derive next_billing_at from billing_cycle when not explicitly provided.
	nextBillingAt := req.NextBillingAt
	if nextBillingAt == nil {
		t := calculateNextBillingAt(req.BillingCycle, time.Now().UTC())
		nextBillingAt = &t
	}

	currency := req.Currency
	if currency == "" {
		currency = "IDR"
	}

	now := time.Now().UTC()
	sub := &domainsubscription.Subscription{
		ID:            uuid.NewString(),
		UserID:        userID,
		Name:          name,
		Amount:        req.Amount,
		Currency:      currency,
		BillingCycle:  req.BillingCycle,
		NextBillingAt: *nextBillingAt,
		Category:      req.Category,
		Status:        domainsubscription.SubscriptionStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.subscriptions.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	_ = s.usageLimits.Increment(ctx, userID, "subscriptions")

	// Schedule a billing notification on the next billing date.
	s.createBillingNotification(ctx, userID, sub)

	s.writeAudit(ctx, userID, "subscription", "subscription_created", sub.ID)

	return sub, nil
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetSubscriptions returns a paginated list of subscriptions for a user.
func (s *Service) GetSubscriptions(ctx context.Context, userID string, filter database.SubscriptionFilter, page, perPage int) (*database.PaginatedResult[*domainsubscription.Subscription], error) {
	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "next_billing_at",
		Order:   "asc",
	}
	return s.subscriptions.FindAllByUser(ctx, userID, filter, p)
}

// GetSubscriptionByID returns a single subscription, enforcing ownership.
func (s *Service) GetSubscriptionByID(ctx context.Context, userID, subscriptionID string) (*domainsubscription.Subscription, error) {
	return s.ownershipCheck(ctx, userID, subscriptionID)
}

// ─── Update ───────────────────────────────────────────────────────────────────

// UpdateSubscription applies partial updates to a subscription. When
// next_billing_at changes, the old billing notification is canceled and a new
// one is created.
func (s *Service) UpdateSubscription(ctx context.Context, userID, subscriptionID string, req UpdateSubscriptionRequest) (*domainsubscription.Subscription, error) {
	sub, err := s.ownershipCheck(ctx, userID, subscriptionID)
	if err != nil {
		return nil, err
	}

	billingDateChanged := false

	if req.Name != nil {
		sub.Name = sanitize(*req.Name, 200)
	}
	if req.Amount != nil {
		sub.Amount = *req.Amount
	}
	if req.Currency != nil {
		sub.Currency = *req.Currency
	}
	if req.BillingCycle != nil {
		if !validBillingCycle(*req.BillingCycle) {
			return nil, fmt.Errorf("%w: unknown billing_cycle", ErrInvalidInput)
		}
		sub.BillingCycle = *req.BillingCycle
	}
	if req.NextBillingAt != nil {
		sub.NextBillingAt = *req.NextBillingAt
		billingDateChanged = true
	}
	if req.Category != nil {
		sub.Category = *req.Category
	}

	sub.UpdatedAt = time.Now().UTC()

	if err := s.subscriptions.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("update subscription: %w", err)
	}

	if billingDateChanged {
		s.cancelBillingNotifications(ctx, sub.ID)
		s.createBillingNotification(ctx, userID, sub)
	}

	return sub, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

// DeleteSubscription soft-deletes a subscription, cancels its pending
// notifications, and decrements the usage counter.
func (s *Service) DeleteSubscription(ctx context.Context, userID, subscriptionID string) error {
	sub, err := s.ownershipCheck(ctx, userID, subscriptionID)
	if err != nil {
		return err
	}

	if err := s.subscriptions.SoftDelete(ctx, sub.ID); err != nil {
		return fmt.Errorf("soft delete subscription: %w", err)
	}

	s.cancelBillingNotifications(ctx, sub.ID)
	_ = s.usageLimits.Decrement(ctx, userID, "subscriptions")
	s.writeAudit(ctx, userID, "subscription", "subscription_deleted", sub.ID)

	return nil
}

// ArchiveSubscription sets a subscription's status to archived and cancels its
// pending notifications.
func (s *Service) ArchiveSubscription(ctx context.Context, userID, subscriptionID string) error {
	sub, err := s.ownershipCheck(ctx, userID, subscriptionID)
	if err != nil {
		return err
	}

	sub.Status = domainsubscription.SubscriptionStatusArchived
	sub.UpdatedAt = time.Now().UTC()

	if err := s.subscriptions.Update(ctx, sub); err != nil {
		return fmt.Errorf("archive subscription: %w", err)
	}

	s.cancelBillingNotifications(ctx, sub.ID)

	return nil
}

// ─── Aggregates ───────────────────────────────────────────────────────────────

// GetMonthlyCostSummary returns the total subscription cost per currency for all
// active subscriptions belonging to the user.
func (s *Service) GetMonthlyCostSummary(ctx context.Context, userID string) ([]*MonthlyCostSummary, error) {
	items, err := s.subscriptions.GetMonthlyCostSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get monthly cost summary: %w", err)
	}
	return items, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// ownershipCheck fetches a subscription by ID and verifies it belongs to userID.
// Returns ErrNotFound (opaque) on missing or unowned records.
func (s *Service) ownershipCheck(ctx context.Context, userID, subscriptionID string) (*domainsubscription.Subscription, error) {
	sub, err := s.subscriptions.FindByID(ctx, subscriptionID)
	if err != nil || sub == nil {
		return nil, ErrNotFound
	}
	if sub.UserID != userID {
		return nil, ErrNotFound // intentionally opaque to prevent enumeration
	}
	return sub, nil
}

// createBillingNotification schedules a notification on the subscription's
// next_billing_at date.
func (s *Service) createBillingNotification(ctx context.Context, userID string, sub *domainsubscription.Subscription) {
	now := time.Now().UTC()
	if !sub.NextBillingAt.After(now) {
		return // date already in the past, skip
	}
	notifID := uuid.NewString()
	n := &domainnotification.Notification{
		ID:             notifID,
		UserID:         userID,
		EntityType:     domainnotification.EntitySubscription,
		EntityID:       sub.ID,
		Channel:        domainnotification.ChannelWhatsApp,
		ScheduledAt:    sub.NextBillingAt,
		Status:         domainnotification.StatusScheduled,
		IdempotencyKey: fmt.Sprintf("subscription:%s:billing:channel:whatsapp", sub.ID),
		RetryCount:     0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	_ = s.notifications.Create(ctx, n)
}

// cancelBillingNotifications sets all scheduled notifications for a subscription
// to "canceled". Errors are best-effort.
func (s *Service) cancelBillingNotifications(ctx context.Context, subscriptionID string) {
	entityType := domainnotification.EntitySubscription
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
		if n.EntityID == subscriptionID {
			_ = s.notifications.UpdateStatus(ctx, n.ID, domainnotification.StatusCanceled)
		}
	}
}

// calculateNextBillingAt returns the next billing date relative to now based on
// the billing cycle. For custom cycles it defaults to 1 month.
func calculateNextBillingAt(cycle domainsubscription.BillingCycle, from time.Time) time.Time {
	switch cycle {
	case domainsubscription.BillingYearly:
		return from.AddDate(1, 0, 0)
	case domainsubscription.BillingMonthly, domainsubscription.BillingCustom:
		fallthrough
	default:
		return from.AddDate(0, 1, 0)
	}
}

// validBillingCycle returns true when c is one of the defined enum values.
func validBillingCycle(c domainsubscription.BillingCycle) bool {
	switch c {
	case domainsubscription.BillingMonthly,
		domainsubscription.BillingYearly,
		domainsubscription.BillingCustom:
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
