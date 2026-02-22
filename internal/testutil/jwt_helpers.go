package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/infrastructure"
)

const testJWTSecret = "test-secret-do-not-use-in-production"

// TestJWTService creates a JWT service with a deterministic test secret.
func TestJWTService() *infrastructure.JWTService {
	return infrastructure.NewJWTService(testJWTSecret, 24*time.Hour)
}

// GenerateTestToken creates a signed access token for test requests.
func GenerateTestToken(t *testing.T, userID uuid.UUID, tenantID string, perms []string) string {
	t.Helper()

	pair, err := TestJWTService().GenerateTokenPair(userID, tenantID, "test@example.com", perms)
	if err != nil {
		t.Fatalf("generate test token: %v", err)
	}
	return pair.AccessToken
}
