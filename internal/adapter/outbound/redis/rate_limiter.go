package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// RateLimiterAdapter implements cache.RateLimiter using a Redis sorted-set
// sliding window executed atomically in a Lua script.
//
// Key format: rl:{identifier}:{endpoint}
// Each member is a unique request timestamp (nanoseconds) — the score is also
// the timestamp so old members outside the window can be pruned with
// ZREMRANGEBYSCORE.
type RateLimiterAdapter struct {
	rdb *goredis.Client
}

// NewRateLimiter creates a new RateLimiterAdapter.
func NewRateLimiter(c *Client) *RateLimiterAdapter {
	return &RateLimiterAdapter{rdb: c.rdb}
}

// slidingWindowScript atomically:
//  1. Removes members whose score (timestamp ns) is older than the window.
//  2. Counts remaining members.
//  3. If count < limit: adds the current request timestamp, refreshes TTL.
//  4. Returns {allowed (0/1), current_count_after_add}.
var slidingWindowScript = goredis.NewScript(`
local key     = KEYS[1]
local now     = tonumber(ARGV[1])
local window  = tonumber(ARGV[2])
local limit   = tonumber(ARGV[3])
local ttl_ms  = tonumber(ARGV[4])
local min_ts  = now - window

redis.call('ZREMRANGEBYSCORE', key, '-inf', min_ts)
local count = redis.call('ZCARD', key)

if count < limit then
  redis.call('ZADD', key, now, now)
  redis.call('PEXPIRE', key, ttl_ms)
  return {1, count + 1}
else
  redis.call('PEXPIRE', key, ttl_ms)
  return {0, count}
end
`)

// Allow checks and records one request against the sliding window.
// Returns (allowed, remaining, error).
func (r *RateLimiterAdapter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	nowNs := time.Now().UnixNano()
	windowNs := window.Nanoseconds()
	ttlMs := window.Milliseconds()

	res, err := slidingWindowScript.Run(ctx, r.rdb,
		[]string{key},
		nowNs, windowNs, limit, ttlMs,
	).Int64Slice()
	if err != nil {
		return false, 0, fmt.Errorf("rate_limiter: lua: %w", err)
	}

	allowed := res[0] == 1
	count := int(res[1])
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}
	return allowed, remaining, nil
}
