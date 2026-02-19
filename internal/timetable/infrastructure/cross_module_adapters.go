package infrastructure

import (
	"context"

	"github.com/google/uuid"

	hrDomain      "github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
	roomDomain    "github.com/HuynhHoangPhuc/mcs-erp/internal/room/domain"
	subjectDomain "github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
)

// This file provides thin adapters that convert the real HR/Subject/Room
// repository interfaces into the narrow local interfaces expected by
// CrossModuleReader.  All concrete wiring is done here so that module.go
// (and main.go) only need to call NewCrossModuleReader with these adapters.

// --- Teacher adapter ---

// TeacherRepoAdapter adapts hrDomain.TeacherRepository to the teacherLister interface.
type TeacherRepoAdapter struct {
	repo hrDomain.TeacherRepository
}

// NewTeacherRepoAdapter wraps an hr TeacherRepository.
func NewTeacherRepoAdapter(repo hrDomain.TeacherRepository) *TeacherRepoAdapter {
	return &TeacherRepoAdapter{repo: repo}
}

// ListAll paginates through all teachers and returns minimal rows.
func (a *TeacherRepoAdapter) ListAll(ctx context.Context) ([]teacherRow, error) {
	const pageSize = 200
	var result []teacherRow
	offset := 0
	for {
		teachers, _, err := a.repo.List(ctx, hrDomain.TeacherFilter{}, offset, pageSize)
		if err != nil {
			return nil, err
		}
		for _, t := range teachers {
			result = append(result, teacherRow{
				ID:             t.ID,
				Qualifications: t.Qualifications,
			})
		}
		if len(teachers) < pageSize {
			break
		}
		offset += pageSize
	}
	return result, nil
}

// --- Teacher availability adapter ---

// TeacherAvailAdapter adapts hrDomain.AvailabilityRepository to teacherAvailGetter.
type TeacherAvailAdapter struct {
	repo hrDomain.AvailabilityRepository
}

// NewTeacherAvailAdapter wraps an hr AvailabilityRepository.
func NewTeacherAvailAdapter(repo hrDomain.AvailabilityRepository) *TeacherAvailAdapter {
	return &TeacherAvailAdapter{repo: repo}
}

// GetByTeacherID returns availability rows in the local availRow shape.
func (a *TeacherAvailAdapter) GetByTeacherID(ctx context.Context, id uuid.UUID) ([]availRow, error) {
	slots, err := a.repo.GetByTeacherID(ctx, id)
	if err != nil {
		return nil, err
	}
	rows := make([]availRow, len(slots))
	for i, s := range slots {
		rows[i] = availRow{
			Day:         s.Day,
			Period:      s.Period,
			IsAvailable: s.IsAvailable,
		}
	}
	return rows, nil
}

// --- Subject adapter ---

// SubjectRepoAdapter adapts subjectDomain.SubjectRepository to the subjectLister interface.
type SubjectRepoAdapter struct {
	repo subjectDomain.SubjectRepository
}

// NewSubjectRepoAdapter wraps a subject SubjectRepository.
func NewSubjectRepoAdapter(repo subjectDomain.SubjectRepository) *SubjectRepoAdapter {
	return &SubjectRepoAdapter{repo: repo}
}

// ListAll paginates through all subjects and returns minimal rows.
func (a *SubjectRepoAdapter) ListAll(ctx context.Context) ([]subjectRow, error) {
	const pageSize = 200
	var result []subjectRow
	offset := 0
	for {
		subjects, _, err := a.repo.List(ctx, offset, pageSize)
		if err != nil {
			return nil, err
		}
		for _, s := range subjects {
			result = append(result, subjectRow{
				ID:           s.ID,
				HoursPerWeek: s.HoursPerWeek,
			})
		}
		if len(subjects) < pageSize {
			break
		}
		offset += pageSize
	}
	return result, nil
}

// --- Room adapter ---

// RoomRepoAdapter adapts roomDomain.RoomRepository to the roomLister interface.
type RoomRepoAdapter struct {
	repo roomDomain.RoomRepository
}

// NewRoomRepoAdapter wraps a room RoomRepository.
func NewRoomRepoAdapter(repo roomDomain.RoomRepository) *RoomRepoAdapter {
	return &RoomRepoAdapter{repo: repo}
}

// ListAll fetches all active rooms and returns minimal rows.
func (a *RoomRepoAdapter) ListAll(ctx context.Context) ([]roomRow, error) {
	rooms, err := a.repo.List(ctx, roomDomain.ListFilter{})
	if err != nil {
		return nil, err
	}
	result := make([]roomRow, len(rooms))
	for i, r := range rooms {
		result[i] = roomRow{
			ID:        r.ID,
			Capacity:  r.Capacity,
			Equipment: r.Equipment,
		}
	}
	return result, nil
}

// --- Room availability adapter ---

// RoomAvailAdapter adapts roomDomain.RoomAvailabilityRepository to roomAvailGetter.
type RoomAvailAdapter struct {
	repo roomDomain.RoomAvailabilityRepository
}

// NewRoomAvailAdapter wraps a room RoomAvailabilityRepository.
func NewRoomAvailAdapter(repo roomDomain.RoomAvailabilityRepository) *RoomAvailAdapter {
	return &RoomAvailAdapter{repo: repo}
}

// GetByRoomID returns availability rows in the local availRow shape.
func (a *RoomAvailAdapter) GetByRoomID(ctx context.Context, id uuid.UUID) ([]availRow, error) {
	avail, err := a.repo.GetByRoomID(ctx, id)
	if err != nil {
		return nil, err
	}
	rows := make([]availRow, 0, len(avail))
	for slot, isAvail := range avail {
		rows = append(rows, availRow{
			Day:         slot.Day,
			Period:      slot.Period,
			IsAvailable: isAvail,
		})
	}
	return rows, nil
}

// --- Constructor helper ---

// NewCrossModuleReaderFromRepos is the convenience constructor used in module.go.
// It wraps all four concrete repos and wires them into a CrossModuleReader.
func NewCrossModuleReaderFromRepos(
	teacherRepo  hrDomain.TeacherRepository,
	availRepo    hrDomain.AvailabilityRepository,
	subjectRepo  subjectDomain.SubjectRepository,
	roomRepo     roomDomain.RoomRepository,
	roomAvail    roomDomain.RoomAvailabilityRepository,
) *CrossModuleReader {
	return NewCrossModuleReader(
		NewTeacherRepoAdapter(teacherRepo),
		NewTeacherAvailAdapter(availRepo),
		NewSubjectRepoAdapter(subjectRepo),
		NewRoomRepoAdapter(roomRepo),
		NewRoomAvailAdapter(roomAvail),
	)
}
