// Package database defines all outbound repository port interfaces.
// No implementations live here — only contracts.
package database

import (
	"context"
	"io"
	"time"

	"github.com/woles/woles-backend/internal/domain/billing"
	"github.com/woles/woles-backend/internal/domain/chat"
	"github.com/woles/woles-backend/internal/domain/document"
	"github.com/woles/woles-backend/internal/domain/family"
	"github.com/woles/woles-backend/internal/domain/goal"
	"github.com/woles/woles-backend/internal/domain/identity"
	"github.com/woles/woles-backend/internal/domain/notification"
	"github.com/woles/woles-backend/internal/domain/reminder"
	"github.com/woles/woles-backend/internal/domain/subscription"
)

// ─── Shared pagination / filter types ────────────────────────────────────────

// PaginationParams carries offset-pagination and optional sorting instructions.
type PaginationParams struct {
	Page    int
	PerPage int
	Sort    string // column name — must be validated against an allowlist before use in SQL
	Order   string // "asc" or "desc"
}

// PaginatedResult is a generic page of results with metadata.
type PaginatedResult[T any] struct {
	Items      []T
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

// AuditLog is a minimal view of an audit event used by the repository layer.
type AuditLog struct {
	ID         string
	UserID     *string
	ActorType  string
	Action     string
	EntityType *string
	EntityID   *string
	IPAddress  *string
	UserAgent  *string
	CreatedAt  time.Time
}

// RefreshToken is the database representation of a refresh token record.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	FamilyID  string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// UserSession is the database representation of an active user session.
type UserSession struct {
	ID             string
	UserID         string
	RefreshTokenID string
	DeviceName     *string
	IPAddress      *string
	UserAgent      *string
	LastActiveAt   time.Time
	CreatedAt      time.Time
}

// InboundMessage is the database representation of an inbound WhatsApp message.
type InboundMessage struct {
	ID                string
	UserID            *string
	Channel           string
	ProviderMessageID string
	FromPhone         string
	RawText           string
	ParsedIntent      []byte // JSONB
	ProcessingStatus  string
	CreatedAt         time.Time
}

// StorageUsage is returned by DocumentRepository.GetStorageUsage.
type StorageUsage struct {
	UsedBytes int64
	UsedMB    float64
	UsedGB    float64
	FileCount int
}

// VaultCategoryCount holds the document count for one vault category.
type VaultCategoryCount struct {
	Category document.VaultCategory
	Count    int
}

// VaultHealth is returned by DocumentRepository.GetVaultHealth.
type VaultHealth struct {
	Categories        []VaultCategoryCount
	CompletenessScore int // 0–100
}

// MonthlyCostItem is one row in a monthly cost summary.
type MonthlyCostItem struct {
	Currency          string
	TotalAmount       float64
	SubscriptionCount int
}

// NotificationStats is returned by NotificationRepository.GetStats.
type NotificationStats struct {
	TotalSent       int
	TotalFailed     int
	DeliveryRatePct float64
	TopCategory     string
}

// TimelineItem is the normalised result returned by TimelineRepository.
type TimelineItem struct {
	ID       string
	Type     string // "reminder", "document", "subscription", "goal"
	Title    string
	DueAt    time.Time
	Status   string
	EntityID string
}

// ReminderFilter holds optional filters for listing reminders.
type ReminderFilter struct {
	Status   *reminder.ReminderStatus
	Category *reminder.ReminderCategory
	Search   *string
}

// DocumentFilter holds optional filters for listing documents.
type DocumentFilter struct {
	VaultCategory    *document.VaultCategory
	ExpiryWithinDays *int
	Search           *string
}

// SubscriptionFilter holds optional filters for listing subscriptions.
type SubscriptionFilter struct {
	Status   *subscription.SubscriptionStatus
	Category *subscription.SubscriptionCategory
}

// GoalFilter holds optional filters for listing goals.
type GoalFilter struct {
	Status *goal.GoalStatus
}

// NotificationFilter holds optional filters for listing notifications.
type NotificationFilter struct {
	EntityType *notification.NotificationEntityType
	Status     *notification.NotificationStatus
	From       *time.Time
	To         *time.Time
}

// ─── Repository interfaces ────────────────────────────────────────────────────

// UserRepository manages persistence for the identity.User entity.
type UserRepository interface {
	Create(ctx context.Context, u *identity.User) error
	FindByID(ctx context.Context, id string) (*identity.User, error)
	FindByEmail(ctx context.Context, email string) (*identity.User, error)
	FindByPhone(ctx context.Context, phone string) (*identity.User, error)
	Update(ctx context.Context, u *identity.User) error
	UpdateFailedLoginCount(ctx context.Context, id string, count int) error
	UpdateLockedUntil(ctx context.Context, id string, until *time.Time) error
	UpdateAvatarURL(ctx context.Context, id string, avatarURL string) error
	UpdatePlan(ctx context.Context, id string, plan identity.Plan) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
}

// RefreshTokenRepository manages persistence for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, t *RefreshToken) error
	FindByID(ctx context.Context, id string) (*RefreshToken, error)
	FindByFamilyID(ctx context.Context, familyID string) ([]*RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	RevokeFamily(ctx context.Context, familyID string) error
}

// UserSessionRepository manages active authenticated sessions.
type UserSessionRepository interface {
	Create(ctx context.Context, s *UserSession) error
	FindAllByUser(ctx context.Context, userID string) ([]*UserSession, error)
	FindByID(ctx context.Context, id string) (*UserSession, error)
	UpdateLastActive(ctx context.Context, id string, at time.Time) error
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

// ReminderRepository manages persistence for reminder.Reminder.
type ReminderRepository interface {
	Create(ctx context.Context, r *reminder.Reminder) error
	FindByID(ctx context.Context, id string) (*reminder.Reminder, error)
	FindAllByUser(ctx context.Context, userID string, filter ReminderFilter, p PaginationParams) (*PaginatedResult[*reminder.Reminder], error)
	Update(ctx context.Context, r *reminder.Reminder) error
	SoftDelete(ctx context.Context, id string) error
	FindDueOccurrences(ctx context.Context, before time.Time, limit int) ([]*reminder.ReminderOccurrence, error)
}

// ReminderOccurrenceRepository manages reminder occurrence records.
type ReminderOccurrenceRepository interface {
	Create(ctx context.Context, o *reminder.ReminderOccurrence) error
	FindByReminderID(ctx context.Context, reminderID string) ([]*reminder.ReminderOccurrence, error)
	UpdateStatus(ctx context.Context, id string, status reminder.OccurrenceStatus) error
	ClaimForSending(ctx context.Context, batchSize int) ([]*reminder.ReminderOccurrence, error) // FOR UPDATE SKIP LOCKED
}

// DocumentRepository manages persistence for document.Document.
type DocumentRepository interface {
	Create(ctx context.Context, d *document.Document) error
	FindByID(ctx context.Context, id string) (*document.Document, error)
	FindAllByUser(ctx context.Context, userID string, filter DocumentFilter, p PaginationParams) (*PaginatedResult[*document.Document], error)
	Update(ctx context.Context, d *document.Document) error
	SoftDelete(ctx context.Context, id string) error
	FindExpiringSoon(ctx context.Context, userID string, withinDays int) ([]*document.Document, error)
	GetStorageUsage(ctx context.Context, userID string) (*StorageUsage, error)
	GetVaultHealth(ctx context.Context, userID string) (*VaultHealth, error)
}

// SubscriptionRepository manages persistence for subscription.Subscription.
type SubscriptionRepository interface {
	Create(ctx context.Context, s *subscription.Subscription) error
	FindByID(ctx context.Context, id string) (*subscription.Subscription, error)
	FindAllByUser(ctx context.Context, userID string, filter SubscriptionFilter, p PaginationParams) (*PaginatedResult[*subscription.Subscription], error)
	Update(ctx context.Context, s *subscription.Subscription) error
	SoftDelete(ctx context.Context, id string) error
	GetMonthlyCostSummary(ctx context.Context, userID string) ([]*MonthlyCostItem, error)
}

// GoalRepository manages persistence for goal.Goal.
type GoalRepository interface {
	Create(ctx context.Context, g *goal.Goal) error
	FindByID(ctx context.Context, id string) (*goal.Goal, error)
	FindAllByUser(ctx context.Context, userID string, filter GoalFilter, p PaginationParams) (*PaginatedResult[*goal.Goal], error)
	FindActiveGoal(ctx context.Context, userID string) (*goal.Goal, error)
	UpdateProgress(ctx context.Context, id string, currentAmount float64) error
	Update(ctx context.Context, g *goal.Goal) error
	SoftDelete(ctx context.Context, id string) error
}

// NotificationRepository manages persistence for notification.Notification.
type NotificationRepository interface {
	Create(ctx context.Context, n *notification.Notification) error
	FindByID(ctx context.Context, id string) (*notification.Notification, error)
	FindAllByUser(ctx context.Context, userID string, filter NotificationFilter, p PaginationParams) (*PaginatedResult[*notification.Notification], error)
	ClaimDue(ctx context.Context, batchSize int) ([]*notification.Notification, error) // FOR UPDATE SKIP LOCKED
	UpdateStatus(ctx context.Context, id string, status notification.NotificationStatus) error
	IncrementRetry(ctx context.Context, id string) error
	GetStats(ctx context.Context, userID string) (*NotificationStats, error)
	ExportRange(ctx context.Context, userID string, from, to time.Time) ([]*notification.Notification, error)
}

// FamilyMemberRepository manages persistence for family.FamilyMember.
type FamilyMemberRepository interface {
	Create(ctx context.Context, m *family.FamilyMember) error
	FindByID(ctx context.Context, id string) (*family.FamilyMember, error)
	FindAllByOwner(ctx context.Context, ownerUserID string) ([]*family.FamilyMember, error)
	Update(ctx context.Context, m *family.FamilyMember) error
	Delete(ctx context.Context, id string) error
}

// ChatMessageRepository manages persistence for chat.ChatMessage.
type ChatMessageRepository interface {
	Create(ctx context.Context, m *chat.ChatMessage) error
	FindAllByUser(ctx context.Context, userID string, p PaginationParams) (*PaginatedResult[*chat.ChatMessage], error)
	DeleteAllByUser(ctx context.Context, userID string) error
}

// ChatUsageRepository manages monthly chat usage counters.
type ChatUsageRepository interface {
	GetOrCreate(ctx context.Context, userID string, month time.Time) (*chat.ChatUsage, error)
	Increment(ctx context.Context, userID string, month time.Time) error
	GetQuota(ctx context.Context, userID string, month time.Time) (*chat.ChatUsage, error)
}

// InboundMessageRepository manages persistence for inbound WhatsApp messages.
type InboundMessageRepository interface {
	Create(ctx context.Context, m *InboundMessage) error
	FindByProviderMessageID(ctx context.Context, providerMessageID string) (*InboundMessage, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

// AuditLogRepository provides append-only audit logging.
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	FindAllByUser(ctx context.Context, userID string, p PaginationParams) (*PaginatedResult[*AuditLog], error)
}

// UsageLimitRepository tracks per-user resource consumption counters.
type UsageLimitRepository interface {
	Get(ctx context.Context, userID string) (*billing.UsageLimit, error)
	Increment(ctx context.Context, userID string, resource string) error
	Decrement(ctx context.Context, userID string, resource string) error
	IsWithinLimit(ctx context.Context, userID string, resource string) (bool, error)
}

// TimelineRepository queries the aggregated timeline view across multiple tables.
type TimelineRepository interface {
	GetTimelineItems(ctx context.Context, userID string, from, to time.Time, p PaginationParams) (*PaginatedResult[*TimelineItem], error)
}

// Ensure io is used (via UploadDocumentFile in the application layer which passes io.Reader).
var _ io.Reader = (*dummyReader)(nil)

type dummyReader struct{}

func (*dummyReader) Read(_ []byte) (int, error) { return 0, io.EOF }
