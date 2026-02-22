import { useState } from "react";
import { Button, ScrollArea, ConfirmDialog, cn } from "@mcs-erp/ui";
import { useConversations, useDeleteConversation } from "@mcs-erp/api-client";
import type { Conversation } from "@mcs-erp/api-client";

interface ConversationSidebarProps {
  selectedId: string | null;
  onSelect: (id: string) => void;
  onNew: () => void;
}

function formatDate(isoString: string): string {
  const date = new Date(isoString);
  const now = new Date();
  const diffDays = Math.floor((now.getTime() - date.getTime()) / 86400000);
  if (diffDays === 0) return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  if (diffDays === 1) return "Yesterday";
  if (diffDays < 7) return date.toLocaleDateString([], { weekday: "short" });
  return date.toLocaleDateString([], { month: "short", day: "numeric" });
}

interface ConversationItemProps {
  conversation: Conversation;
  isSelected: boolean;
  onSelect: () => void;
  onDelete: () => void;
}

function ConversationItem({ conversation, isSelected, onSelect, onDelete }: ConversationItemProps) {
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <>
      <div
        className={cn(
          "group flex items-center gap-2 rounded-md px-3 py-2 cursor-pointer hover:bg-accent transition-colors",
          isSelected && "bg-accent",
        )}
        onClick={onSelect}
      >
        <div className="flex-1 min-w-0">
          <p className="truncate text-sm font-medium">{conversation.title || "New conversation"}</p>
          <p className="text-xs text-muted-foreground">{formatDate(conversation.updated_at)}</p>
        </div>
        <button
          onClick={(e) => { e.stopPropagation(); setConfirmOpen(true); }}
          className="shrink-0 opacity-0 group-hover:opacity-100 p-1 rounded hover:bg-destructive/10 hover:text-destructive transition-all"
          aria-label="Delete conversation"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6" />
          </svg>
        </button>
      </div>

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Delete conversation"
        description="This will permanently delete this conversation and all its messages."
        confirmLabel="Delete"
        onConfirm={onDelete}
      />
    </>
  );
}

// Left sidebar listing all conversations with new-chat button and delete support.
export function ConversationSidebar({ selectedId, onSelect, onNew }: ConversationSidebarProps) {
  const { data: conversationsData } = useConversations();
  const { mutate: deleteConversation } = useDeleteConversation();

  const conversations = conversationsData?.items ?? [];

  return (
    <div className="flex flex-col h-full w-64 border-r bg-background">
      <div className="p-3 border-b">
        <Button onClick={onNew} className="w-full" variant="outline" size="sm">
          + New Chat
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-2 flex flex-col gap-1">
          {conversations.length === 0 && (
            <p className="text-xs text-muted-foreground text-center py-4">No conversations yet</p>
          )}
          {conversations.map((conv) => (
            <ConversationItem
              key={conv.id}
              conversation={conv}
              isSelected={selectedId === conv.id}
              onSelect={() => onSelect(conv.id)}
              onDelete={() => deleteConversation(conv.id)}
            />
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}
