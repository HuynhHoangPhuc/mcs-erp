package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/infrastructure"
	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

const (
	maxToolIterations = 5
	toolTimeout       = 10 * time.Second
	historyLimit      = 20 // messages loaded from cache/DB per request
)

// ProcessMessageRequest carries the input for a single chat turn.
type ProcessMessageRequest struct {
	ConversationID uuid.UUID
	UserMessage    string
}

// AgentService orchestrates the LLM + tool-calling loop for a single chat turn.
type AgentService struct {
	repo         domain.ConversationRepository
	cache        *infrastructure.RedisMessageCache
	registry     *infrastructure.ToolRegistry
	providerSvc  *ProviderService
}

// NewAgentService creates a wired AgentService.
func NewAgentService(
	repo domain.ConversationRepository,
	cache *infrastructure.RedisMessageCache,
	registry *infrastructure.ToolRegistry,
	providerSvc *ProviderService,
) *AgentService {
	return &AgentService{
		repo:        repo,
		cache:       cache,
		registry:    registry,
		providerSvc: providerSvc,
	}
}

// ProcessMessage runs the full agent loop: persist user message → call LLM with tool loop →
// stream tokens via tokenCh → persist assistant message. Closes tokenCh when done.
func (s *AgentService) ProcessMessage(ctx context.Context, req ProcessMessageRequest, tokenCh chan<- string) {
	defer close(tokenCh)

	claims, err := auth.UserFromContext(ctx)
	if err != nil {
		sendError(tokenCh, "unauthorized")
		return
	}

	// 1. Persist user message.
	userMsg := domain.NewMessage(req.ConversationID, domain.RoleUser, req.UserMessage)
	if err := s.repo.SaveMessage(ctx, userMsg); err != nil {
		sendError(tokenCh, "failed to save user message")
		return
	}
	if s.cache != nil {
		_ = s.cache.AddMessage(ctx, req.ConversationID, userMsg)
	}

	// 2. Load recent history.
	history, err := s.loadHistory(ctx, req.ConversationID)
	if err != nil {
		sendError(tokenCh, "failed to load conversation history")
		return
	}

	// 3. Get permitted tools.
	permittedTools := s.registry.GetTools(claims.Permissions)
	systemPrompt := buildSystemPrompt(claims, permittedTools)

	// 4. Build langchaingo messages.
	lc := buildLangchainMessages(systemPrompt, history)

	// 5. Get LLM.
	llm, err := s.providerSvc.GetLLM()
	if err != nil {
		sendError(tokenCh, "LLM provider unavailable")
		return
	}

	// 6. Tool-calling loop (max maxToolIterations rounds).
	var assistantContent strings.Builder
	var allToolCalls []domain.ToolCall

	for iteration := 0; iteration < maxToolIterations; iteration++ {
		var streamBuf strings.Builder

		opts := []llms.CallOption{
			llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
				token := string(chunk)
				streamBuf.WriteString(token)
				assistantContent.WriteString(token)
				select {
				case tokenCh <- token:
				case <-ctx.Done():
					return ctx.Err()
				}
				return nil
			}),
		}

		resp, err := llm.GenerateContent(ctx, lc, opts...)
		if err != nil {
			sendError(tokenCh, "LLM call failed")
			return
		}

		// Check if LLM requested tool calls.
		toolCallRequests := extractToolCalls(resp)
		if len(toolCallRequests) == 0 {
			// No tool calls — we are done.
			break
		}

		// 7. Execute each requested tool.
		toolResults := s.executeTools(ctx, toolCallRequests, claims.Permissions)
		allToolCalls = append(allToolCalls, toolResults...)

		// Append tool results back into the message chain for next LLM turn.
		lc = appendToolResults(lc, toolCallRequests, toolResults)
	}

	// 8. Persist assistant message.
	assistantMsg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: req.ConversationID,
		Role:           domain.RoleAssistant,
		Content:        assistantContent.String(),
		ToolCalls:      allToolCalls,
		CreatedAt:      time.Now(),
	}
	if err := s.repo.SaveMessage(ctx, assistantMsg); err != nil {
		// Non-fatal: message was already streamed to client.
		_ = err
	}
	if s.cache != nil {
		_ = s.cache.AddMessage(ctx, req.ConversationID, assistantMsg)
	}
}

