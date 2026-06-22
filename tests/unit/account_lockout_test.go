package unit_test

import (
	"testing"
	"time"

	"github.com/woles/woles-backend/internal/domain/identity"
)

func makeLockedUser(lockedUntil time.Time) *identity.User {
	return &identity.User{
		AccountStatus: identity.AccountStatusActive,
		LockedUntil:   &lockedUntil,
	}
}

// ─── IsLocked ─────────────────────────────────────────────────────────────────

func TestIsLocked_NoLockout(t *testing.T) {
	u := &identity.User{AccountStatus: identity.AccountStatusActive}
	if u.IsLocked() {
		t.Error("user without LockedUntil should not be locked")
	}
}

func TestIsLocked_LockedInFuture(t *testing.T) {
	u := makeLockedUser(time.Now().Add(10 * time.Minute))
	if !u.IsLocked() {
		t.Error("user with future LockedUntil should be locked")
	}
}

func TestIsLocked_LockExpired(t *testing.T) {
	u := makeLockedUser(time.Now().Add(-1 * time.Minute))
	if u.IsLocked() {
		t.Error("user whose lock expired should NOT be locked")
	}
}

// ─── CanLogin ─────────────────────────────────────────────────────────────────

func TestCanLogin_ActiveUnlocked(t *testing.T) {
	u := &identity.User{AccountStatus: identity.AccountStatusActive}
	if !u.CanLogin() {
		t.Error("active unlocked user should be able to login")
	}
}

func TestCanLogin_LockedAfter5Failures(t *testing.T) {
	// Simulate lockout for 15 minutes after 5 failures.
	lockUntil := time.Now().Add(15 * time.Minute)
	u := &identity.User{
		AccountStatus:    identity.AccountStatusActive,
		FailedLoginCount: 5,
		LockedUntil:      &lockUntil,
	}
	if u.CanLogin() {
		t.Error("locked account should not be able to login")
	}
}

func TestCanLogin_UnlockedAfter15Minutes(t *testing.T) {
	// Lock expired 1 second ago.
	lockUntil := time.Now().Add(-1 * time.Second)
	u := &identity.User{
		AccountStatus:    identity.AccountStatusActive,
		FailedLoginCount: 5,
		LockedUntil:      &lockUntil,
	}
	if !u.CanLogin() {
		t.Error("user whose lock expired should be able to login")
	}
}

func TestCanLogin_SuspendedAccount(t *testing.T) {
	u := &identity.User{AccountStatus: identity.AccountStatusSuspended}
	if u.CanLogin() {
		t.Error("suspended account should not be able to login")
	}
}

func TestCanLogin_DeletedAccount(t *testing.T) {
	u := &identity.User{AccountStatus: identity.AccountStatusDeleted}
	if u.CanLogin() {
		t.Error("deleted account should not be able to login")
	}
}
