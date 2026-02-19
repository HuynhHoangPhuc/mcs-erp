package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/application/services"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

// ChatHandler handles the SSE chat endpoint.
type ChatHandler struct {
	agentSvc *services.AgentService
	convRepo domain.ConversationRepository
}

// NewChatHandler creates a ChatHandler.
func NewChatHandler(agentSvc *services.AgentService, convRepo domain.ConversationRepository) *ChatHandler {
	return &ChatHandler{agentSvc: agentSvc, convRepo: convRepo}
}

type chatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

// HandleChat handles POST /api/v1/agent/chat with SSE streaming response.
// The client sends JSON body; the server streams tokens via text/event-stream.
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.UserFromContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}

	// Resolve or create conversation.
	convID, err := h.resolveConversation(r, claims, req.ConversationID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Verify SSE support.
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming not supported"})
		return
	}

	// Set SSE headers before writing any body.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Conversation-ID", convID.String())
	w.WriteHeader(http.StatusOK)

	tokenCh := make(chan string, 100)
	go h.agentSvc.ProcessMessage(r.Context(), services.ProcessMessageRequest{
		ConversationID: convID,
		UserMessage:    req.Message,
	}, tokenCh)

	for token := range tokenCh {
		encoded, _ := json.Marshal(token)
		fmt.Fprintf(w, "data: %s\n\n", encoded)
		flusher.Flush()
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// resolveConversation returns an existing conversation ID or creates a new conversation.
func (h *ChatHandler) resolveConversation(r *http.Request, claims *auth.Claims, rawID string) (uuid.UUID, error) {
	if rawID != "" {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid conversation_id")
		}
		// Verify conversation exists (repo uses tenant from context).
		if _, err := h.convRepo.FindConversationByID(r.Context(), id); err != nil {
			return uuid.Nil, fmt.Errorf("conversation not found")
		}
		return id, nil
	}

	// Auto-create a new conversation.
	conv := domain.NewConversation(claims.UserID, "New conversation")
	if err := h.convRepo.SaveConversation(r.Context(), conv); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create conversation")
	}
	return conv.ID, nil
}
