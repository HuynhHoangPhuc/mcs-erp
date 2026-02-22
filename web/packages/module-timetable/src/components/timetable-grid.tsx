import { Fragment, useState } from "react";
import type { Assignment } from "@mcs-erp/api-client";
import { TimetableCell } from "./timetable-cell";
import { AssignmentEditDialog } from "./assignment-edit-dialog";
import { type Conflict } from "./conflict-panel";

const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const PERIODS = Array.from({ length: 10 }, (_, i) => i + 1);

interface Props {
  assignments: Assignment[];
  conflicts: Conflict[];
  subjectMap: Map<string, { id: string; code: string; name: string }>;
  teacherMap: Map<string, { id: string; name: string }>;
  roomMap: Map<string, { id: string; name: string }>;
  readOnly?: boolean;
}

export function TimetableGrid({ assignments, conflicts, subjectMap, teacherMap, roomMap, readOnly }: Props) {
  const [editingAssignment, setEditingAssignment] = useState<Assignment | null>(null);

  // Build lookup: "day-period" → Assignment[]
  const grid = new Map<string, Assignment[]>();
  for (const a of assignments) {
    const key = `${a.day}-${a.period}`;
    const arr = grid.get(key) ?? [];
    arr.push(a);
    grid.set(key, arr);
  }

  // Build conflict set for quick lookup
  const conflictAssignmentIds = new Set<string>();
  for (const c of conflicts) {
    for (const a of c.assignments) {
      conflictAssignmentIds.add(a.id);
    }
  }

  return (
    <>
      <div
        className="grid gap-px bg-border rounded-md overflow-hidden"
        style={{
          gridTemplateColumns: `4rem repeat(${DAYS.length}, 1fr)`,
          gridTemplateRows: `2.5rem repeat(${PERIODS.length}, 1fr)`,
        }}
      >
        {/* Header: empty corner */}
        <div className="bg-muted flex items-center justify-center text-xs font-medium text-muted-foreground" />
        {/* Header: day labels */}
        {DAYS.map((d) => (
          <div key={d} className="bg-muted flex items-center justify-center text-xs font-medium">
            {d}
          </div>
        ))}

        {/* Rows */}
        {PERIODS.map((p) => (
          <Fragment key={`row-${p}`}>
            {/* Period label */}
            <div className="bg-muted flex items-center justify-center text-xs text-muted-foreground">
              P{p}
            </div>
            {/* Day cells */}
            {DAYS.map((_, dayIdx) => {
              const key = `${dayIdx}-${p}`;
              const cellAssignments = grid.get(key) ?? [];
              return (
                <div key={key} className="bg-background p-0.5 min-h-[3.5rem]">
                  {cellAssignments.map((a) => (
                    <TimetableCell
                      key={a.id}
                      subjectCode={subjectMap.get(a.subject_id)?.code ?? "?"}
                      teacherName={teacherMap.get(a.teacher_id)?.name ?? "?"}
                      roomName={roomMap.get(a.room_id)?.name ?? "?"}
                      hasConflict={conflictAssignmentIds.has(a.id)}
                      onClick={() => !readOnly && setEditingAssignment(a)}
                    />
                  ))}
                </div>
              );
            })}
          </Fragment>
        ))}
      </div>

      {editingAssignment && (
        <AssignmentEditDialog
          assignment={editingAssignment}
          teachers={Array.from(teacherMap.values())}
          rooms={Array.from(roomMap.values())}
          subjectName={subjectMap.get(editingAssignment.subject_id)?.name ?? "Unknown"}
          open={!!editingAssignment}
          onOpenChange={(open) => !open && setEditingAssignment(null)}
        />
      )}
    </>
  );
}
