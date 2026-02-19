package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
)

// TeacherHandler handles teacher CRUD endpoints.
type TeacherHandler struct {
	repo domain.TeacherRepository
}

// NewTeacherHandler creates a new teacher handler.
func NewTeacherHandler(repo domain.TeacherRepository) *TeacherHandler {
	return &TeacherHandler{repo: repo}
}

type createTeacherRequest struct {
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	DepartmentID   *string  `json:"department_id,omitempty"`
	Qualifications []string `json:"qualifications"`
}

type updateTeacherRequest struct {
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	DepartmentID   *string  `json:"department_id,omitempty"`
	Qualifications []string `json:"qualifications"`
	IsActive       bool     `json:"is_active"`
}

// CreateTeacher handles POST /api/v1/teachers
func (h *TeacherHandler) CreateTeacher(w http.ResponseWriter, r *http.Request) {
	var req createTeacherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}

	now := time.Now()
	t := &domain.Teacher{
		ID:             uuid.New(),
		Name:           req.Name,
		Email:          req.Email,
		Qualifications: req.Qualifications,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if req.DepartmentID != nil {
		id, err := uuid.Parse(*req.DepartmentID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid department_id"})
			return
		}
		t.DepartmentID = &id
	}

	if err := h.repo.Save(r.Context(), t); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "teacher already exists or save failed"})
		return
	}

	writeJSON(w, http.StatusCreated, teacherResponse(t))
}

// ListTeachers handles GET /api/v1/teachers
func (h *TeacherHandler) ListTeachers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	filter := domain.TeacherFilter{}
	if raw := q.Get("department_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid department_id"})
			return
		}
		filter.DepartmentID = &id
	}
	if raw := q.Get("status"); raw != "" {
		active := raw == "active"
		filter.IsActive = &active
	}
	if qual := q.Get("qualification"); qual != "" {
		filter.Qualification = qual
	}

	teachers, total, err := h.repo.List(r.Context(), filter, offset, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list teachers"})
		return
	}

	items := make([]map[string]any, len(teachers))
	for i, t := range teachers {
		items[i] = teacherResponse(t)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// GetTeacher handles GET /api/v1/teachers/{id}
func (h *TeacherHandler) GetTeacher(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid teacher id"})
		return
	}

	t, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "teacher not found"})
		return
	}
	writeJSON(w, http.StatusOK, teacherResponse(t))
}

// UpdateTeacher handles PUT /api/v1/teachers/{id}
func (h *TeacherHandler) UpdateTeacher(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid teacher id"})
		return
	}

	existing, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "teacher not found"})
		return
	}

	var req updateTeacherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}

	existing.Name = req.Name
	existing.Email = req.Email
	existing.Qualifications = req.Qualifications
	existing.IsActive = req.IsActive
	existing.DepartmentID = nil
	if req.DepartmentID != nil {
		deptID, err := uuid.Parse(*req.DepartmentID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid department_id"})
			return
		}
		existing.DepartmentID = &deptID
	}

	if err := h.repo.Update(r.Context(), existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update teacher"})
		return
	}
	writeJSON(w, http.StatusOK, teacherResponse(existing))
}

// teacherResponse converts a Teacher entity to a JSON-safe map.
func teacherResponse(t *domain.Teacher) map[string]any {
	resp := map[string]any{
		"id":             t.ID,
		"name":           t.Name,
		"email":          t.Email,
		"qualifications": t.Qualifications,
		"is_active":      t.IsActive,
		"created_at":     t.CreatedAt,
		"updated_at":     t.UpdatedAt,
	}
	if t.DepartmentID != nil {
		resp["department_id"] = *t.DepartmentID
	} else {
		resp["department_id"] = nil
	}
	return resp
}
