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
	ID               string        `json:"id"`
	Email            *string       `json:"email"`
	Phone            *string       `json:"phone"`
	PasswordHash     *string       `json:"-"`
	Name             *string       `json:"name"`
	AvatarURL        *string       `json:"avatar_url"`
	Timezone         string        `json:"timezone"`
	Plan             Plan          `json:"plan"`
	AccountStatus    AccountStatus `json:"account_status"`
	FailedLoginCount int           `json:"-"`
	LockedUntil      *time.Time    `json:"-"`
	TOTPSecret       *string       `json:"-"`
	TOTPEnabled      bool          `json:"totp_enabled"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
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
	ID                string                 `json:"id"`
	UserID            string                 `json:"user_id"`
	Phone             string                 `json:"phone"`
	Provider          string                 `json:"provider"`
	ProviderContactID *string                `json:"provider_contact_id,omitempty"`
	Status            WhatsAppIdentityStatus `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
}
