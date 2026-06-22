// Package cache defines outbound port interfaces for the cache layer (Redis).
// No implementations live here — only contracts.
package cache

import (
	"context"
	"time"
)

// RateLimiter implements a sliding-window rate limit check.
// Returns (allowed bool, remaining int, error).
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error)
}

// IdempotencyStore provides a one-shot set/check for idempotency keys.
type IdempotencyStore interface {
	// Set stores the key with the given TTL. Returns ErrAlreadyProcessed if
	// the key already exists (NX semantics).
	Set(ctx context.Context, key string, ttl time.Duration) error
	// Exists returns true when the key is present in the store.
	Exists(ctx context.Context, key string) (bool, error)
}

// ErrAlreadyProcessed is returned by IdempotencyStore.Set when the key exists.
var ErrAlreadyProcessed = errAlreadyProcessed("already processed")

type errAlreadyProcessed string

func (e errAlreadyProcessed) Error() string { return string(e) }

// OTPStore stores hashed one-time passwords with a short TTL.
type OTPStore interface {
	// Set stores hashedOTP under key otp:{phone} with the given TTL.
	Set(ctx context.Context, phone, hashedOTP string, ttl time.Duration) error
	// Get retrieves the stored hash for the phone number.
	Get(ctx context.Context, phone string) (string, error)
	// Delete removes the OTP entry (call after successful verification).
	Delete(ctx context.Context, phone string) error
}

// SessionTokenCache provides generic key/value storage for short-lived session tokens.
type SessionTokenCache interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}
