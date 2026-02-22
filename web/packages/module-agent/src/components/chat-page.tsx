import { useState, useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useChatSSE, queryKeys } from "@mcs-erp/api-client";
import { ConversationSidebar } from "./conversation-sidebar";
import { MessageThread } from "./message-thread";
import { SuggestionBar } from "./suggestion-bar";
import { ChatInput } from "./chat-input";

// Full-height chat layout: sidebar + message thread + suggestion bar + input.
export function ChatPage() {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const { sendMessage, streamedText, isStreaming } = useChatSSE();
  const qc = useQueryClient();

  const handleSend = useCallback(async (message: string) => {
    const newId = await sendMessage(selectedId ?? undefined, message);

    // New conversation created: select it and refresh sidebar list.
    if (newId) {
      setSelectedId(newId);
      qc.invalidateQueries({ queryKey: queryKeys.conversations.all });
      qc.invalidateQueries({ queryKey: queryKeys.conversations.detail(newId) });
      return;
    }

    // Existing conversation: refresh its messages and the list (title may update).
    if (selectedId) {
      qc.invalidateQueries({ queryKey: queryKeys.conversations.detail(selectedId) });
      qc.invalidateQueries({ queryKey: queryKeys.conversations.all });
    }
  }, [selectedId, sendMessage, qc]);

  const handleNew = useCallback(() => {
    setSelectedId(null);
  }, []);

  return (
    <div className="flex h-full w-full overflow-hidden">
      <ConversationSidebar
        selectedId={selectedId}
        onSelect={setSelectedId}
        onNew={handleNew}
      />

      <div className="flex flex-col flex-1 overflow-hidden">
        {selectedId ? (
          <>
            <MessageThread
              conversationId={selectedId}
              streamedText={streamedText}
              isStreaming={isStreaming}
            />
            <SuggestionBar onSend={handleSend} disabled={isStreaming} />
            <ChatInput onSend={handleSend} disabled={isStreaming} />
          </>
        ) : (
          <div className="flex flex-col flex-1 overflow-hidden">
            <div className="flex flex-1 items-center justify-center text-muted-foreground">
              <div className="text-center space-y-2">
                <p className="text-lg font-medium">Start a new conversation</p>
                <p className="text-sm">Ask the AI assistant anything about your ERP data.</p>
              </div>
            </div>
            <SuggestionBar onSend={handleSend} disabled={isStreaming} />
            <ChatInput onSend={handleSend} disabled={isStreaming} />
          </div>
        )}
      </div>
    </div>
  );
}
