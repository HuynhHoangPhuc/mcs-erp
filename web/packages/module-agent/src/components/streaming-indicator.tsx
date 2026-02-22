import { cn } from "@mcs-erp/ui";

interface StreamingIndicatorProps {
  className?: string;
}

// Animated "thinking" indicator shown while AI is generating a response.
export function StreamingIndicator({ className }: StreamingIndicatorProps) {
  return (
    <div className={cn("flex items-center gap-1 px-1 py-0.5", className)}>
      <span className="text-xs text-muted-foreground mr-1">Thinking</span>
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          className="inline-block w-1.5 h-1.5 rounded-full bg-muted-foreground animate-bounce"
          style={{ animationDelay: `${i * 0.15}s` }}
        />
      ))}
    </div>
  );
}
