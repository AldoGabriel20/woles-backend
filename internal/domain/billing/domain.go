package billing

import "time"

// Plan holds the feature flags and limits for a subscription tier.
type Plan struct {
	ID                string
	Name              string
	PriceIDR          float64
	ReminderLimit     int // -1 = unlimited
	DocumentLimit     int
	SubscriptionLimit int
	GoalTracker       bool
	Timeline          bool
	FamilyAccount     bool
	AIChat            bool
	AIChatQuota       int // -1 = unlimited, 0 = none
	OCR               bool
	CreatedAt         time.Time
}

// UsageLimit tracks per-user resource consumption counters.
type UsageLimit struct {
	ID                string
	UserID            string
	RemindersUsed     int
	DocumentsUsed     int
	SubscriptionsUsed int
	UpdatedAt         time.Time
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
