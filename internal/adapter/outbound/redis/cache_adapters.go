package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	portcache "github.com/woles/woles-backend/internal/port/outbound/cache"
)

// IdempotencyStoreAdapter implements cache.IdempotencyStore.
type IdempotencyStoreAdapter struct {
	rdb *goredis.Client
}

// NewIdempotencyStore creates a new IdempotencyStoreAdapter.
func NewIdempotencyStore(c *Client) *IdempotencyStoreAdapter {
	return &IdempotencyStoreAdapter{rdb: c.rdb}
}

// Set stores the key with NX semantics (only if not already present).
// Returns cache.ErrAlreadyProcessed when the key already exists.
func (s *IdempotencyStoreAdapter) Set(ctx context.Context, key string, ttl time.Duration) error {
	ok, err := s.rdb.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return fmt.Errorf("idempotency: set: %w", err)
	}
	if !ok {
		return portcache.ErrAlreadyProcessed
	}
	return nil
}

// Exists returns true when the key is present in Redis.
func (s *IdempotencyStoreAdapter) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("idempotency: exists: %w", err)
	}
	return n > 0, nil
}

// ─── OTPStore ─────────────────────────────────────────────────────────────────

// OTPStoreAdapter implements cache.OTPStore.
// Keys are stored under otp:{phone}.
type OTPStoreAdapter struct {
	rdb *goredis.Client
}

// NewOTPStore creates a new OTPStoreAdapter.
func NewOTPStore(c *Client) *OTPStoreAdapter {
	return &OTPStoreAdapter{rdb: c.rdb}
}

func otpKey(phone string) string { return "otp:" + phone }

// Set stores hashedOTP under otp:{phone} with the given TTL (typically 5 min = 300s).
func (s *OTPStoreAdapter) Set(ctx context.Context, phone, hashedOTP string, ttl time.Duration) error {
	if err := s.rdb.Set(ctx, otpKey(phone), hashedOTP, ttl).Err(); err != nil {
		return fmt.Errorf("otp: set: %w", err)
	}
	return nil
}

// Get retrieves the stored hash for the phone. Returns an empty string and
// a wrapped goredis.Nil error when the key has expired or does not exist.
func (s *OTPStoreAdapter) Get(ctx context.Context, phone string) (string, error) {
	val, err := s.rdb.Get(ctx, otpKey(phone)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", fmt.Errorf("otp: not found or expired")
	}
	if err != nil {
		return "", fmt.Errorf("otp: get: %w", err)
	}
	return val, nil
}

// Delete removes the OTP entry (called after successful verification).
func (s *OTPStoreAdapter) Delete(ctx context.Context, phone string) error {
	if err := s.rdb.Del(ctx, otpKey(phone)).Err(); err != nil {
		return fmt.Errorf("otp: delete: %w", err)
	}
	return nil
}

// ─── SessionTokenCache ────────────────────────────────────────────────────────

// SessionTokenCacheAdapter implements cache.SessionTokenCache.
type SessionTokenCacheAdapter struct {
	rdb *goredis.Client
}

// NewSessionTokenCache creates a new SessionTokenCacheAdapter.
func NewSessionTokenCache(c *Client) *SessionTokenCacheAdapter {
	return &SessionTokenCacheAdapter{rdb: c.rdb}
}

// Set stores a string value under key with the given TTL.
func (s *SessionTokenCacheAdapter) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := s.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("session_cache: set: %w", err)
	}
	return nil
}

// Get retrieves the value stored under key.
func (s *SessionTokenCacheAdapter) Get(ctx context.Context, key string) (string, error) {
	val, err := s.rdb.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return "", fmt.Errorf("session_cache: key not found: %s", key)
	}
	if err != nil {
		return "", fmt.Errorf("session_cache: get: %w", err)
	}
	return val, nil
}

// Delete removes the key.
func (s *SessionTokenCacheAdapter) Delete(ctx context.Context, key string) error {
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("session_cache: delete: %w", err)
	}
	return nil
}
