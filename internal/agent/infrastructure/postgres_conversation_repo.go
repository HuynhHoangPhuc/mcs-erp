package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/database"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/tenant"
	"github.com/HuynhHoangPhuc/mcs-erp/pkg/erptypes"
)

// PostgresConversationRepo implements domain.ConversationRepository using pgx.
type PostgresConversationRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresConversationRepo creates a new conversation repository.
func NewPostgresConversationRepo(pool *pgxpool.Pool) *PostgresConversationRepo {
	return &PostgresConversationRepo{pool: pool}
}

func (r *PostgresConversationRepo) schema(ctx context.Context) (string, error) {
	return tenant.FromContext(ctx)
}

// SaveConversation inserts a new conversation row.
func (r *PostgresConversationRepo) SaveConversation(ctx context.Context, c *domain.Conversation) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}
	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO conversations (id, user_id, title, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			c.ID, c.UserID, c.Title, c.CreatedAt, c.UpdatedAt,
		)
		return err
	})
}

// FindConversationByID retrieves a conversation by its primary key.
func (r *PostgresConversationRepo) FindConversationByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var c domain.Conversation
	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, user_id, title, created_at, updated_at
			 FROM conversations WHERE id = $1`,
			id,
		).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, erptypes.ErrNotFound
		}
		return nil, fmt.Errorf("find conversation by id: %w", err)
	}
	return &c, nil
}

// ListConversationsByUser returns paginated conversations for a user, newest first.
func (r *PostgresConversationRepo) ListConversationsByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.Conversation, int, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, 0, err
	}

	var conversations []*domain.Conversation
	var total int

	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM conversations WHERE user_id = $1`, userID,
		).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx,
			`SELECT id, user_id, title, created_at, updated_at
			 FROM conversations WHERE user_id = $1
			 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`,
			userID, limit, offset,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var c domain.Conversation
			if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
				return err
			}
			conversations = append(conversations, &c)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list conversations: %w", err)
	}
	return conversations, total, nil
}

// DeleteConversation removes a conversation and cascades to messages.
func (r *PostgresConversationRepo) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}
	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM conversations WHERE id = $1`, id)
		return err
	})
}

// UpdateConversationTitle updates a conversation's title and bumps updated_at.
func (r *PostgresConversationRepo) UpdateConversationTitle(ctx context.Context, id uuid.UUID, title string) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}
	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE conversations SET title = $2, updated_at = $3 WHERE id = $1`,
			id, title, time.Now(),
		)
		return err
	})
}

// SaveMessage inserts a message row, serialising ToolCalls to JSONB.
func (r *PostgresConversationRepo) SaveMessage(ctx context.Context, m *domain.Message) error {
	schema, err := r.schema(ctx)
	if err != nil {
		return err
	}

	var toolCallsJSON []byte
	if len(m.ToolCalls) > 0 {
		toolCallsJSON, err = json.Marshal(m.ToolCalls)
		if err != nil {
			return fmt.Errorf("marshal tool_calls: %w", err)
		}
	}

	return database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO messages (id, conversation_id, role, content, tool_calls, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			m.ID, m.ConversationID, string(m.Role), m.Content, toolCallsJSON, m.CreatedAt,
		)
		return err
	})
}

// ListMessages returns the most recent `limit` messages for a conversation, ordered oldest-first.
func (r *PostgresConversationRepo) ListMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*domain.Message, error) {
	schema, err := r.schema(ctx)
	if err != nil {
		return nil, err
	}

	var messages []*domain.Message

	err = database.WithTenantTx(ctx, r.pool, schema, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT id, conversation_id, role, content, tool_calls, created_at
			 FROM (
			     SELECT id, conversation_id, role, content, tool_calls, created_at
			     FROM messages WHERE conversation_id = $1
			     ORDER BY created_at DESC LIMIT $2
			 ) sub
			 ORDER BY created_at ASC`,
			conversationID, limit,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m domain.Message
			var roleStr string
			var toolCallsJSON []byte

			if err := rows.Scan(&m.ID, &m.ConversationID, &roleStr, &m.Content, &toolCallsJSON, &m.CreatedAt); err != nil {
				return err
			}
			m.Role = domain.Role(roleStr)

			if len(toolCallsJSON) > 0 {
				if err := json.Unmarshal(toolCallsJSON, &m.ToolCalls); err != nil {
					return fmt.Errorf("unmarshal tool_calls: %w", err)
				}
			}
			messages = append(messages, &m)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return messages, nil
}

// Ensure interface compliance at compile time.
var _ domain.ConversationRepository = (*PostgresConversationRepo)(nil)
