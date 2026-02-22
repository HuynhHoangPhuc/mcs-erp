import { useState, useCallback, useRef } from "react";
import { fetchEventSource } from "@microsoft/fetch-event-source";
import { getAccessToken } from "../lib/api-client";

const API_BASE_URL = typeof import.meta !== "undefined"
  ? (import.meta as any).env?.VITE_API_BASE_URL ?? ""
  : "";

interface UseChatSSEReturn {
  sendMessage: (conversationId: string | undefined, message: string) => Promise<string | undefined>;
  streamedText: string;
  isStreaming: boolean;
  error: string | null;
  abort: () => void;
}

// SSE streaming hook for POST-based chat endpoint with Bearer auth.
export function useChatSSE(): UseChatSSEReturn {
  const [streamedText, setStreamedText] = useState("");
  const [isStreaming, setIsStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const abort = useCallback(() => {
    abortRef.current?.abort();
    setIsStreaming(false);
  }, []);

  const sendMessage = useCallback(async (conversationId: string | undefined, message: string) => {
    abortRef.current?.abort();
    const ctrl = new AbortController();
    abortRef.current = ctrl;

    setStreamedText("");
    setIsStreaming(true);
    setError(null);

    let newConversationId: string | undefined;

    try {
      await fetchEventSource(`${API_BASE_URL}/api/v1/agent/chat`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(getAccessToken() ? { Authorization: `Bearer ${getAccessToken()}` } : {}),
        },
        body: JSON.stringify({
          conversation_id: conversationId ?? "",
          message,
        }),
        signal: ctrl.signal,
        onopen: async (response: Response) => {
          if (!response.ok) {
            throw new Error(`Chat failed: ${response.status}`);
          }
          const convId = response.headers.get("X-Conversation-ID");
          if (convId) newConversationId = convId;
        },
        onmessage: (ev: { data: string }) => {
          if (ev.data === "[DONE]") {
            setIsStreaming(false);
            return;
          }
          try {
            const token = JSON.parse(ev.data) as string;
            setStreamedText((prev: string) => prev + token);
          } catch {
            setStreamedText((prev: string) => prev + ev.data);
          }
        },
        onerror: (err: unknown) => {
          setError(err instanceof Error ? err.message : "Stream error");
          setIsStreaming(false);
          throw err; // Stop retrying
        },
        onclose: () => {
          setIsStreaming(false);
        },
      });
    } catch (e) {
      if ((e as Error).name !== "AbortError") {
        setError(e instanceof Error ? e.message : "Unknown error");
      }
      setIsStreaming(false);
    }

    return newConversationId;
  }, []);

  return { sendMessage, streamedText, isStreaming, error, abort };
}
