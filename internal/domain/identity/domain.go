package identity

import "time"

// AccountStatus represents the state of a user account.
type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusDeleted   AccountStatus = "deleted"
)

// Plan represents the subscription plan of a user.
type Plan string

const (
	PlanFree     Plan = "free"
	PlanPremium  Plan = "premium"
	PlanAdvanced Plan = "advanced"
)

// User is the core identity entity. It must not import any infrastructure package.
type User struct {
	ID               string
	Email            *string
	Phone            *string
	PasswordHash     *string
	Name             *string
	AvatarURL        *string
	Timezone         string
	Plan             Plan
	AccountStatus    AccountStatus
	FailedLoginCount int
	LockedUntil      *time.Time
	TOTPSecret       *string
	TOTPEnabled      bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsLocked returns true when the account has an active lockout window.
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// CanLogin returns true when the account is active and not locked.
func (u *User) CanLogin() bool {
	return u.AccountStatus == AccountStatusActive && !u.IsLocked()
}

// WhatsAppIdentityStatus represents the connection state of a WhatsApp identity.
type WhatsAppIdentityStatus string

const (
	WhatsAppStatusActive       WhatsAppIdentityStatus = "active"
	WhatsAppStatusBlocked      WhatsAppIdentityStatus = "blocked"
	WhatsAppStatusDisconnected WhatsAppIdentityStatus = "disconnected"
)

// WhatsAppIdentity links a WhatsApp phone number to a user account.
type WhatsAppIdentity struct {
	ID                string
	UserID            string
	Phone             string
	Provider          string
	ProviderContactID *string
	Status            WhatsAppIdentityStatus
	CreatedAt         time.Time
}
