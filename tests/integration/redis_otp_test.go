package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/woles/woles-backend/internal/adapter/outbound/redis"
)

func TestRedis_OTPStore_SetGetDelete(t *testing.T) {
	client := skipIfNoRedis(t)
	store := redis.NewOTPStore(client)
	ctx := context.Background()

	phone := "628111222333"
	hashed := "hashedOTPvalue"
	ttl := 5 * time.Minute

	// Set
	if err := store.Set(ctx, phone, hashed, ttl); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Get
	got, err := store.Get(ctx, phone)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != hashed {
		t.Errorf("Get: want %q, got %q", hashed, got)
	}

	// Delete
	if err := store.Delete(ctx, phone); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = store.Get(ctx, phone)
	if err == nil {
		t.Error("Get after Delete: expected error (key not found), got nil")
	}
}

func TestRedis_OTPStore_ExpiredAfterTTL(t *testing.T) {
	client := skipIfNoRedis(t)
	store := redis.NewOTPStore(client)
	ctx := context.Background()

	phone := "628999000111"
	// TTL of 1 second so the test completes quickly.
	if err := store.Set(ctx, phone, "myotp", 1*time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Wait for expiry.
	time.Sleep(1500 * time.Millisecond)

	_, err := store.Get(ctx, phone)
	if err == nil {
		t.Error("Get after TTL expiry: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "expired") {
		t.Logf("error: %v (acceptable)", err)
	}
}
