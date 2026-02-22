# Phase 07: Agent Module

## Context
- [Plan](./plan.md) | [SSE Research](./research/researcher-02-shadcn-sse-reactflow.md)
- Depends on: Phase 01, Phase 02
- Backend routes: `POST /api/v1/agent/chat` (SSE), `GET/POST /api/v1/agent/conversations`, `GET/PATCH/DELETE /api/v1/agent/conversations/{id}`, `GET /api/v1/agent/suggestions`

## Overview
- **Priority:** P2
- **Status:** pending
- **Effort:** 4h

Build AI chat interface with SSE streaming, conversation management, and inline suggestion bar.

## Key Insights
- Chat endpoint: POST body `{ conversation_id, message }` → SSE response: `data: "token"\n\n` chunks + `data: [DONE]\n\n`
- X-Conversation-ID header returned in response for new conversations
- Conversations are user-scoped (backend filters by auth claims)
- Suggestions endpoint: returns rule-based contextual suggestions (no LLM)
- Uses `@microsoft/fetch-event-source` for POST+Bearer SSE

## Requirements

### Functional
- Chat page with full-height layout:
  - Left sidebar: conversation list (title, date, delete)
  - Main area: message thread + input box
- Message rendering: user messages (right-aligned) + assistant messages (left-aligned, markdown)
- SSE streaming: tokens appear incrementally as agent responds
- New conversation: auto-created on first message if no conversation selected
- Conversation management: rename (PATCH), delete with confirmation
- Suggestion bar: horizontal bar above input showing contextual action buttons
- Loading indicator while agent is thinking

### Non-functional
- Smooth streaming without flicker
- Auto-scroll to bottom on new messages
- Input disabled while streaming

## Architecture

### File Structure
```
packages/module-agent/src/
├── index.ts
├── components/
│   ├── chat-page.tsx                # Full chat layout (sidebar + main)
│   ├── conversation-sidebar.tsx     # Conversation list + new button
│   ├── conversation-item.tsx        # Single conversation in list
│   ├── message-thread.tsx           # Scrollable message list
│   ├── message-bubble.tsx           # Single message (user/assistant)
│   ├── chat-input.tsx               # Text input + send button
│   ├── suggestion-bar.tsx           # Contextual suggestion chips
│   └── streaming-indicator.tsx      # Typing indicator during SSE
├── hooks/
│   └── use-chat-stream.ts           # SSE streaming hook (module-specific wrapper)
└── lib/
    └── markdown-renderer.tsx        # Simple markdown → React (bold, code, lists)
```

## Related Code Files

### Modify
- `web/packages/module-agent/package.json` — add deps (+ `@microsoft/fetch-event-source`)
- `web/packages/module-agent/src/index.ts`
- Shell route: `chat.index.tsx`

### Create
- All files under `packages/module-agent/src/`

## Implementation Steps

1. **Add deps**: standard deps + `@microsoft/fetch-event-source` + `react-markdown`

2. **Create `use-chat-stream.ts`** — Hook wrapping `fetchEventSource`:
   - Input: conversation_id (optional), message
   - Manages streaming state: `isStreaming`, `streamedText`, `error`
   - Accumulates tokens into `streamedText` on each `data:` event
   - On `[DONE]`: sets `isStreaming=false`, invalidates conversation messages query
   - Returns `{ sendMessage, streamedText, isStreaming, error }`

3. **Create `conversation-sidebar.tsx`** — Uses `useConversations` hook. List of conversations sorted by updated_at desc. "New Chat" button. Click selects conversation. Right-click/menu: rename, delete.

4. **Create `message-thread.tsx`** — Scrollable container. Maps messages to `<MessageBubble>`. Appends streaming message at bottom. Auto-scrolls via `useEffect` + `scrollIntoView`.

5. **Create `message-bubble.tsx`** — User: right-aligned, blue bg. Assistant: left-aligned, gray bg, rendered as markdown. Shows timestamp.

6. **Create `markdown-renderer.tsx`** — Use `react-markdown` package for rendering AI responses. Handles code blocks, tables, nested lists reliably. (Validation Session 1)
<!-- Updated: Validation Session 1 - Use react-markdown instead of custom regex renderer -->

7. **Create `chat-input.tsx`** — Textarea (auto-resize) + send button. Disabled during streaming. Enter to send (Shift+Enter for newline).

8. **Create `suggestion-bar.tsx`** — Fetches `GET /agent/suggestions`. Shows as horizontal chip/badge row above input. Click chip → sends that text as message.

9. **Create `chat-page.tsx`** — Full layout: conversation sidebar (250px) | main area (flex-1) with message thread + suggestion bar + chat input. Manages `selectedConversationId` state.

10. **Wire route, build check**

## Todo List
- [ ] Add dependencies
- [ ] Create SSE streaming hook
- [ ] Create conversation sidebar
- [ ] Create message thread + bubble components
- [ ] Create markdown renderer
- [ ] Create chat input with auto-resize
- [ ] Create suggestion bar
- [ ] Create chat page layout
- [ ] Wire route in shell
- [ ] Verify build passes

## Success Criteria
- Send message → tokens stream in real-time
- Conversation list updates with new conversation
- Switching conversations loads history
- Delete conversation works
- Suggestion chips trigger messages
- Input disabled during streaming

## Risk Assessment
- **SSE reconnect** — If connection drops mid-stream, need graceful error handling
- **Markdown XSS** — Must sanitize assistant markdown output (no raw innerHTML)
- **Large conversations** — 50+ messages may need virtualization; defer to post-MVP

## Security Considerations
- Bearer token sent via `@microsoft/fetch-event-source` headers (not URL param)
- Conversation ownership enforced server-side (UserID check)
- Markdown rendered as React elements, not innerHTML (XSS-safe)

## Next Steps
- This is the final phase. After all 7 phases, run full E2E validation.
