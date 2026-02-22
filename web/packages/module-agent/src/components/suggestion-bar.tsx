import { Badge } from "@mcs-erp/ui";
import { useSuggestions } from "@mcs-erp/api-client";

interface SuggestionBarProps {
  onSend: (message: string) => void;
  disabled?: boolean;
}

// Horizontal scrollable row of suggestion chips above chat input.
export function SuggestionBar({ onSend, disabled = false }: SuggestionBarProps) {
  const { data: suggestions } = useSuggestions();

  if (!suggestions || suggestions.length === 0) return null;

  return (
    <div className="flex gap-2 px-3 py-2 overflow-x-auto scrollbar-none border-t">
      {suggestions.map((suggestion, idx) => (
        <button
          key={idx}
          onClick={() => !disabled && onSend(suggestion.text)}
          disabled={disabled}
          className="shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-full"
        >
          <Badge
            variant="outline"
            className="cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors whitespace-nowrap disabled:opacity-50"
          >
            {suggestion.text}
          </Badge>
        </button>
      ))}
    </div>
  );
}
