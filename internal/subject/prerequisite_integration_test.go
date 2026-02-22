//go:build integration

package subject_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
)

func TestPrerequisiteLifecycleAndChain(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)

	subjectA := createSubject(t, srv.URL, token, schema, "Subject A", "SUBA")
	subjectB := createSubject(t, srv.URL, token, schema, "Subject B", "SUBB")
	subjectC := createSubject(t, srv.URL, token, schema, "Subject C", "SUBC")
	subjectD := createSubject(t, srv.URL, token, schema, "Subject D", "SUBD")

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+subjectA+"/prerequisites", token, schema, jsonBody(t, map[string]any{
		"prerequisite_id":  subjectB,
		"expected_version": 0,
	})), http.StatusCreated)

	list := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects/"+subjectA+"/prerequisites", token, schema, nil), http.StatusOK)
	items, ok := list["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected 1 prerequisite edge, got %v", list["items"])
	}
	if fmt.Sprintf("%v", list["version"]) == "" {
		t.Fatalf("expected version in prerequisite list response")
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+subjectB+"/prerequisites", token, schema, jsonBody(t, map[string]any{
		"prerequisite_id":  subjectC,
		"expected_version": 0,
	})), http.StatusCreated)
	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+subjectC+"/prerequisites", token, schema, jsonBody(t, map[string]any{
		"prerequisite_id":  subjectD,
		"expected_version": 0,
	})), http.StatusCreated)

	chain := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects/"+subjectA+"/prerequisite-chain", token, schema, nil), http.StatusOK)
	chainItems, ok := chain["chain"].([]any)
	if !ok || len(chainItems) < 3 {
		t.Fatalf("expected transitive prerequisite chain length >= 3, got %v", chain["chain"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodDelete, srv.URL+"/api/v1/subjects/"+subjectA+"/prerequisites/"+subjectB, token, schema, nil), http.StatusOK)
	afterDelete := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects/"+subjectA+"/prerequisites", token, schema, nil), http.StatusOK)
	afterItems, ok := afterDelete["items"].([]any)
	if !ok {
		t.Fatalf("expected items array after delete, got %T", afterDelete["items"])
	}
	if len(afterItems) != 0 {
		t.Fatalf("expected 0 prerequisites after delete, got %d", len(afterItems))
	}
}

func TestPrerequisiteCycleAndOptimisticLocking(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)

	a := createSubject(t, srv.URL, token, schema, "A", "A001")
	b := createSubject(t, srv.URL, token, schema, "B", "B001")
	c := createSubject(t, srv.URL, token, schema, "C", "C001")

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+a+"/prerequisites", token, schema, jsonBody(t, map[string]any{"prerequisite_id": b, "expected_version": 0})), http.StatusCreated)
	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+b+"/prerequisites", token, schema, jsonBody(t, map[string]any{"prerequisite_id": c, "expected_version": 0})), http.StatusCreated)

	cycleResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+c+"/prerequisites", token, schema, jsonBody(t, map[string]any{"prerequisite_id": a, "expected_version": 0})), http.StatusConflict)
	if fmt.Sprintf("%v", cycleResp["error"]) == "" {
		t.Fatalf("expected cycle detection error")
	}

	x := createSubject(t, srv.URL, token, schema, "X", "X001")
	y := createSubject(t, srv.URL, token, schema, "Y", "Y001")
	z := createSubject(t, srv.URL, token, schema, "Z", "Z001")

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+x+"/prerequisites", token, schema, jsonBody(t, map[string]any{"prerequisite_id": y, "expected_version": 0})), http.StatusCreated)
	conflict := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+x+"/prerequisites", token, schema, jsonBody(t, map[string]any{"prerequisite_id": z, "expected_version": 0})), http.StatusConflict)
	if fmt.Sprintf("%v", conflict["error"]) == "" {
		t.Fatalf("expected optimistic locking conflict error")
	}

	invalidReq := mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+x+"/prerequisites", token, schema, jsonBody(t, map[string]any{"expected_version": 1}))
	_, invalidStatus := testutil.DoJSON[map[string]any](t, http.DefaultClient, invalidReq)
	if invalidStatus != http.StatusBadRequest {
		t.Fatalf("expected 400 when prerequisite_id missing, got %d", invalidStatus)
	}
}

func TestPrerequisitePermissions(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	readOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermSubjectRead})
	writeOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermSubjectWrite})
	noPerm := testutil.GenerateTestToken(t, uuid.New(), schema, nil)

	a := createSubjectWithToken(t, srv.URL, writeOnly, schema, "Perm A", "PRA")
	b := createSubjectWithToken(t, srv.URL, writeOnly, schema, "Perm B", "PRB")

	_, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+a+"/prerequisites", writeOnly, schema, jsonBody(t, map[string]any{"prerequisite_id": b, "expected_version": 0})))
	if status != http.StatusCreated {
		t.Fatalf("expected write permission to add prerequisite (201), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects/"+a+"/prerequisites", readOnly, schema, nil))
	if status != http.StatusOK {
		t.Fatalf("expected read permission to list prerequisites (200), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/subjects/"+a+"/prerequisites", noPerm, schema, jsonBody(t, map[string]any{"prerequisite_id": b, "expected_version": 0})))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token forbidden on prerequisite write, got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/subjects/"+a+"/prerequisites", noPerm, schema, nil))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token forbidden on prerequisite read, got %d", status)
	}
}

func createSubject(t *testing.T, baseURL, token, schema, name, code string) string {
	t.Helper()
	return createSubjectWithToken(t, baseURL, token, schema, name, code)
}

func createSubjectWithToken(t *testing.T, baseURL, token, schema, name, code string) string {
	t.Helper()
	resp := getJSON(t, mustAuthReq(t, http.MethodPost, baseURL+"/api/v1/subjects", token, schema, jsonBody(t, map[string]any{
		"name":           name,
		"code":           code,
		"description":    "subject for prerequisite tests",
		"credits":        3,
		"hours_per_week": 3,
	})), http.StatusCreated)
	id := fmt.Sprintf("%v", resp["id"])
	if id == "" {
		t.Fatalf("subject id empty")
	}
	return id
}
