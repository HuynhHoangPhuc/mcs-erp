import { Badge } from "@mcs-erp/ui";

interface Props {
  subjectCode: string;
  teacherName: string;
  roomName: string;
  hasConflict: boolean;
  onClick: () => void;
}

export function TimetableCell({ subjectCode, teacherName, roomName, hasConflict, onClick }: Props) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`p-1.5 rounded text-left text-xs border transition-colors hover:bg-accent ${
        hasConflict ? "border-destructive bg-destructive/10" : "border-border"
      }`}
    >
      <div className="font-mono font-semibold truncate">{subjectCode}</div>
      <div className="text-muted-foreground truncate">{teacherName}</div>
      <div className="text-muted-foreground truncate">{roomName}</div>
      {hasConflict && <Badge variant="destructive" className="mt-0.5 text-[10px] px-1 py-0">Conflict</Badge>}
    </button>
  );
}
