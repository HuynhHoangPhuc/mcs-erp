package domain

import (
	"context"

	"github.com/google/uuid"
)

// ConversationRepository defines persistence operations for conversations and messages.
type ConversationRepository interface {
	// Conversation CRUD
	SaveConversation(ctx context.Context, c *Conversation) error
	FindConversationByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	ListConversationsByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Conversation, int, error)
	DeleteConversation(ctx context.Context, id uuid.UUID) error
	UpdateConversationTitle(ctx context.Context, id uuid.UUID, title string) error

	// Message operations
	SaveMessage(ctx context.Context, m *Message) error
	ListMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*Message, error)
}
