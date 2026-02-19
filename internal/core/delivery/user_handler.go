package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

// UserHandler handles user CRUD endpoints.
type UserHandler struct {
	userRepo   domain.UserRepository
	roleRepo   domain.RoleRepository
	lookupRepo domain.UsersLookupRepository
}

func NewUserHandler(userRepo domain.UserRepository, roleRepo domain.RoleRepository, lookupRepo domain.UsersLookupRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo, roleRepo: roleRepo, lookupRepo: lookupRepo}
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, password, and name required"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.userRepo.Save(r.Context(), user); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists or save failed"})
		return
	}

	// Register in public.users_lookup for cross-tenant login resolution
	claims, _ := auth.UserFromContext(r.Context())
	if claims != nil {
		h.lookupRepo.Upsert(r.Context(), user.Email, claims.TenantID)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id": user.ID, "email": user.Email, "name": user.Name, "is_active": user.IsActive,
	})
}

// ListUsers handles GET /api/v1/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	users, total, err := h.userRepo.List(r.Context(), offset, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}

	items := make([]map[string]any, len(users))
	for i, u := range users {
		items[i] = map[string]any{
			"id": u.ID, "email": u.Email, "name": u.Name,
			"is_active": u.IsActive, "created_at": u.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// GetUser handles GET /api/v1/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id": user.ID, "email": user.Email, "name": user.Name,
		"is_active": user.IsActive, "created_at": user.CreatedAt,
	})
}

type assignRoleRequest struct {
	RoleID string `json:"role_id"`
}

// AssignRole handles POST /api/v1/users/{id}/roles
func (h *UserHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	if err := h.roleRepo.AssignRoleToUser(r.Context(), userID, roleID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to assign role"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "role assigned"})
}
