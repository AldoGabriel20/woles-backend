// Package redis implements the cache port interfaces using go-redis/v9.
package redis

import (
	"context"
	"errors"
	"fmt"
	"os"

	goredis "github.com/redis/go-redis/v9"
)

// Client wraps a go-redis client.
type Client struct {
	rdb *goredis.Client
}

// New creates a new Redis Client using the REDIS_URL environment variable.
// It pings the server to confirm connectivity.
func New(ctx context.Context) (*Client, error) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return nil, errors.New("REDIS_URL is not set")
	}
	opt, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis: parse url: %w", err)
	}
	rdb := goredis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: ping: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// Close releases the underlying connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}
