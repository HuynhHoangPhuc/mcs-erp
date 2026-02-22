// Agent module types matching backend chat/conversation handlers.

export type MessageRole = "user" | "assistant" | "tool" | "system";

export interface ToolCall {
  name: string;
  arguments: string;
  result: string;
}

export interface Message {
  id: string;
  conversation_id: string;
  role: MessageRole;
  content: string;
  tool_calls?: ToolCall[];
  created_at: string;
}

export interface Conversation {
  id: string;
  user_id: string;
  title: string;
  created_at: string;
  updated_at: string;
}

export interface ConversationWithMessages extends Conversation {
  messages: Message[];
}

export interface ChatRequest {
  conversation_id?: string;
  message: string;
}

export interface UpdateConversationRequest {
  title: string;
}

export interface Suggestion {
  text: string;
  category: string;
}
