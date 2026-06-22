// Package integration contains integration tests that require real infrastructure
// (PostgreSQL and Redis). They are skipped unless the TEST_DATABASE_URL environment
// variable is set.
//
// Run with: TEST_DATABASE_URL=postgres://... TEST_REDIS_URL=redis://... go test ./tests/integration/...
package integration_test
