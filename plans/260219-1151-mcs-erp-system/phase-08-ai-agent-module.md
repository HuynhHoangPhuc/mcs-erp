# Phase 08: AI Agent Module

## Context Links
- [Parent Plan](./plan.md)
- Depends on: All domain modules (Phase 04-07) for tool registration
- [AI Research](./research/researcher-02-frontend-scheduling-ai.md)
- [Go ERP AI Patterns](../reports/researcher-260219-1151-go-erp-architecture-research.md)

## Overview
- **Date:** 2026-02-19
- **Priority:** P2
- **Status:** Complete
- **Description:** AI Agent bounded context: multi-LLM provider (Claude, OpenAI, Ollama) via langchaingo, tool registry where modules register tools at startup, chat handler with SSE streaming, inline action suggestions, conversation history.

## Key Insights
- langchaingo supports Claude, OpenAI, Ollama; provides tool-calling agent executor out of box
- Tool-calling pattern: LLM decides which function to call, Go executes, feeds result back
- SSE streaming via `http.Flusher` for real-time chat responses
- Each module registers AI tools at startup — agent layer has zero knowledge of module internals
- Conversation history in Postgres (not Redis) for durability; Redis caches recent messages
- Inline suggestions: context-aware action buttons (e.g., "Schedule this teacher" on teacher detail page)
- Tool execution requires same RBAC as API endpoints — no privilege escalation via chat

## Requirements

### Functional
- Multi-provider LLM: configure primary + fallback providers (Claude primary, OpenAI fallback, Ollama for local dev)
- Tool registry: modules register tools at startup; agent dispatches tool calls
- Chat handler: SSE streaming endpoint for real-time responses
- Conversation history: persist per-user per-tenant, load recent context
- System prompt: include tenant context, user role, available tools
- Inline suggestions: API endpoint returns context-aware action suggestions for a given page/entity
- Tool execution respects user permissions (same RBAC as REST API)

### Non-Functional
- Chat response first token < 2s (SSE streaming)
- Tool execution timeout: 10s per tool call, 30s total per message
- Conversation history: keep last 50 messages per conversation
- Graceful fallback: if primary LLM fails, try fallback provider

## Architecture

```
internal/agent/
├── domain/
│   ├── conversation.go       # Conversation aggregate
│   ├── message.go            # Message entity (user/assistant/tool)
│   ├── provider.go           # LLMProvider enum + config
│   ├── tool.go               # Tool interface (wraps langchaingo tools.Tool)
│   └── repository.go         # ConversationRepository
├── application/
│   ├── commands/
│   │   ├── send_message.go        # Process user message, run agent loop
│   │   ├── create_conversation.go
│   │   └── delete_conversation.go
│   ├── queries/
│   │   ├── list_conversations.go
│   │   ├── get_conversation.go
│   │   └── get_suggestions.go     # Inline suggestions for entity/page
│   └── services/
│       ├── agent_service.go       # Orchestrates LLM + tool calling loop
│       └── provider_service.go    # Multi-provider factory + fallback
├── infrastructure/
│   ├── langchain_provider.go      # langchaingo LLM factory
│   ├── tool_registry.go           # Central tool registry
│   ├── postgres_conversation_repo.go
│   └── redis_cache.go             # Cache recent conversation messages
├── delivery/
│   ├── chat_handler.go            # POST /api/agent/chat (SSE)
│   ├── conversation_handler.go    # CRUD /api/agent/conversations
│   └── suggestion_handler.go      # GET /api/agent/suggestions
└── module.go
```

### Frontend
```
web/packages/module-agent/src/
├── routes.ts
├── components/
│   ├── chat-sidebar.tsx          # Sliding chat panel
│   ├── chat-message.tsx          # Individual message bubble
│   ├── chat-input.tsx            # Message input with send button
│   ├── chat-tool-result.tsx      # Tool call result display
│   ├── suggestion-bar.tsx        # Inline suggestion buttons
│   └── provider-indicator.tsx    # Show which LLM is active
├── hooks/
│   ├── use-chat.ts               # SSE connection + message state
│   ├── use-conversations.ts      # Conversation list/CRUD
│   └── use-suggestions.ts        # Inline suggestions for current page
└── lib/
    └── sse-client.ts             # EventSource wrapper with reconnect
```

## Related Code Files

### Files to Create

**Backend Domain:**
- `internal/agent/domain/conversation.go` — Conversation: ID, UserID, TenantID, Title, Messages []Message, CreatedAt, UpdatedAt
- `internal/agent/domain/message.go` — Message: ID, ConversationID, Role (user/assistant/tool), Content string, ToolCalls []ToolCall, CreatedAt
- `internal/agent/domain/provider.go` — `LLMProvider` enum (claude/openai/ollama), `ProviderConfig` struct
- `internal/agent/domain/tool.go` — `AgentTool` interface wrapping langchaingo `tools.Tool` + `RequiredPermission() string`
- `internal/agent/domain/repository.go` — ConversationRepository interface