// loadHistory returns recent messages from Redis cache or falls back to DB.
func (s *AgentService) loadHistory(ctx context.Context, conversationID uuid.UUID) ([]*domain.Message, error) {
	if s.cache != nil {
		cached, err := s.cache.GetMessages(ctx, conversationID)
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
	}
	return s.repo.ListMessages(ctx, conversationID, historyLimit)
}

// executeTools runs each tool call, checking permissions before execution.
func (s *AgentService) executeTools(ctx context.Context, requests []toolCallRequest, userPerms []string) []domain.ToolCall {
	results := make([]domain.ToolCall, 0, len(requests))

	for _, req := range requests {
		tool, ok := s.registry.GetByName(req.name)
		if !ok {
			results = append(results, domain.ToolCall{
				Name:      req.name,
				Arguments: req.args,
				Result:    `{"error":"tool not found"}`,
			})
			continue
		}

		// Enforce RBAC before executing tool.
		if perm := tool.RequiredPermission(); perm != "" && !coredomain.HasPermission(userPerms, perm) {
			results = append(results, domain.ToolCall{
				Name:      req.name,
				Arguments: req.args,
				Result:    `{"error":"permission denied"}`,
			})
			continue
		}

		toolCtx, cancel := context.WithTimeout(ctx, toolTimeout)
		result, err := tool.Call(toolCtx, req.args)
		cancel()
		if err != nil {
			result = fmt.Sprintf(`{"error":%q}`, err.Error())
		}

		results = append(results, domain.ToolCall{
			Name:      req.name,
			Arguments: req.args,
			Result:    result,
		})
	}
	return results
}

// --- helpers ---

type toolCallRequest struct {
	id   string // tool call ID from the LLM response
	name string
	args string
}

func buildSystemPrompt(claims *auth.Claims, tools []domain.AgentTool) string {
	var toolDescs strings.Builder
	for _, t := range tools {
		toolDescs.WriteString(fmt.Sprintf("- %s: %s\n", t.Name(), t.Description()))
	}

	return fmt.Sprintf(
		"You are an AI assistant for the academic management system.\n"+
			"Current user: %s (tenant: %s).\n"+
			"Available actions:\n%s\n"+
			"Always confirm before making changes. Be concise.",
		claims.Email, claims.TenantID, toolDescs.String(),
	)
}

func buildLangchainMessages(systemPrompt string, history []*domain.Message) []llms.MessageContent {
	msgs := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextPart(systemPrompt)}},
	}
	for _, m := range history {
		role := lcRole(m.Role)
		msgs = append(msgs, llms.MessageContent{
			Role:  role,
			Parts: []llms.ContentPart{llms.TextPart(m.Content)},
		})
	}
	return msgs
}

func lcRole(r domain.Role) llms.ChatMessageType {
	switch r {
	case domain.RoleUser:
		return llms.ChatMessageTypeHuman
	case domain.RoleAssistant:
		return llms.ChatMessageTypeAI
	case domain.RoleSystem:
		return llms.ChatMessageTypeSystem
	default:
		return llms.ChatMessageTypeHuman
	}
}

func extractToolCalls(resp *llms.ContentResponse) []toolCallRequest {
	if resp == nil {
		return nil
	}
	var calls []toolCallRequest
	for _, choice := range resp.Choices {
		for _, tc := range choice.ToolCalls {
			if tc.FunctionCall == nil {
				continue
			}
			calls = append(calls, toolCallRequest{
				id:   tc.ID,
				name: tc.FunctionCall.Name,
				args: tc.FunctionCall.Arguments, // already a JSON string
			})
		}
	}
	return calls
}

func appendToolResults(msgs []llms.MessageContent, _ []toolCallRequest, results []domain.ToolCall) []llms.MessageContent {
	for _, tc := range results {
		msgs = append(msgs, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{llms.ToolCallResponse{
				Name:    tc.Name,
				Content: tc.Result,
			}},
		})
	}
	return msgs
}

func sendError(ch chan<- string, msg string) {
	select {
	case ch <- fmt.Sprintf(`{"error":%q}`, msg):
	default:
	}
}
