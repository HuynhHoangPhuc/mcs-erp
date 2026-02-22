import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../lib/api-client";
import { queryKeys } from "../query-keys";
import type { Conversation, ConversationWithMessages, UpdateConversationRequest, Suggestion } from "../types/agent";
import type { ListResponse } from "../types/common";

export function useConversations() {
  return useQuery({
    queryKey: queryKeys.conversations.list(),
    queryFn: () => apiFetch<ListResponse<Conversation>>("/agent/conversations"),
  });
}

export function useConversation(id: string) {
  return useQuery({
    queryKey: queryKeys.conversations.detail(id),
    queryFn: () => apiFetch<ConversationWithMessages>(`/agent/conversations/${id}`),
    enabled: !!id,
  });
}

export function useUpdateConversation(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateConversationRequest) =>
      apiFetch(`/agent/conversations/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.conversations.all });
    },
  });
}

export function useDeleteConversation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiFetch(`/agent/conversations/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.conversations.all }),
  });
}

export function useSuggestions() {
  return useQuery({
    queryKey: queryKeys.suggestions.all,
    queryFn: () => apiFetch<Suggestion[]>("/agent/suggestions"),
  });
}
