//go:build integration

package core_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	coreinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/core/infrastructure"
	platformauth "github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
)

func TestLogin_ValidCredentials_ReturnsTokenPair(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	payload := map[string]string{"email": admin.Email, "password": admin.Password}
	resp := postJSON(t, srv.URL+"/api/v1/auth/login", payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	decodeJSON(t, resp, &body)
	if body["access_token"] == "" || body["refresh_token"] == "" {
		t.Fatalf("expected non-empty token pair")
	}
}

func TestLogin_WrongPassword_Returns401(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	payload := map[string]string{"email": admin.Email, "password": "wrong-password"}
	resp := postJSON(t, srv.URL+"/api/v1/auth/login", payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_NonExistentEmail_Returns401(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	_ = testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	payload := map[string]string{"email": "none@example.com", "password": "whatever"}
	resp := postJSON(t, srv.URL+"/api/v1/auth/login", payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_DeactivatedUser_Returns401(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	userEmail, password := seedDeactivatedUser(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	payload := map[string]string{"email": userEmail, "password": password}
	resp := postJSON(t, srv.URL+"/api/v1/auth/login", payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthenticatedRequest_ValidToken_Succeeds(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	accessToken := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)
	req, err := testutil.AuthenticatedRequest(http.MethodGet, srv.URL+"/api/v1/users", accessToken, schema, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request users: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuthenticatedRequest_NoToken_Returns401(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	_ = testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/users", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("X-Tenant-ID", schema)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request users: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthenticatedRequest_ExpiredToken_Returns401(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	expiredToken := generateExpiredToken(t, admin.UserID, admin.Email, schema)
	req, err := testutil.AuthenticatedRequest(http.MethodGet, srv.URL+"/api/v1/users", expiredToken, schema, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request users: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthenticatedRequest_TamperedToken_Returns401(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	validToken := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)
	tampered := tamperToken(validToken)

	req, err := testutil.AuthenticatedRequest(http.MethodGet, srv.URL+"/api/v1/users", tampered, schema, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request users: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRefresh_ValidRefreshToken_ReturnsNewPair(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	loginResp := loginAndGetPair(t, srv.URL, admin.Email, admin.Password)
	refreshPayload := map[string]string{"refresh_token": loginResp.RefreshToken}
	resp := postJSONWithTenant(t, srv.URL+"/api/v1/auth/refresh", refreshPayload, schema)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var refreshed coreinfra.TokenPair
	decodeJSON(t, resp, &refreshed)
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatalf("expected refreshed token pair")
	}
	if refreshed.AccessToken == loginResp.AccessToken {
		t.Fatalf("expected new access token")
	}
}

func TestRefresh_InvalidToken_Returns401(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	_ = testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	payload := map[string]string{"refresh_token": "invalid-token"}
	resp := postJSONWithTenant(t, srv.URL+"/api/v1/auth/refresh", payload, schema)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRBAC_WithoutPermission_Returns403(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := testutil.GenerateTestToken(t, admin.UserID, schema, []string{})
	req, err := testutil.AuthenticatedRequest(http.MethodGet, srv.URL+"/api/v1/users", token, schema, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request users: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestRBAC_AdminCanCreateUser(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	accessToken := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)
	createReq := map[string]string{
		"email":    fmt.Sprintf("u_%s@example.com", uuid.NewString()),
		"password": "pass123!",
		"name":     "New User",
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("marshal create user payload: %v", err)
	}

	req, err := testutil.AuthenticatedRequest(http.MethodPost, srv.URL+"/api/v1/users", accessToken, schema, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create user request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func postJSON(t *testing.T, url string, payload any) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	return resp
}

func postJSONWithTenant(t *testing.T, url string, payload any, tenantID string) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request %s: %v", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, out any) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func loginAndGetPair(t *testing.T, baseURL, email, password string) coreinfra.TokenPair {
	t.Helper()
	resp := postJSON(t, baseURL+"/api/v1/auth/login", map[string]string{"email": email, "password": password})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	var pair coreinfra.TokenPair
	decodeJSON(t, resp, &pair)
	return pair
}

func loginAndGetToken(t *testing.T, baseURL, email, password string) string {
	t.Helper()
	return loginAndGetPair(t, baseURL, email, password).AccessToken
}

func seedDeactivatedUser(t *testing.T, pool *pgxpool.Pool, schema string) (string, string) {
	t.Helper()

	password := "deactivated123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	userID := uuid.New()
	roleID := uuid.New()
	email := fmt.Sprintf("deactivated_%s@example.com", uuid.NewString())

	if _, err := pool.Exec(context.Background(),
		`INSERT INTO public.tenants (name, schema_name, is_active, created_at, updated_at)
		 VALUES ($1, $2, true, now(), now())
		 ON CONFLICT (schema_name) DO NOTHING`,
		"Tenant "+schema, schema,
	); err != nil {
		t.Fatalf("insert tenant: %v", err)
	}

	if _, err := pool.Exec(context.Background(),
		`INSERT INTO public.users_lookup (email, tenant_schema)
		 VALUES ($1, $2)
		 ON CONFLICT (email) DO UPDATE SET tenant_schema = EXCLUDED.tenant_schema`,
		email, schema,
	); err != nil {
		t.Fatalf("upsert users_lookup: %v", err)
	}

	err = database.WithTenantTx(context.Background(), pool, schema, func(tx pgx.Tx) error {
		now := time.Now().UTC()
		if _, err := tx.Exec(context.Background(),
			`INSERT INTO roles (id, name, permissions, description, created_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			roleID,
			"inactive_role_"+uuid.NewString(),
			[]string{coredomain.PermUserRead},
			"inactive user role",
			now,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(context.Background(),
			`INSERT INTO users (id, email, password_hash, name, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, false, $5, $5)`,
			userID,
			email,
			string(hash),
			"Deactivated User",
			now,
		); err != nil {
			return err
		}

		_, err := tx.Exec(context.Background(), `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, roleID)
		return err
	})
	if err != nil {
		t.Fatalf("seed deactivated user: %v", err)
	}

	return email, password
}

func generateExpiredToken(t *testing.T, userID uuid.UUID, email, schema string) string {
	t.Helper()
	claims := &platformauth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			ID:        uuid.NewString(),
		},
		UserID:      userID,
		TenantID:    schema,
		Email:       email,
		Permissions: []string{coredomain.PermUserRead},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret-do-not-use-in-production"))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	return token
}

func tamperToken(token string) string {
	parts := bytes.Split([]byte(token), []byte("."))
	if len(parts) != 3 {
		return token + "tampered"
	}
	decoded, err := base64.RawURLEncoding.DecodeString(string(parts[1]))
	if err != nil {
		return token + "tampered"
	}
	if len(decoded) > 0 {
		decoded[len(decoded)-1] ^= 0x01
	}
	parts[1] = []byte(base64.RawURLEncoding.EncodeToString(decoded))
	return string(bytes.Join(parts, []byte(".")))
}
