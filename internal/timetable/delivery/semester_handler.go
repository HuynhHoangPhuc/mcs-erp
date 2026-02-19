package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/timetable/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// SemesterHandler handles semester CRUD and subject-assignment endpoints.
type SemesterHandler struct {
	semesterRepo domain.SemesterRepository
}

// NewSemesterHandler creates a new semester handler.
func NewSemesterHandler(semesterRepo domain.SemesterRepository) *SemesterHandler {
	return &SemesterHandler{semesterRepo: semesterRepo}
}

// --- Request/response types ---

type createSemesterRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"` // RFC3339 date
	EndDate   string `json:"end_date"`
}

type setSubjectsRequest struct {
	SubjectIDs []string `json:"subject_ids"`
}

type assignTeacherRequest struct {
	TeacherID string `json:"teacher_id"`
}

// --- Handlers ---

// CreateSemester handles POST /api/v1/timetable/semesters
func (h *SemesterHandler) CreateSemester(w http.ResponseWriter, r *http.Request) {
	var req createSemesterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errResp("name is required"))
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("start_date must be RFC3339"))
		return
	}
	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("end_date must be RFC3339"))
		return
	}
	if !endDate.After(startDate) {
		writeJSON(w, http.StatusBadRequest, errResp("end_date must be after start_date"))
		return
	}

	now := time.Now()
	s := &domain.Semester{
		ID:        uuid.New(),
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    domain.SemesterStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.semesterRepo.Save(r.Context(), s); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to create semester"))
		return
	}
	writeJSON(w, http.StatusCreated, semesterResponse(s))
}

// ListSemesters handles GET /api/v1/timetable/semesters
func (h *SemesterHandler) ListSemesters(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	semesters, total, err := h.semesterRepo.List(r.Context(), offset, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to list semesters"))
		return
	}

	items := make([]map[string]any, len(semesters))
	for i, s := range semesters {
		items[i] = semesterResponse(s)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// GetSemester handles GET /api/v1/timetable/semesters/{id}
func (h *SemesterHandler) GetSemester(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid semester id"))
		return
	}

	s, err := h.semesterRepo.FindByID(r.Context(), id)
	if err != nil {
		if err == erptypes.ErrNotFound {
			writeJSON(w, http.StatusNotFound, errResp("semester not found"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errResp("failed to get semester"))
		return
	}
	writeJSON(w, http.StatusOK, semesterResponse(s))
}

// SetSubjects handles POST /api/v1/timetable/semesters/{id}/subjects
func (h *SemesterHandler) SetSubjects(w http.ResponseWriter, r *http.Request) {
	semID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid semester id"))
		return
	}

	// Verify semester exists.
	if _, err := h.semesterRepo.FindByID(r.Context(), semID); err != nil {
		if err == erptypes.ErrNotFound {
			writeJSON(w, http.StatusNotFound, errResp("semester not found"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errResp("failed to load semester"))
		return
	}

	var req setSubjectsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
		return
	}
	if len(req.SubjectIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, errResp("subject_ids must not be empty"))
		return
	}

	subjectIDs := make([]uuid.UUID, 0, len(req.SubjectIDs))
	for _, raw := range req.SubjectIDs {
		sid, err := uuid.Parse(raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResp("invalid subject_id: "+raw))
			return
		}
		subjectIDs = append(subjectIDs, sid)
	}

	if err := h.semesterRepo.AddSubjects(r.Context(), semID, subjectIDs); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to set subjects"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"added": len(subjectIDs)})
}

// AssignTeacher handles POST /api/v1/timetable/semesters/{id}/subjects/{subjectId}/teacher
func (h *SemesterHandler) AssignTeacher(w http.ResponseWriter, r *http.Request) {
	semID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid semester id"))
		return
	}
	subjectID, err := uuid.Parse(r.PathValue("subjectId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid subject id"))
		return
	}

	var req assignTeacherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
		return
	}

	var teacherID *uuid.UUID
	if req.TeacherID != "" {
		tid, err := uuid.Parse(req.TeacherID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errResp("invalid teacher_id"))
			return
		}
		teacherID = &tid
	}

	if err := h.semesterRepo.SetTeacherAssignment(r.Context(), semID, subjectID, teacherID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to assign teacher"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// --- Helpers ---

func semesterResponse(s *domain.Semester) map[string]any {
	return map[string]any{
		"id":         s.ID,
		"name":       s.Name,
		"start_date": s.StartDate,
		"end_date":   s.EndDate,
		"status":     s.Status,
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
	}
}

func errResp(msg string) map[string]string {
	return map[string]string{"error": msg}
}
