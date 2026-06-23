package notification

import (
	"fmt"
	"time"
)

// NotificationStatus represents the delivery state of a notification.
type NotificationStatus string

const (
	StatusScheduled NotificationStatus = "scheduled"
	StatusSending   NotificationStatus = "sending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusCanceled  NotificationStatus = "canceled"
)

// NotificationChannel represents the delivery channel.
type NotificationChannel string

const (
	ChannelWhatsApp NotificationChannel = "whatsapp"
	ChannelEmail    NotificationChannel = "email"
	ChannelWebPush  NotificationChannel = "web_push"
)

// NotificationEntityType indicates which domain entity triggered the notification.
type NotificationEntityType string

const (
	EntityReminder     NotificationEntityType = "reminder"
	EntityDocument     NotificationEntityType = "document"
	EntitySubscription NotificationEntityType = "subscription"
	EntityGoal         NotificationEntityType = "goal"
)

// Notification is the core notification delivery entity.
type Notification struct {
	ID                string                 `json:"id"`
	UserID            string                 `json:"user_id"`
	EntityType        NotificationEntityType `json:"entity_type"`
	EntityID          string                 `json:"entity_id"`
	OccurrenceID      *string                `json:"occurrence_id,omitempty"`
	Channel           NotificationChannel    `json:"channel"`
	ScheduledAt       time.Time              `json:"scheduled_at"`
	SentAt            *time.Time             `json:"sent_at,omitempty"`
	Status            NotificationStatus     `json:"status"`
	IdempotencyKey    string                 `json:"idempotency_key"`
	ProviderMessageID *string                `json:"provider_message_id,omitempty"`
	FailureReason     *string                `json:"failure_reason,omitempty"`
	RetryCount        int                    `json:"retry_count"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// BuildIdempotencyKey returns the canonical idempotency key for this notification.
// Format: notification:{id}:channel:{channel}
func (n *Notification) BuildIdempotencyKey() string {
	return fmt.Sprintf("notification:%s:channel:%s", n.ID, n.Channel)
}
