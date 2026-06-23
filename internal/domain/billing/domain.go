package billing

import "time"

// Plan holds the feature flags and limits for a subscription tier.
type Plan struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	PriceIDR          float64   `json:"price_idr"`
	ReminderLimit     int       `json:"reminder_limit"`
	DocumentLimit     int       `json:"document_limit"`
	SubscriptionLimit int       `json:"subscription_limit"`
	GoalTracker       bool      `json:"goal_tracker"`
	Timeline          bool      `json:"timeline"`
	FamilyAccount     bool      `json:"family_account"`
	AIChat            bool      `json:"ai_chat"`
	AIChatQuota       int       `json:"ai_chat_quota"`
	OCR               bool      `json:"ocr"`
	CreatedAt         time.Time `json:"created_at"`
}

// UsageLimit tracks per-user resource consumption counters.
type UsageLimit struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	RemindersUsed     int       `json:"reminders_used"`
	DocumentsUsed     int       `json:"documents_used"`
	SubscriptionsUsed int       `json:"subscriptions_used"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// IsWithinLimit returns true when the current usage for resource is below the
// plan's limit.
//
// resource must be one of "reminders", "documents", or "subscriptions".
// A limit of -1 means unlimited (always returns true).
func (p *Plan) IsWithinLimit(resource string, current int) bool {
	var limit int
	switch resource {
	case "reminders":
		limit = p.ReminderLimit
	case "documents":
		limit = p.DocumentLimit
	case "subscriptions":
		limit = p.SubscriptionLimit
	default:
		return true // unknown resource — allow by default
	}
	if limit < 0 {
		return true // unlimited
	}
	return current < limit
}
