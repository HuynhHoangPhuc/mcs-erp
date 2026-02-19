package delivery

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
)

// RoleHandler handles role CRUD endpoints.
type RoleHandler struct {
	roleRepo domain.RoleRepository
}

func NewRoleHandler(roleRepo domain.RoleRepository) *RoleHandler {
	return &RoleHandler{roleRepo: roleRepo}
}

type createRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description"`
}

// CreateRole handles POST /api/v1/roles
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	role := &domain.Role{
		ID:          uuid.New(),
		Name:        req.Name,
		Permissions: req.Permissions,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	if err := h.roleRepo.Save(r.Context(), role); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "role already exists or save failed"})
		return
	}

	writeJSON(w, http.StatusCreated, role)
}

// ListRoles handles GET /api/v1/roles
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roleRepo.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list roles"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": roles})
}

// GetRole handles GET /api/v1/roles/{id}
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	role, err := h.roleRepo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "role not found"})
		return
	}

	writeJSON(w, http.StatusOK, role)
}

// DeleteRole handles DELETE /api/v1/roles/{id}
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	if err := h.roleRepo.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete role"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "role deleted"})
}
