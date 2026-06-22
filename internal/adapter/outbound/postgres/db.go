// Package postgres provides pgx/v5-based implementations of all repository
// port interfaces defined in internal/port/outbound/database.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a queried row does not exist.
var ErrNotFound = errors.New("record not found")

// DB wraps a pgxpool connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new DB pool using the DATABASE_URL environment variable.
// It pings the database to confirm connectivity before returning.
func New(ctx context.Context) (*DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return &DB{Pool: pool}, nil
}

// Close closes the underlying connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// ─── Pagination helpers ───────────────────────────────────────────────────────

// pageOffset returns the SQL OFFSET value for the given 1-based page number.
func pageOffset(page, perPage int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * perPage
}

// totalPages calculates the number of pages needed to show total items.
func totalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	return (total + perPage - 1) / perPage
}

// safeOrderBy returns a validated "column direction" fragment for ORDER BY.
// col is checked against allowlist; falls back to defaultCol if absent.
// order must be "asc" or "desc" — any other value becomes "asc".
func safeOrderBy(col, order string, allowlist map[string]struct{}, defaultCol string) string {
	if _, ok := allowlist[col]; !ok {
		col = defaultCol
	}
	if order != "desc" {
		order = "asc"
	}
	return col + " " + order
}

// isNotFound returns true when err is a pgx "no rows" sentinel.
func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
