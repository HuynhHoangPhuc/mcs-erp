//go:build integration

package timetable_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	hrinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/hr/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	roominfra "github.com/HuynhHoangPhuc/mcs-erp/internal/room/infrastructure"
	subjectinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/subject/infrastructure"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/testutil"
	timetabledomain "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	timetableinfra "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/infrastructure"
	timetablescheduler "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/scheduler"
)

func TestCrossModuleReader_BuildProblemWithData(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	ctx := tenant.WithTenant(context.Background(), schema)

	teacherA := testutil.SeedTeacher(t, db.Pool, schema)
	teacherB := testutil.SeedTeacher(t, db.Pool, schema)
	subjectA := testutil.SeedSubject(t, db.Pool, schema)
	subjectB := testutil.SeedSubject(t, db.Pool, schema)
	roomA := testutil.SeedRoom(t, db.Pool, schema)
	roomB := testutil.SeedRoom(t, db.Pool, schema)

	reader := timetableinfra.NewCrossModuleReaderFromRepos(
		hrinfra.NewPostgresTeacherRepo(db.Pool),
		hrinfra.NewPostgresAvailabilityRepo(db.Pool),
		subjectinfra.NewPostgresSubjectRepo(db.Pool),
		roominfra.NewPostgresRoomRepo(db.Pool),
		roominfra.NewPostgresAvailabilityRepo(db.Pool),
	)

	assignedTeacher := teacherA.ID
	semesterSubjects := []*timetabledomain.SemesterSubject{
		{SemesterID: uuid.New(), SubjectID: subjectA.ID, TeacherID: &assignedTeacher},
		{SemesterID: uuid.New(), SubjectID: subjectB.ID},
	}

	problem, err := reader.BuildProblem(ctx, semesterSubjects)
	if err != nil {
		t.Fatalf("build problem: %v", err)
	}

	if len(problem.Subjects) != 2 {
		t.Fatalf("expected 2 subjects in problem, got %d", len(problem.Subjects))
	}
	if len(problem.Teachers) < 2 {
		t.Fatalf("expected at least 2 teachers in problem, got %d", len(problem.Teachers))
	}
	if len(problem.Rooms) < 2 {
		t.Fatalf("expected at least 2 rooms in problem, got %d", len(problem.Rooms))
	}
	if len(problem.Slots) != len(timetabledomain.AllSlots()) {
		t.Fatalf("expected %d slots, got %d", len(timetabledomain.AllSlots()), len(problem.Slots))
	}

	if !containsSubject(problem.Subjects, subjectA.ID) || !containsSubject(problem.Subjects, subjectB.ID) {
		t.Fatalf("expected seeded subjects to be included in problem")
	}
	if !containsTeacher(problem.Teachers, teacherA.ID) || !containsTeacher(problem.Teachers, teacherB.ID) {
		t.Fatalf("expected seeded teachers to be included in problem")
	}
	if !containsRoom(problem.Rooms, roomA.ID) || !containsRoom(problem.Rooms, roomB.ID) {
		t.Fatalf("expected seeded rooms to be included in problem")
	}

	gotTeacher, ok := problem.TeacherAssign[subjectA.ID]
	if !ok || gotTeacher != teacherA.ID {
		t.Fatalf("expected subject %s pre-assigned to teacher %s, got %s", subjectA.ID, teacherA.ID, gotTeacher)
	}
	if _, ok := problem.TeacherAssign[subjectB.ID]; ok {
		t.Fatalf("did not expect teacher assignment for subject %s", subjectB.ID)
	}
}

func TestCrossModuleReader_BuildProblemEmptyData(t *testing.T) {
	db := testutil.NewTestDB(t)
	schema := db.CreateTenantSchema(t)
	ctx := tenant.WithTenant(context.Background(), schema)

	reader := timetableinfra.NewCrossModuleReaderFromRepos(
		hrinfra.NewPostgresTeacherRepo(db.Pool),
		hrinfra.NewPostgresAvailabilityRepo(db.Pool),
		subjectinfra.NewPostgresSubjectRepo(db.Pool),
		roominfra.NewPostgresRoomRepo(db.Pool),
		roominfra.NewPostgresAvailabilityRepo(db.Pool),
	)

	problem, err := reader.BuildProblem(ctx, nil)
	if err != nil {
		t.Fatalf("build problem with empty data: %v", err)
	}

	if len(problem.Subjects) != 0 || len(problem.Teachers) != 0 || len(problem.Rooms) != 0 {
		t.Fatalf("expected empty problem, got subjects=%d teachers=%d rooms=%d", len(problem.Subjects), len(problem.Teachers), len(problem.Rooms))
	}
	if len(problem.TeacherAssign) != 0 {
		t.Fatalf("expected empty teacher assignment map")
	}
	if len(problem.Slots) != len(timetabledomain.AllSlots()) {
		t.Fatalf("expected slots initialized to all valid slots")
	}
}

func containsSubject(subjects []timetablescheduler.SubjectInfo, id uuid.UUID) bool {
	for _, s := range subjects {
		if s.ID == id {
			return true
		}
	}
	return false
}

func containsTeacher(teachers []timetablescheduler.TeacherInfo, id uuid.UUID) bool {
	for _, t := range teachers {
		if t.ID == id {
			return true
		}
	}
	return false
}

func containsRoom(rooms []timetablescheduler.RoomInfo, id uuid.UUID) bool {
	for _, r := range rooms {
		if r.ID == id {
			return true
		}
	}
	return false
}
