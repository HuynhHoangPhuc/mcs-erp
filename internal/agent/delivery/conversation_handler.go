package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

// ConversationHandler handles conversation CRUD endpoints.
type ConversationHandler struct {
	repo domain.ConversationRepository
}

// NewConversationHandler creates a ConversationHandler.
func NewConversationHandler(repo domain.ConversationRepository) *ConversationHandler {
	return &ConversationHandler{repo: repo}
}

// ListConversations handles GET /api/v1/agent/conversations
func (h *ConversationHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.UserFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	q := r.URL.Query()
	offset, _ := strconv.Atoi(q.Get("offset"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	convs, total, err := h.repo.ListConversationsByUser(r.Context(), claims.UserID, offset, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list conversations"})
		return
	}

	items := make([]map[string]any, len(convs))
	for i, c := range convs {
		items[i] = conversationResponse(c)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// GetConversation handles GET /api/v1/agent/conversations/{id}
func (h *ConversationHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.UserFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid conversation id"})
		return
	}

	conv, err := h.repo.FindConversationByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
		return
	}
	if conv.UserID != claims.UserID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	msgs, err := h.repo.ListMessages(r.Context(), id, 50)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load messages"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"conversation": conversationResponse(conv),
		"messages":     msgs,
	})
}

// CreateConversation handles POST /api/v1/agent/conversations
func (h *ConversationHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.UserFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	conv := domain.NewConversation(claims.UserID, body.Title)
	if err := h.repo.SaveConversation(r.Context(), conv); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create conversation"})
		return
	}

	writeJSON(w, http.StatusCreated, conversationResponse(conv))
}

// UpdateConversation handles PATCH /api/v1/agent/conversations/{id}
func (h *ConversationHandler) UpdateConversation(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.UserFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid conversation id"})
		return
	}

	conv, err := h.repo.FindConversationByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
		return
	}
	if conv.UserID != claims.UserID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	if err := h.repo.UpdateConversationTitle(r.Context(), id, body.Title); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update conversation"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DeleteConversation handles DELETE /api/v1/agent/conversations/{id}
func (h *ConversationHandler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.UserFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid conversation id"})
		return
	}

	conv, err := h.repo.FindConversationByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
		return
	}
	if conv.UserID != claims.UserID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	if err := h.repo.DeleteConversation(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete conversation"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func conversationResponse(c *domain.Conversation) map[string]any {
	return map[string]any{
		"id":         c.ID,
		"user_id":    c.UserID,
		"title":      c.Title,
		"created_at": c.CreatedAt,
		"updated_at": c.UpdatedAt,
	}
}
