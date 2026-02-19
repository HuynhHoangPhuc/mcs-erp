package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// SubjectHandler handles CRUD endpoints for subjects.
type SubjectHandler struct {
	subjectRepo domain.SubjectRepository
}

// NewSubjectHandler creates a new SubjectHandler.
func NewSubjectHandler(subjectRepo domain.SubjectRepository) *SubjectHandler {
	return &SubjectHandler{subjectRepo: subjectRepo}
}

type createSubjectRequest struct {
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	Description  string     `json:"description"`
	CategoryID   *uuid.UUID `json:"category_id"`
	Credits      int        `json:"credits"`
	HoursPerWeek int        `json:"hours_per_week"`
}

type updateSubjectRequest struct {
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	Description  string     `json:"description"`
	CategoryID   *uuid.UUID `json:"category_id"`
	Credits      int        `json:"credits"`
	HoursPerWeek int        `json:"hours_per_week"`
	IsActive     bool       `json:"is_active"`
}

// CreateSubject handles POST /api/v1/subjects
func (h *SubjectHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	var req createSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and code are required"})
		return
	}

	now := time.Now()
	s := &domain.Subject{
		ID:           uuid.New(),
		Name:         req.Name,
		Code:         req.Code,
		Description:  req.Description,
		CategoryID:   req.CategoryID,
		Credits:      req.Credits,
		HoursPerWeek: req.HoursPerWeek,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.subjectRepo.Save(r.Context(), s); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "subject code already exists or save failed"})
		return
	}

	writeJSON(w, http.StatusCreated, subjectResponse(s))
}

// ListSubjects handles GET /api/v1/subjects
// Optional query params: offset, limit, category_id
func (h *SubjectHandler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	categoryIDStr := r.URL.Query().Get("category_id")
	if categoryIDStr != "" {
		catID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
			return
		}
		subjects, total, err := h.subjectRepo.ListByCategory(r.Context(), catID, offset, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list subjects"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": subjectList(subjects), "total": total})
		return
	}

	subjects, total, err := h.subjectRepo.List(r.Context(), offset, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list subjects"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": subjectList(subjects), "total": total})
}

// GetSubject handles GET /api/v1/subjects/{id}
func (h *SubjectHandler) GetSubject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subject id"})
		return
	}

	s, err := h.subjectRepo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subject not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get subject"})
		return
	}

	writeJSON(w, http.StatusOK, subjectResponse(s))
}

// UpdateSubject handles PUT /api/v1/subjects/{id}
func (h *SubjectHandler) UpdateSubject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid subject id"})
		return
	}

	var req updateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and code are required"})
		return
	}

	s, err := h.subjectRepo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subject not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get subject"})
		return
	}

	s.Name = req.Name
	s.Code = req.Code
	s.Description = req.Description
	s.CategoryID = req.CategoryID
	s.Credits = req.Credits
	s.HoursPerWeek = req.HoursPerWeek
	s.IsActive = req.IsActive

	if err := h.subjectRepo.Update(r.Context(), s); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update subject"})
		return
	}

	writeJSON(w, http.StatusOK, subjectResponse(s))
}

// subjectResponse converts a Subject to a map for JSON serialization.
func subjectResponse(s *domain.Subject) map[string]any {
	return map[string]any{
		"id":             s.ID,
		"name":           s.Name,
		"code":           s.Code,
		"description":    s.Description,
		"category_id":    s.CategoryID,
		"credits":        s.Credits,
		"hours_per_week": s.HoursPerWeek,
		"is_active":      s.IsActive,
		"created_at":     s.CreatedAt,
		"updated_at":     s.UpdatedAt,
	}
}

func subjectList(subjects []*domain.Subject) []map[string]any {
	items := make([]map[string]any, len(subjects))
	for i, s := range subjects {
		items[i] = subjectResponse(s)
	}
	return items
}
