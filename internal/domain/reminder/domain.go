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
	CategoryBill         ReminderCategory = "bill"
	CategoryHealth       ReminderCategory = "health"
	CategoryVehicle      ReminderCategory = "vehicle"
	CategoryInsurance    ReminderCategory = "insurance"
	CategorySubscription ReminderCategory = "subscription"
	CategoryTax          ReminderCategory = "tax"
	CategoryPersonal     ReminderCategory = "personal"
	CategoryWork         ReminderCategory = "work"
	CategoryFamily       ReminderCategory = "family"
	CategoryDocument     ReminderCategory = "document"
	CategoryCustom       ReminderCategory = "custom"
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
	ID             string           `json:"id"`
	UserID         string           `json:"user_id"`
	Title          string           `json:"title"`
	Category       ReminderCategory `json:"category"`
	RecurrenceType RecurrenceType   `json:"recurrence_type"`
	RecurrenceRule json.RawMessage  `json:"recurrence_rule,omitempty"`
	NextRunAt      time.Time        `json:"next_run_at"`
	Timezone       string           `json:"timezone"`
	Status         ReminderStatus   `json:"status"`
	Source         ReminderSource   `json:"source"`
	OriginalText   *string          `json:"original_text,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
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
	ID          string           `json:"id"`
	ReminderID  string           `json:"reminder_id"`
	UserID      string           `json:"user_id"`
	ScheduledAt time.Time        `json:"scheduled_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Status      OccurrenceStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
}