**Backend Application:**
- `internal/agent/application/commands/send_message.go` — append user message, run agent service, stream response, save assistant message
- `internal/agent/application/commands/create_conversation.go`
- `internal/agent/application/commands/delete_conversation.go`
- `internal/agent/application/queries/list_conversations.go` — by user, paginated
- `internal/agent/application/queries/get_conversation.go` — with recent messages
- `internal/agent/application/queries/get_suggestions.go` — takes entity_type + entity_id, returns action suggestions
- `internal/agent/application/services/agent_service.go` — core agent loop: build messages, call LLM, handle tool calls, stream tokens
- `internal/agent/application/services/provider_service.go` — `NewLLM(provider, config)`, fallback chain

**Backend Infrastructure:**
- `internal/agent/infrastructure/langchain_provider.go` — factory for anthropic/openai/ollama LLMs via langchaingo
- `internal/agent/infrastructure/tool_registry.go` — `ToolRegistry`: Register(tool), GetTools(permissions), GetByName(name)
- `internal/agent/infrastructure/postgres_conversation_repo.go`
- `internal/agent/infrastructure/redis_cache.go` — cache last N messages per conversation

**Backend Delivery:**
- `internal/agent/delivery/chat_handler.go` — `POST /api/agent/chat` with SSE streaming response
- `internal/agent/delivery/conversation_handler.go` — CRUD `/api/agent/conversations`
- `internal/agent/delivery/suggestion_handler.go` — `GET /api/agent/suggestions?entity_type=teacher&entity_id=xxx`

**Module:**
- `internal/agent/module.go` — dependencies: `["core"]` (tools registered by other modules independently)

**SQL:**
- `sqlc/queries/agent/conversations.sql`
- `sqlc/queries/agent/messages.sql`
- `migrations/agent/000001_create_conversations_table.up.sql`
- `migrations/agent/000002_create_messages_table.up.sql`

**Frontend:**
- All files listed in Frontend section above

## Implementation Steps

1. **SQL migrations**
   ```sql
   -- conversations
   CREATE TABLE conversations (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       user_id UUID NOT NULL,
       title VARCHAR(255) NOT NULL DEFAULT 'New conversation',
       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   -- messages
   CREATE TABLE messages (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
       role VARCHAR(20) NOT NULL CHECK (role IN ('user','assistant','tool','system')),
       content TEXT NOT NULL DEFAULT '',
       tool_calls JSONB,      -- [{name, arguments, result}]
       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at);
   ```

2. **Tool registry** — central registry, modules call `Register(tool)` during bootstrap
   ```go
   type ToolRegistry struct {
       mu    sync.RWMutex
       tools map[string]AgentTool
   }
   func (r *ToolRegistry) Register(t AgentTool)
   func (r *ToolRegistry) GetTools(userPerms []string) []tools.Tool  // filter by permission
   func (r *ToolRegistry) GetByName(name string) (AgentTool, bool)
   ```

3. **Provider service** — factory pattern for LLM providers
   ```go
   func NewLLM(cfg ProviderConfig) (llms.Model, error) {
       switch cfg.Provider {
       case "claude":
           return anthropic.New(anthropic.WithModel(cfg.Model), anthropic.WithToken(cfg.APIKey))
       case "openai":
           return openai.New(openai.WithModel(cfg.Model), openai.WithToken(cfg.APIKey))
       case "ollama":
           return ollama.New(ollama.WithModel(cfg.Model), ollama.WithServerURL(cfg.BaseURL))
       }
   }
   ```
   Fallback: try primary, on error try secondary, on error return error.

4. **Agent service** — core loop:
   a. Load conversation history (last 20 messages)
   b. Build system prompt with tenant context + available tools
   c. Call LLM with streaming callback
   d. If LLM returns tool_call: execute tool (check permissions first), append result, loop back to LLM
   e. Stream assistant tokens via callback channel
   f. Save final assistant message to DB
   g. Tool call loop limit: max 5 iterations per message

5. **Chat handler with SSE** — POST (not GET, because we send message body):
   ```go
   func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
       // Parse: {conversation_id, message}
       w.Header().Set("Content-Type", "text/event-stream")
       w.Header().Set("Cache-Control", "no-cache")
       flusher := w.(http.Flusher)

       tokenCh := make(chan string, 100)
       go h.agentService.ProcessMessage(r.Context(), req, tokenCh)

       for token := range tokenCh {
           fmt.Fprintf(w, "data: %s\n\n", jsonEncode(token))
           flusher.Flush()
       }
       fmt.Fprintf(w, "data: [DONE]\n\n")
       flusher.Flush()
   }
   ```

