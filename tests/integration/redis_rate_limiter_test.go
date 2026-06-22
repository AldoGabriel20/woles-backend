package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/woles/woles-backend/internal/adapter/outbound/redis"
)

func skipIfNoRedis(t *testing.T) *redis.Client {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set — skipping integration test")
	}
	os.Setenv("REDIS_URL", url)
	client, err := redis.New(context.Background())
	if err != nil {
		t.Fatalf("connect to test Redis: %v", err)
	}
	return client
}

func TestRedis_RateLimiter_10AllowedThen11thBlocked(t *testing.T) {
	client := skipIfNoRedis(t)
	limiter := redis.NewRateLimiter(client)
	ctx := context.Background()

	key := "rl:test-user:test-endpoint:" + t.Name()
	limit := 10
	window := time.Minute

	for i := 1; i <= limit; i++ {
		allowed, _, err := limiter.Allow(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if !allowed {
			t.Errorf("request %d should be allowed", i)
		}
	}

	allowed, _, err := limiter.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow 11: %v", err)
	}
	if allowed {
		t.Error("11th request within window should be blocked")
	}
}
