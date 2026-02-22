//go:build integration

package timetable_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
	timetabledomain "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	timetableinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/infrastructure"
)

func TestSemesterCRUDAndSubjectManagement(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)

	teacher := testutil.SeedTeacher(t, db.Pool, schema)
	subjectA := testutil.SeedSubject(t, db.Pool, schema)
	subjectB := testutil.SeedSubject(t, db.Pool, schema)

	semesterID := createSemester(t, srv.URL, token, schema, "2026 Spring")

	list := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/timetable/semesters", token, schema, nil), http.StatusOK)
	if list["total"].(float64) < 1 {
		t.Fatalf("expected semesters total >= 1, got %v", list["total"])
	}

	byID := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/timetable/semesters/"+semesterID, token, schema, nil), http.StatusOK)
	if fmt.Sprintf("%v", byID["name"]) != "2026 Spring" {
		t.Fatalf("expected semester name 2026 Spring, got %v", byID["name"])
	}
	if fmt.Sprintf("%v", byID["status"]) != "draft" {
		t.Fatalf("expected semester status draft, got %v", byID["status"])
	}

	setSubjects := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/subjects", token, schema, jsonBody(t, map[string]any{
		"subject_ids": []string{subjectA.ID.String(), subjectB.ID.String()},
	})), http.StatusOK)
	if int(setSubjects["added"].(float64)) != 2 {
		t.Fatalf("expected added=2, got %v", setSubjects["added"])
	}

	assignResp := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/subjects/"+subjectA.ID.String()+"/teacher", token, schema, jsonBody(t, map[string]any{
		"teacher_id": teacher.ID.String(),
	})), http.StatusOK)
	if ok, _ := assignResp["ok"].(bool); !ok {
		t.Fatalf("expected assign teacher response ok=true, got %v", assignResp)
	}

	repo := timetableinfra.NewPostgresSemesterRepo(db.Pool)
	items, err := repo.GetSubjects(tenant.WithTenant(context.Background(), schema), mustUUID(t, semesterID))
	if err != nil {
		t.Fatalf("repo get semester subjects: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 semester subjects, got %d", len(items))
	}
	if !hasTeacherAssignment(items, subjectA.ID, teacher.ID) {
		t.Fatalf("expected subject %s assigned to teacher %s", subjectA.ID, teacher.ID)
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/subjects", token, schema, jsonBody(t, map[string]any{
		"subject_ids": []string{"not-a-uuid"},
	})), http.StatusBadRequest)

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters", token, schema, jsonBody(t, map[string]any{
		"name":       "Invalid",
		"start_date": time.Now().UTC().Format(time.RFC3339),
		"end_date":   time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
	})), http.StatusBadRequest)
}

func TestSemesterAndSchedulePermissions(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	readOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermTimetableRead})
	writeOnly := testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermTimetableWrite})
	noPerm := testutil.GenerateTestToken(t, uuid.New(), schema, nil)

	semesterBody := jsonBody(t, map[string]any{
		"name":       "Permission Semester",
		"start_date": time.Now().UTC().Format(time.RFC3339),
		"end_date":   time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	_, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters", writeOnly, schema, semesterBody))
	if status != http.StatusCreated {
		t.Fatalf("expected write token to create semester (201), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/timetable/semesters", readOnly, schema, nil))
	if status != http.StatusOK {
		t.Fatalf("expected read token to list semesters (200), got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters", readOnly, schema, semesterBody))
	if status != http.StatusForbidden {
		t.Fatalf("expected read-only token forbidden on create semester, got %d", status)
	}

	_, status = testutil.DoJSON[map[string]any](t, http.DefaultClient, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/timetable/semesters", noPerm, schema, nil))
	if status != http.StatusForbidden {
		t.Fatalf("expected no-permission token forbidden on list semesters, got %d", status)
	}
}

func createSemester(t *testing.T, baseURL, token, schema, name string) string {
	t.Helper()
	start := time.Now().UTC().AddDate(0, 0, 1)
	end := start.AddDate(0, 4, 0)
	resp := getJSON(t, mustAuthReq(t, http.MethodPost, baseURL+"/api/v1/timetable/semesters", token, schema, jsonBody(t, map[string]any{
		"name":       name,
		"start_date": start.Format(time.RFC3339),
		"end_date":   end.Format(time.RFC3339),
	})), http.StatusCreated)
	id := fmt.Sprintf("%v", resp["id"])
	if id == "" {
		t.Fatalf("semester id empty")
	}
	return id
}

func loginAndGetToken(t *testing.T, baseURL, email, password string) string {
	t.Helper()
	resp := postJSON(t, baseURL+"/api/v1/auth/login", map[string]string{"email": email, "password": password})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	var pair map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&pair); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	token := fmt.Sprintf("%v", pair["access_token"])
	if token == "" {
		t.Fatalf("empty access token")
	}
	return token
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

func mustAuthReq(t *testing.T, method, url, token, schema string, body []byte) *http.Request {
	t.Helper()
	reader := bytes.NewReader(body)
	req, err := testutil.AuthenticatedRequest(method, url, token, schema, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	return req
}

func jsonBody(t *testing.T, payload any) []byte {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return b
}

func getJSON(t *testing.T, req *http.Request, expected int) map[string]any {
	t.Helper()
	payload, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, req)
	if status != expected {
		t.Fatalf("expected status %d, got %d (payload=%v)", expected, status, payload)
	}
	return payload
}

func mustUUID(t *testing.T, raw string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", raw, err)
	}
	return id
}

func hasTeacherAssignment(items []*timetabledomain.SemesterSubject, subjectID, teacherID uuid.UUID) bool {
	for _, item := range items {
		if item.SubjectID != subjectID || item.TeacherID == nil {
			continue
		}
		if *item.TeacherID == teacherID {
			return true
		}
	}
	return false
}
