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
	ID                string
	UserID            string
	EntityType        NotificationEntityType
	EntityID          string
	OccurrenceID      *string
	Channel           NotificationChannel
	ScheduledAt       time.Time
	SentAt            *time.Time
	Status            NotificationStatus
	IdempotencyKey    string
	ProviderMessageID *string
	FailureReason     *string
	RetryCount        int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// BuildIdempotencyKey returns the canonical idempotency key for this notification.
// Format: notification:{id}:channel:{channel}
func (n *Notification) BuildIdempotencyKey() string {
	return fmt.Sprintf("notification:%s:channel:%s", n.ID, n.Channel)
}
