import { useEffect, useRef } from "react";
import { ScrollArea } from "@mcs-erp/ui";
import { useConversation } from "@mcs-erp/api-client";
import { MessageBubble } from "./message-bubble";
import { StreamingIndicator } from "./streaming-indicator";

interface MessageThreadProps {
  conversationId: string;
  streamedText: string;
  isStreaming: boolean;
}

// Scrollable message history with live streaming bubble appended at bottom.
export function MessageThread({ conversationId, streamedText, isStreaming }: MessageThreadProps) {
  const { data: conversation } = useConversation(conversationId);
  const bottomRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when messages or streamed text changes.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [conversation?.messages, streamedText]);

  const messages = conversation?.messages ?? [];
  const showStreaming = isStreaming || streamedText.length > 0;

  return (
    <ScrollArea className="flex-1 px-4">
      <div className="py-4">
        {messages.map((message) => (
          <MessageBubble key={message.id} message={message} />
        ))}

        {/* Live streaming bubble */}
        {showStreaming && (
          <div className="flex w-full mb-4 justify-start">
            <div className="max-w-[75%] flex flex-col gap-1 items-start">
              <div className="rounded-2xl rounded-bl-sm px-4 py-2.5 text-sm leading-relaxed bg-muted text-foreground">
                {streamedText.length > 0 ? (
                  <p className="whitespace-pre-wrap break-words">{streamedText}</p>
                ) : (
                  <StreamingIndicator />
                )}
              </div>
            </div>
          </div>
        )}

        <div ref={bottomRef} />
      </div>
    </ScrollArea>
  );
}
