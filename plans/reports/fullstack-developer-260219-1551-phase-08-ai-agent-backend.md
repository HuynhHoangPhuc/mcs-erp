# Phase Implementation Report

## Executed Phase
- Phase: phase-08-ai-agent-module (BACKEND ONLY)
- Plan: /Users/phuc/Developer/mcs-erp/plans/260219-1151-mcs-erp-system/
- Status: completed

## Files Modified

### New files created
| File | Lines |
|------|-------|
| `internal/agent/domain/conversation.go` | 33 |
| `internal/agent/domain/message.go` | 46 |
| `internal/agent/domain/provider.go` | 23 |
| `internal/agent/domain/tool.go` | 17 |
| `internal/agent/domain/repository.go` | 22 |
| `internal/agent/infrastructure/tool_registry.go` | 57 |
| `internal/agent/infrastructure/langchain_provider.go` | 57 |
| `internal/agent/infrastructure/postgres_conversation_repo.go` | 183 |
| `internal/agent/infrastructure/redis_cache.go` | 88 |
| `internal/agent/application/services/provider_service.go` | 30 |
| `internal/agent/application/services/agent_service.go` | 200 |
| `internal/agent/delivery/json_helpers.go` | 13 |
| `internal/agent/delivery/chat_handler.go` | 87 |
| `internal/agent/delivery/conversation_handler.go` | 143 |
| `internal/agent/delivery/suggestion_handler.go` | 68 |
| `internal/agent/module.go` | 74 |
| `migrations/agent/000001_create_conversations_table.up.sql` | 9 |
| `migrations/agent/000002_create_messages_table.up.sql` | 11 |

### Modified files
| File | Change |
|------|--------|
| `internal/core/domain/permission.go` | Added `PermAgentChatRead`, `PermAgentChatWrite`; added both to `AllPermissions()` |
| `internal/platform/config/config.go` | Added `LLMProvider`, `LLMModel`, `LLMAPIKey`, `LLMFallbackProvider`, `LLMFallbackModel`, `LLMFallbackAPIKey`, `OllamaURL` fields + env loading |
| `go.mod` / `go.sum` | Added `github.com/tmc/langchaingo v0.1.14`, `github.com/redis/go-redis/v9 v9.18.0` |

## Tasks Completed
- [x] SQL migrations (conversations, messages tables)
- [x] Domain entities (Conversation, Message, ToolCall, Role)
- [x] Tool registry with permission filtering (thread-safe, startup panic on duplicate)
- [x] Provider service (Claude, OpenAI, Ollama factory + fallback)
- [x] Agent service (LLM + tool-calling loop, max 5 iterations, 10s tool timeout)
- [x] Chat handler with SSE streaming (POST /api/v1/agent/chat)
- [x] Permission-scoped tool filtering (RBAC enforced before every tool call)
- [x] System prompt builder (includes user context + available tool descriptions)
- [x] Inline suggestions endpoint (rule-based, GET /api/v1/agent/suggestions)
- [x] Conversation CRUD handlers (list, get, create, update title, delete)
- [x] Redis message cache (last 20 messages, 2h TTL, pipeline RPUSH+LTRIM+EXPIRE)
- [x] Agent module registration (Name=agent, Dependencies=["core"])
- [x] New permissions PermAgentChatRead + PermAgentChatWrite added
- [x] Config updated with all LLM env vars

## Tests Status
- Type check: pass (`go build ./...` clean, no errors)
- Unit tests: not written (phase scope is backend implementation only)
- Integration tests: not written

## Issues Encountered
1. `langchaingo` `ContentChoice.Content` is a plain `string`, not a struct with `Parts` — fixed by reading actual pkg.go.dev docs for v0.1.14.
2. `FunctionCall.Arguments` is already a JSON string, not a struct — `json.Marshal` call removed.
3. Initial import alias collision in `tool_registry.go` between `internal/agent/domain` and `internal/core/domain` — fixed with `agentdomain`/`coredomain` aliases.
4. `go.sum` missing entries after `go get langchaingo` — resolved with `go get github.com/tmc/langchaingo/llms@v0.1.14` + `go mod tidy`.

## Architecture Notes
- `ToolRegistry` is a shared singleton passed into `NewModule`; other modules call `registry.Register(tool)` at bootstrap
- Agent loop: load history (Redis first, DB fallback) → build system prompt → stream LLM → execute tools (RBAC-gated, 10s timeout each) → persist assistant message → invalidate nothing (cache append only)
- SSE chat uses `POST` not `GET` (message body required); auto-creates conversation when `conversation_id` omitted
- No frontend code implemented (out of scope per instructions)
- Module NOT wired into `main.go` per instructions

## Next Steps
- Wire `agent.NewModule(...)` into `cmd/server/main.go` (left to user)
- Register domain-specific tools from HR/Subject/Room/Timetable modules via `registry.Register(tool)`
- Add rate limiting middleware (20 msg/min per user) on chat endpoint
- Integration test: send message → tool call → streamed response
