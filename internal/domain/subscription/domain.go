package subscription

import "time"

// BillingCycle defines how often a subscription is billed.
type BillingCycle string

const (
	BillingMonthly BillingCycle = "monthly"
	BillingYearly  BillingCycle = "yearly"
	BillingCustom  BillingCycle = "custom"
)

// SubscriptionStatus represents the lifecycle state of a subscription.
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusArchived SubscriptionStatus = "archived"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
)

// SubscriptionCategory classifies what a subscription is for.
type SubscriptionCategory string

const (
	CategoryEntertainment SubscriptionCategory = "entertainment"
	CategoryProductivity  SubscriptionCategory = "productivity"
	CategoryBill          SubscriptionCategory = "bill"
	CategoryOther         SubscriptionCategory = "other"
)

// Subscription is the core subscription tracking entity.
type Subscription struct {
	ID            string
	UserID        string
	Name          string
	Amount        float64
	Currency      string
	BillingCycle  BillingCycle
	NextBillingAt time.Time
	Category      SubscriptionCategory
	Status        SubscriptionStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
