//go:build integration

package timetable_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
	timetabledomain "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	timetableinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/infrastructure"
	timetablescheduler "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/scheduler"
)

func TestScheduleGenerationAndApprovalFlow(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)

	teacherA := testutil.SeedTeacher(t, db.Pool, schema)
	teacherB := testutil.SeedTeacher(t, db.Pool, schema)
	subjectA := testutil.SeedSubject(t, db.Pool, schema, testutil.WithSubjectCode("SCH-A"), testutil.WithSubjectName("Sched A"))
	subjectB := testutil.SeedSubject(t, db.Pool, schema, testutil.WithSubjectCode("SCH-B"), testutil.WithSubjectName("Sched B"))
	_ = testutil.SeedRoom(t, db.Pool, schema, testutil.WithRoomCode("ROOMA"))
	roomB := testutil.SeedRoom(t, db.Pool, schema, testutil.WithRoomCode("ROOMB"))

	semesterID := createSemester(t, srv.URL, token, schema, "Sched Semester")
	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/subjects", token, schema, jsonBody(t, map[string]any{
		"subject_ids": []string{subjectA.ID.String(), subjectB.ID.String()},
	})), http.StatusOK)
	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/subjects/"+subjectA.ID.String()+"/teacher", token, schema, jsonBody(t, map[string]any{
		"teacher_id": teacherA.ID.String(),
	})), http.StatusOK)

	start := time.Now()
	generated := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/generate", token, schema, nil), http.StatusOK)
	elapsed := time.Since(start)
	if elapsed > 20*time.Second {
		t.Fatalf("expected schedule generation to finish within sanity bound, got %v", elapsed)
	}
	if int(generated["hard_violations"].(float64)) != 0 {
		t.Fatalf("expected 0 hard violations after generation, got %v", generated["hard_violations"])
	}

	assignmentsAny, ok := generated["assignments"].([]any)
	if !ok || len(assignmentsAny) == 0 {
		t.Fatalf("expected non-empty generated assignments, got %v", generated["assignments"])
	}

	repo := timetableinfra.NewPostgresScheduleRepo(db.Pool)
	sched, err := repo.FindLatestBySemester(tenant.WithTenant(context.Background(), schema), mustUUID(t, semesterID))
	if err != nil {
		t.Fatalf("load latest schedule from repo: %v", err)
	}
	if sched.HardViolations != 0 {
		t.Fatalf("expected repo schedule hard violations 0, got %d", sched.HardViolations)
	}
	assertNoDoubleBookingConflicts(t, sched)

	semesterRepo := timetableinfra.NewPostgresSemesterRepo(db.Pool)
	sem, err := semesterRepo.FindByID(tenant.WithTenant(context.Background(), schema), mustUUID(t, semesterID))
	if err != nil {
		t.Fatalf("load semester after generation: %v", err)
	}
	if sem.Status != timetabledomain.SemesterStatusReview {
		t.Fatalf("expected semester status review after generation, got %s", sem.Status)
	}

	approved := getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/approve", token, schema, nil), http.StatusOK)
	if fmt.Sprintf("%v", approved["status"]) != "approved" {
		t.Fatalf("expected approve response status approved, got %v", approved["status"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/approve", token, schema, nil), http.StatusBadRequest)

	latest := getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/schedule", token, schema, nil), http.StatusOK)
	if int(latest["hard_violations"].(float64)) != 0 {
		t.Fatalf("expected latest schedule hard_violations=0, got %v", latest["hard_violations"])
	}

	if len(sched.Assignments) == 0 {
		t.Fatalf("expected persisted schedule assignments to be non-empty")
	}
	firstAssignment := sched.Assignments[0]
	updateResp := getJSON(t, mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/timetable/assignments/"+firstAssignment.ID.String(), token, schema, jsonBody(t, map[string]any{
		"teacher_id": teacherB.ID.String(),
		"room_id":    roomB.ID.String(),
		"day":        2,
		"period":     3,
	})), http.StatusOK)
	if fmt.Sprintf("%v", updateResp["teacher_id"]) != teacherB.ID.String() {
		t.Fatalf("expected updated teacher_id %s, got %v", teacherB.ID, updateResp["teacher_id"])
	}
	if int(updateResp["day"].(float64)) != 2 || int(updateResp["period"].(float64)) != 3 {
		t.Fatalf("expected updated slot day=2 period=3, got day=%v period=%v", updateResp["day"], updateResp["period"])
	}

	_ = getJSON(t, mustAuthReq(t, http.MethodPut, srv.URL+"/api/v1/timetable/assignments/"+firstAssignment.ID.String(), token, schema, jsonBody(t, map[string]any{
		"teacher_id": "not-a-uuid",
	})), http.StatusBadRequest)
}

func TestScheduleGenerationValidationErrors(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	admin := testutil.SeedAdmin(t, db.Pool, schema)
	srv := testutil.TestServer(t, db.Pool)
	defer srv.Close()

	token := loginAndGetToken(t, srv.URL, admin.Email, admin.Password)
	semesterID := createSemester(t, srv.URL, token, schema, "Empty Semester")

	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/generate", token, schema, nil), http.StatusBadRequest)
	_ = getJSON(t, mustAuthReq(t, http.MethodGet, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/schedule", token, schema, nil), http.StatusNotFound)
	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/"+semesterID+"/approve", token, schema, nil), http.StatusBadRequest)
	_ = getJSON(t, mustAuthReq(t, http.MethodPost, srv.URL+"/api/v1/timetable/semesters/not-a-uuid/generate", token, schema, nil), http.StatusBadRequest)
}

func assertNoDoubleBookingConflicts(t *testing.T, sched *timetabledomain.Schedule) {
	t.Helper()
	violations := timetablescheduler.TeacherConflictConstraint{}.Evaluate(sched.Assignments) +
		timetablescheduler.RoomConflictConstraint{}.Evaluate(sched.Assignments)
	if violations != 0 {
		t.Fatalf("expected no teacher/room double-booking conflicts in assignments, got %d", violations)
	}
}
