// Custom ReactFlow node component displaying subject info in a card-style box.
import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";

export interface SubjectNodeData {
  label: string;
  code: string;
  credits: number;
  highlighted?: boolean;
}

export const SubjectNode = memo(function SubjectNode({ data }: NodeProps) {
  const nodeData = data as unknown as SubjectNodeData;
  return (
    <div
      className={[
        "rounded-lg border bg-card px-4 py-3 shadow-sm min-w-[140px] text-center transition-colors",
        nodeData.highlighted
          ? "border-primary bg-primary/10 shadow-md"
          : "border-border hover:border-primary/50",
      ].join(" ")}
    >
      <Handle type="target" position={Position.Top} className="!bg-muted-foreground" />
      <p className="font-mono text-xs text-muted-foreground">{nodeData.code}</p>
      <p className="mt-0.5 text-sm font-semibold leading-tight">{nodeData.label}</p>
      <p className="mt-1 text-xs text-muted-foreground">{nodeData.credits} cr</p>
      <Handle type="source" position={Position.Bottom} className="!bg-muted-foreground" />
    </div>
  );
});
