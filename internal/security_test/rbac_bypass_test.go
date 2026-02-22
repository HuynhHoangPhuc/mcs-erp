//go:build integration

package security_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
)

func TestRBACBypass_NoPermissionToken_DeniedAcrossProtectedEndpoints(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	noPerm := testutil.GenerateTestToken(t, uuid.New(), schema, nil)

	cases := []struct {
		name   string
		method string
		path   string
		body   map[string]any
	}{
		{name: "users create", method: http.MethodPost, path: "/api/v1/users", body: map[string]any{"email": "x@example.com", "password": "secret123", "name": "X"}},
		{name: "teachers create", method: http.MethodPost, path: "/api/v1/teachers", body: map[string]any{"name": "T", "email": "t@example.com"}},
		{name: "subjects create", method: http.MethodPost, path: "/api/v1/subjects", body: map[string]any{"name": "S", "code": "S1"}},
		{name: "rooms create", method: http.MethodPost, path: "/api/v1/rooms", body: map[string]any{"name": "R", "code": "R1", "capacity": 10}},
		{name: "semesters create", method: http.MethodPost, path: "/api/v1/timetable/semesters", body: map[string]any{"name": "Sem"}},
		{name: "schedule generate", method: http.MethodPost, path: "/api/v1/timetable/semesters/not-a-uuid/generate"},
		{name: "agent conversations list", method: http.MethodGet, path: "/api/v1/agent/conversations"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			if tc.body != nil {
				body = jsonBody(t, tc.body)
			}
			req := mustAuthReq(t, tc.method, srv.URL+tc.path, noPerm, schema, body)
			payload, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, req)
			if status != http.StatusForbidden {
				t.Fatalf("expected 403 for %s %s, got %d (payload=%v)", tc.method, tc.path, status, payload)
			}
		})
	}
}

func TestRBACBypass_ReadOnlyTokens_CannotWrite(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	cases := []struct {
		name   string
		token  string
		method string
		path   string
		body   map[string]any
	}{
		{
			name:   "teacher read cannot create teacher",
			token:  testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermTeacherRead}),
			method: http.MethodPost,
			path:   "/api/v1/teachers",
			body:   map[string]any{"name": "ReadOnly Teacher", "email": fmt.Sprintf("ro_teacher_%s@example.com", uuid.NewString())},
		},
		{
			name:   "subject read cannot create subject",
			token:  testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermSubjectRead}),
			method: http.MethodPost,
			path:   "/api/v1/subjects",
			body:   map[string]any{"name": "ReadOnly Subject", "code": "ROS-1"},
		},
		{
			name:   "room read cannot create room",
			token:  testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermRoomRead}),
			method: http.MethodPost,
			path:   "/api/v1/rooms",
			body:   map[string]any{"name": "ReadOnly Room", "code": "ROR-1", "capacity": 30},
		},
		{
			name:   "timetable read cannot create semester",
			token:  testutil.GenerateTestToken(t, uuid.New(), schema, []string{coredomain.PermTimetableRead}),
			method: http.MethodPost,
			path:   "/api/v1/timetable/semesters",
			body:   map[string]any{"name": "ReadOnly Semester"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := mustAuthReq(t, tc.method, srv.URL+tc.path, tc.token, schema, jsonBody(t, tc.body))
			payload, status := testutil.DoJSON[map[string]any](t, http.DefaultClient, req)
			if status != http.StatusForbidden {
				t.Fatalf("expected 403 for %s, got %d (payload=%v)", tc.name, status, payload)
			}
		})
	}
}
