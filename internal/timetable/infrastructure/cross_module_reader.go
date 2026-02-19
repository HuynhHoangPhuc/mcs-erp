package infrastructure

import (
	"context"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/scheduler"
)

// --- Local interface definitions (avoid import cycles) ---

// teacherLister can list teachers with pagination.
type teacherLister interface {
	// List returns teachers; the filter parameter is passed as any to avoid
	// importing hr/domain. Callers pass hr/domain.TeacherFilter{}.
	ListAll(ctx context.Context) ([]teacherRow, error)
}

// teacherRow is a minimal teacher record returned by the adapter.
type teacherRow struct {
	ID             uuid.UUID
	Qualifications []string
}

// teacherAvailGetter retrieves teacher weekly availability.
type teacherAvailGetter interface {
	GetByTeacherID(ctx context.Context, id uuid.UUID) ([]availRow, error)
}

// availRow is a minimal availability record.
type availRow struct {
	Day         int
	Period      int
	IsAvailable bool
}

// subjectLister can list subjects with pagination.
type subjectLister interface {
	ListAll(ctx context.Context) ([]subjectRow, error)
}

// subjectRow is a minimal subject record.
type subjectRow struct {
	ID           uuid.UUID
	HoursPerWeek int
}

// roomLister can list rooms with pagination.
type roomLister interface {
	ListAll(ctx context.Context) ([]roomRow, error)
}

// roomRow is a minimal room record.
type roomRow struct {
	ID        uuid.UUID
	Capacity  int
	Equipment []string
}

// roomAvailGetter retrieves room weekly availability.
type roomAvailGetter interface {
	GetByRoomID(ctx context.Context, id uuid.UUID) ([]availRow, error)
}

// --- CrossModuleReader ---

// CrossModuleReader assembles a scheduler.Problem by reading data from other
// modules via narrow interfaces, avoiding direct package imports.
type CrossModuleReader struct {
	teachers     teacherLister
	teacherAvail teacherAvailGetter
	subjects     subjectLister
	rooms        roomLister
	roomAvail    roomAvailGetter
}

// NewCrossModuleReader creates a CrossModuleReader wired to the given adapters.
func NewCrossModuleReader(
	teachers teacherLister,
	teacherAvail teacherAvailGetter,
	subjects subjectLister,
	rooms roomLister,
	roomAvail roomAvailGetter,
) *CrossModuleReader {
	return &CrossModuleReader{
		teachers:     teachers,
		teacherAvail: teacherAvail,
		subjects:     subjects,
		rooms:        rooms,
		roomAvail:    roomAvail,
	}
}

// BuildProblem assembles a scheduler.Problem for the given semester.
// semesterSubjects constrains which subjects and teacher assignments to include.
func (r *CrossModuleReader) BuildProblem(
	ctx context.Context,
	semesterSubjects []*domain.SemesterSubject,
) (scheduler.Problem, error) {
	// --- Subjects ---
	allSubjects, err := r.subjects.ListAll(ctx)
	if err != nil {
		return scheduler.Problem{}, err
	}

	// Build a set of subject IDs in this semester for quick lookup.
	semSubjectSet := make(map[uuid.UUID]uuid.UUID, len(semesterSubjects)) // subjectID -> teacherID (zero if unset)
	for _, ss := range semesterSubjects {
		tid := uuid.Nil
		if ss.TeacherID != nil {
			tid = *ss.TeacherID
		}
		semSubjectSet[ss.SubjectID] = tid
	}

	var subjInfos []scheduler.SubjectInfo
	for _, s := range allSubjects {
		if _, ok := semSubjectSet[s.ID]; ok {
			subjInfos = append(subjInfos, scheduler.SubjectInfo{
				ID:           s.ID,
				HoursPerWeek: s.HoursPerWeek,
			})
		}
	}

	// Build pre-assignment map (subjectID -> teacherID) for subjects with teachers set.
	teacherAssign := make(map[uuid.UUID]uuid.UUID)
	for subjectID, teacherID := range semSubjectSet {
		if teacherID != uuid.Nil {
			teacherAssign[subjectID] = teacherID
		}
	}

	// --- Teachers ---
	allTeachers, err := r.teachers.ListAll(ctx)
	if err != nil {
		return scheduler.Problem{}, err
	}

	var teacherInfos []scheduler.TeacherInfo
	for _, t := range allTeachers {
		avail, err := r.teacherAvail.GetByTeacherID(ctx, t.ID)
		if err != nil {
			return scheduler.Problem{}, err
		}
		grid := buildAvailGrid(avail)
		teacherInfos = append(teacherInfos, scheduler.TeacherInfo{
			ID:             t.ID,
			Available:      grid,
			Qualifications: t.Qualifications,
		})
	}

	// --- Rooms ---
	allRooms, err := r.rooms.ListAll(ctx)
	if err != nil {
		return scheduler.Problem{}, err
	}

	var roomInfos []scheduler.RoomInfo
	for _, room := range allRooms {
		avail, err := r.roomAvail.GetByRoomID(ctx, room.ID)
		if err != nil {
			return scheduler.Problem{}, err
		}
		grid := buildAvailGrid(avail)
		roomInfos = append(roomInfos, scheduler.RoomInfo{
			ID:        room.ID,
			Capacity:  room.Capacity,
			Equipment: room.Equipment,
			Available: grid,
		})
	}

	return scheduler.Problem{
		Subjects:      subjInfos,
		Teachers:      teacherInfos,
		Rooms:         roomInfos,
		TeacherAssign: teacherAssign,
		Slots:         domain.AllSlots(),
	}, nil
}

// buildAvailGrid converts flat availability rows to a TimeSlot → bool map.
func buildAvailGrid(rows []availRow) map[domain.TimeSlot]bool {
	grid := make(map[domain.TimeSlot]bool, len(rows))
	for _, r := range rows {
		grid[domain.TimeSlot{Day: r.Day, Period: r.Period}] = r.IsAvailable
	}
	return grid
}
