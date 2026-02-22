import { useRef, useCallback, type KeyboardEvent } from "react";
import { Button, cn } from "@mcs-erp/ui";

interface ChatInputProps {
  onSend: (message: string) => void;
  disabled?: boolean;
}

// Textarea chat input with auto-resize, Enter-to-send, and Shift+Enter for newline.
export function ChatInput({ onSend, disabled = false }: ChatInputProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const resize = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = "auto";
    // Clamp at 6 lines (~144px with line-height 24px)
    el.style.height = `${Math.min(el.scrollHeight, 144)}px`;
  }, []);

  const handleInput = useCallback(() => {
    resize();
  }, [resize]);

  const submit = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    const value = el.value.trim();
    if (!value || disabled) return;
    onSend(value);
    el.value = "";
    el.style.height = "auto";
  }, [onSend, disabled]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        submit();
      }
    },
    [submit],
  );

  return (
    <div className="flex items-end gap-2 p-3 border-t bg-background">
      <textarea
        ref={textareaRef}
        rows={1}
        placeholder="Type a message… (Enter to send, Shift+Enter for newline)"
        disabled={disabled}
        onInput={handleInput}
        onKeyDown={handleKeyDown}
        className={cn(
          "flex-1 resize-none rounded-lg border border-input bg-transparent px-3 py-2 text-sm",
          "placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2",
          "focus-visible:ring-ring disabled:opacity-50 overflow-y-auto",
          "min-h-[40px] leading-6",
        )}
      />
      <Button
        size="sm"
        onClick={submit}
        disabled={disabled}
        className="shrink-0 h-10"
      >
        Send
      </Button>
    </div>
  );
}
