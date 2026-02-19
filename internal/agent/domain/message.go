package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role defines who authored a message in a conversation.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
	RoleSystem    Role = "system"
)

// ToolCall records a single tool invocation by the LLM within a message.
type ToolCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON-encoded arguments
	Result    string `json:"result"`    // JSON-encoded result
}

// Message is an immutable entity within a Conversation.
type Message struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	Role           Role       `json:"role"`
	Content        string     `json:"content"`
	ToolCalls      []ToolCall `json:"tool_calls,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// NewMessage creates a new message entity.
func NewMessage(conversationID uuid.UUID, role Role, content string) *Message {
	return &Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		CreatedAt:      time.Now(),
	}
}
