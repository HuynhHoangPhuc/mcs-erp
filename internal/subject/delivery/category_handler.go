package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/subject/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// CategoryHandler handles CRUD endpoints for subject categories.
type CategoryHandler struct {
	categoryRepo domain.CategoryRepository
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(categoryRepo domain.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

type createCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateCategory handles POST /api/v1/categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	c := &domain.Category{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	if err := h.categoryRepo.Save(r.Context(), c); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create category"})
		return
	}

	writeJSON(w, http.StatusCreated, categoryResponse(c))
}

// ListCategories handles GET /api/v1/categories
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list categories"})
		return
	}

	items := make([]map[string]any, len(categories))
	for i, c := range categories {
		items[i] = categoryResponse(c)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// GetCategory handles GET /api/v1/categories/{id}
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category id"})
		return
	}

	c, err := h.categoryRepo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, erptypes.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get category"})
		return
	}

	writeJSON(w, http.StatusOK, categoryResponse(c))
}

func categoryResponse(c *domain.Category) map[string]any {
	return map[string]any{
		"id":          c.ID,
		"name":        c.Name,
		"description": c.Description,
		"created_at":  c.CreatedAt,
	}
}
