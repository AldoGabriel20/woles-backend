package reminder

import (
	"encoding/json"
	"time"
)

// RecurrenceType defines how often a reminder repeats.
type RecurrenceType string

const (
	RecurrenceOneTime        RecurrenceType = "one_time"
	RecurrenceDaily          RecurrenceType = "daily"
	RecurrenceWeekly         RecurrenceType = "weekly"
	RecurrenceMonthly        RecurrenceType = "monthly"
	RecurrenceYearly         RecurrenceType = "yearly"
	RecurrenceCustomInterval RecurrenceType = "custom_interval"
)

// ReminderStatus represents the lifecycle state of a reminder.
type ReminderStatus string

const (
	ReminderStatusActive   ReminderStatus = "active"
	ReminderStatusPaused   ReminderStatus = "paused"
	ReminderStatusArchived ReminderStatus = "archived"
)

// ReminderCategory classifies a reminder.
type ReminderCategory string

const (
	CategoryBill     ReminderCategory = "bill"
	CategoryVehicle  ReminderCategory = "vehicle"
	CategoryDocument ReminderCategory = "document"
	CategoryCustom   ReminderCategory = "custom"
)

// ReminderSource indicates how the reminder was created.
type ReminderSource string

const (
	SourceWhatsApp ReminderSource = "whatsapp"
	SourceWeb      ReminderSource = "web"
	SourceSystem   ReminderSource = "system"
)

// recurrenceRule is used internally to decode JSONB fields.
type recurrenceRule struct {
	IntervalDays *int `json:"interval_days"`
}

// Reminder is the core reminder entity.
type Reminder struct {
	ID             string
	UserID         string
	Title          string
	Category       ReminderCategory
	RecurrenceType RecurrenceType
	RecurrenceRule json.RawMessage
	NextRunAt      time.Time
	Timezone       string
	Status         ReminderStatus
	Source         ReminderSource
	OriginalText   *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CalculateNextRunAt computes the next scheduled time after `from` based on the
// reminder's recurrence type. For custom_interval the interval_days value is
// read from RecurrenceRule JSONB.
func (r *Reminder) CalculateNextRunAt(from time.Time) time.Time {
	switch r.RecurrenceType {
	case RecurrenceOneTime:
		// One-time reminders do not recur; return the original run time.
		return r.NextRunAt

	case RecurrenceDaily:
		return from.AddDate(0, 0, 1)

	case RecurrenceWeekly:
		return from.AddDate(0, 0, 7)

	case RecurrenceMonthly:
		return from.AddDate(0, 1, 0)

	case RecurrenceYearly:
		return from.AddDate(1, 0, 0)

	case RecurrenceCustomInterval:
		if len(r.RecurrenceRule) > 0 {
			var rule recurrenceRule
			if err := json.Unmarshal(r.RecurrenceRule, &rule); err == nil && rule.IntervalDays != nil && *rule.IntervalDays > 0 {
				return from.AddDate(0, 0, *rule.IntervalDays)
			}
		}
		// Fallback: treat as one-time.
		return r.NextRunAt

	default:
		return r.NextRunAt
	}
}

// OccurrenceStatus represents the state of a single reminder occurrence.
type OccurrenceStatus string

const (
	OccurrenceScheduled OccurrenceStatus = "scheduled"
	OccurrenceSent      OccurrenceStatus = "sent"
	OccurrenceDone      OccurrenceStatus = "done"
	OccurrenceSkipped   OccurrenceStatus = "skipped"
	OccurrenceFailed    OccurrenceStatus = "failed"
)

// ReminderOccurrence represents one scheduled firing of a reminder.
type ReminderOccurrence struct {
	ID          string
	ReminderID  string
	UserID      string
	ScheduledAt time.Time
	CompletedAt *time.Time
	Status      OccurrenceStatus
	CreatedAt   time.Time
}