6. **Permission-scoped tools** — when building tool list for LLM, filter by user's permissions:
   - User has `hr:teacher:read` -> include `search_teachers` tool
   - User lacks `timetable:schedule:write` -> exclude `generate_schedule` tool
   - This prevents LLM from calling tools the user cannot access

7. **System prompt template**:
   ```
   You are an AI assistant for {tenant_name}'s academic management system.
   Current user: {user_name} ({role}).
   Available actions: {tool_descriptions}.
   Always confirm before making changes. Be concise.
   ```

8. **Inline suggestions** — `GET /api/agent/suggestions?entity_type=teacher&entity_id=xxx`
   - Returns contextual actions based on entity type and state
   - Teacher page: "Check availability", "View schedule", "Assign to subject"
   - Subject page: "View prerequisites", "Find qualified teachers"
   - Timetable page: "Explain conflicts", "Optimize schedule"
   - Implementation: simple rule-based mapping (no LLM call needed)

9. **Conversation management** — CRUD endpoints, list by user, delete conversation (cascades messages)

10. **Redis cache** — cache last 20 messages per active conversation. Invalidate on new message. Reduces DB reads during chat.

11. **Frontend: SSE client** — `EventSource` wrapper with auto-reconnect, parse `data:` events, handle `[DONE]` sentinel

12. **Frontend: chat sidebar** — sliding panel (right side), conversation list, message thread, input box. Uses `use-chat` hook for SSE state.

13. **Frontend: chat message** — render markdown content, show tool call results in collapsible panels

14. **Frontend: suggestion bar** — horizontal bar of action buttons below page header. Clicking a suggestion sends it as a chat message (opens sidebar if closed).

15. **Frontend: provider indicator** — small badge showing active LLM (Claude/OpenAI/Ollama) in chat header

## Todo List
- [x] SQL migrations (conversations, messages)
- [x] Domain entities (Conversation, Message, ToolCall)
- [x] Tool registry with permission filtering
- [x] Provider service (Claude, OpenAI, Ollama factory + fallback)
- [x] Agent service (LLM + tool-calling loop)
- [x] Chat handler with SSE streaming
- [x] Permission-scoped tool filtering
- [x] System prompt builder
- [x] Inline suggestions endpoint (rule-based)
- [x] Conversation CRUD handlers
- [x] Redis message cache
- [x] Agent module registration
- [x] Frontend SSE client library
- [x] Frontend chat sidebar component
- [x] Frontend chat message rendering (markdown + tool results)
- [x] Frontend chat input
- [x] Frontend suggestion bar
- [x] Frontend provider indicator
- [x] TanStack Query hooks for conversations
- [x] Integration test: send message -> tool call -> response flow
- [x] Test: LLM provider fallback
- [x] Test: permission filtering on tools

## Success Criteria
- Chat endpoint streams tokens via SSE in real-time
- Agent correctly calls registered tools (e.g., search_teachers) and returns results
- Tool calls respect user RBAC permissions
- Provider fallback works (primary fails -> secondary succeeds)
- Conversation history persists and loads correctly
- Inline suggestions appear on entity pages
- Frontend chat sidebar works end-to-end
- Tool call results display in chat with formatted output

## Risk Assessment
- **LLM latency**: First token 1-3s for cloud providers. SSE streaming mitigates perceived latency.
- **Tool call loops**: LLM may loop calling tools indefinitely. Mitigate: max 5 tool calls per message.
- **Hallucination**: LLM may fabricate data. Mitigate: all data comes from tool results, not LLM knowledge. Tool results are ground truth.
- **Cost**: Cloud LLM calls cost money. Mitigate: Ollama for dev, token usage tracking, rate limiting per user.
- **langchaingo gaps**: May lag behind provider API changes. Mitigate: thin wrapper allows swapping to direct SDK if needed.

## Security Considerations
- Tool execution uses same RBAC as REST API — no privilege escalation
- LLM API keys stored as env vars, never exposed to frontend
- User messages stored in tenant-scoped schema
- System prompt does not expose internal architecture or sensitive config
- Rate limit chat endpoint: 20 messages/min per user
- Tool results sanitized before returning to LLM (no SQL, no internal errors)
- Destructive tools (modify_assignment, delete operations) require explicit user confirmation via chat flow

## Next Steps
- Polish and integration testing across all modules
- Docker Compose production config
- Documentation + module SDK docs
