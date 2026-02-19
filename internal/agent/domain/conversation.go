package domain

import (
	"time"

	"github.com/google/uuid"
)

// Conversation is the root aggregate for an AI chat session.
// One conversation per user per topic; messages are owned by the conversation.
type Conversation struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ConversationWithMessages embeds a conversation with its loaded messages.
type ConversationWithMessages struct {
	Conversation
	Messages []*Message `json:"messages"`
}

// NewConversation creates a new conversation with sensible defaults.
func NewConversation(userID uuid.UUID, title string) *Conversation {
	now := time.Now()
	if title == "" {
		title = "New conversation"
	}
	return &Conversation{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
