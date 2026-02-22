package testutil

import (
	"time"

	"github.com/google/uuid"

	timetabledomain "github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
)

const defaultAdminPassword = "admin123!"

// SeedResult holds key data for a seeded admin user.
type SeedResult struct {
	UserID   uuid.UUID
	RoleID   uuid.UUID
	Email    string
	Password string
	Schema   string
}

// TeacherFixture is the result of seeding a teacher.
type TeacherFixture struct {
	ID           uuid.UUID
	DepartmentID *uuid.UUID
	Name         string
	Email        string
}

// SubjectFixture is the result of seeding a subject.
type SubjectFixture struct {
	ID   uuid.UUID
	Name string
	Code string
}

// RoomFixture is the result of seeding a room.
type RoomFixture struct {
	ID       uuid.UUID
	Name     string
	Code     string
	Capacity int
}

// SemesterFixture is the result of seeding a semester.
type SemesterFixture struct {
	ID        uuid.UUID
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    timetabledomain.SemesterStatus
}

type TeacherOption func(*teacherSeedOpts)

type teacherSeedOpts struct {
	Name         string
	Email        string
	DepartmentID *uuid.UUID
}

func WithTeacherName(name string) TeacherOption {
	return func(o *teacherSeedOpts) { o.Name = name }
}

func WithTeacherEmail(email string) TeacherOption {
	return func(o *teacherSeedOpts) { o.Email = email }
}

func WithTeacherDepartmentID(departmentID uuid.UUID) TeacherOption {
	return func(o *teacherSeedOpts) { o.DepartmentID = &departmentID }
}

type SubjectOption func(*subjectSeedOpts)

type subjectSeedOpts struct {
	Name         string
	Code         string
	Description  string
	CategoryID   *uuid.UUID
	Credits      int
	HoursPerWeek int
}

func WithSubjectName(name string) SubjectOption {
	return func(o *subjectSeedOpts) { o.Name = name }
}

func WithSubjectCode(code string) SubjectOption {
	return func(o *subjectSeedOpts) { o.Code = code }
}

func WithSubjectCategoryID(categoryID uuid.UUID) SubjectOption {
	return func(o *subjectSeedOpts) { o.CategoryID = &categoryID }
}

type RoomOption func(*roomSeedOpts)

type roomSeedOpts struct {
	Name     string
	Code     string
	Capacity int
}

func WithRoomName(name string) RoomOption {
	return func(o *roomSeedOpts) { o.Name = name }
}

func WithRoomCode(code string) RoomOption {
	return func(o *roomSeedOpts) { o.Code = code }
}

func WithRoomCapacity(capacity int) RoomOption {
	return func(o *roomSeedOpts) { o.Capacity = capacity }
}

type SemesterOption func(*semesterSeedOpts)

type semesterSeedOpts struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    timetabledomain.SemesterStatus
}

func WithSemesterName(name string) SemesterOption {
	return func(o *semesterSeedOpts) { o.Name = name }
}

func stringsNoHyphen(in string) string {
	out := make([]rune, 0, len(in))
	for _, r := range in {
		if r != '-' {
			out = append(out, r)
		}
	}
	return string(out)
}
