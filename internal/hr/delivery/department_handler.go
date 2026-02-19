package delivery

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/hr/domain"
)

// DepartmentHandler handles department CRUD endpoints.
type DepartmentHandler struct {
	repo domain.DepartmentRepository
}

// NewDepartmentHandler creates a new department handler.
func NewDepartmentHandler(repo domain.DepartmentRepository) *DepartmentHandler {
	return &DepartmentHandler{repo: repo}
}

type createDepartmentRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	HeadTeacherID *string `json:"head_teacher_id,omitempty"`
}

type updateDepartmentRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	HeadTeacherID *string `json:"head_teacher_id,omitempty"`
}

// CreateDepartment handles POST /api/v1/departments
func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req createDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	d := &domain.Department{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}
	if req.HeadTeacherID != nil {
		id, err := uuid.Parse(*req.HeadTeacherID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid head_teacher_id"})
			return
		}
		d.HeadTeacherID = &id
	}

	if err := h.repo.Save(r.Context(), d); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "department already exists or save failed"})
		return
	}
	writeJSON(w, http.StatusCreated, deptResponse(d))
}

// ListDepartments handles GET /api/v1/departments
func (h *DepartmentHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	depts, err := h.repo.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list departments"})
		return
	}

	items := make([]map[string]any, len(depts))
	for i, d := range depts {
		items[i] = deptResponse(d)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

// GetDepartment handles GET /api/v1/departments/{id}
func (h *DepartmentHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid department id"})
		return
	}

	d, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "department not found"})
		return
	}
	writeJSON(w, http.StatusOK, deptResponse(d))
}

// UpdateDepartment handles PUT /api/v1/departments/{id}
func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid department id"})
		return
	}

	existing, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "department not found"})
		return
	}

	var req updateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.HeadTeacherID = nil
	if req.HeadTeacherID != nil {
		htID, err := uuid.Parse(*req.HeadTeacherID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid head_teacher_id"})
			return
		}
		existing.HeadTeacherID = &htID
	}

	if err := h.repo.Update(r.Context(), existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update department"})
		return
	}
	writeJSON(w, http.StatusOK, deptResponse(existing))
}

// DeleteDepartment handles DELETE /api/v1/departments/{id}
func (h *DepartmentHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid department id"})
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete department"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "department deleted"})
}

// deptResponse converts a Department entity to a JSON-safe map.
func deptResponse(d *domain.Department) map[string]any {
	resp := map[string]any{
		"id":          d.ID,
		"name":        d.Name,
		"description": d.Description,
		"created_at":  d.CreatedAt,
	}
	if d.HeadTeacherID != nil {
		resp["head_teacher_id"] = *d.HeadTeacherID
	} else {
		resp["head_teacher_id"] = nil
	}
	return resp
}
