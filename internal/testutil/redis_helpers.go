package testutil

import "testing"

const defaultTestRedisURL = "redis://localhost:6379/0"

// TestRedis returns Redis URL used for integration tests.
func TestRedis(t *testing.T) string {
	t.Helper()
	return defaultTestRedisURL
}
