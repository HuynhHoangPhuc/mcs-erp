import { Card, CardContent, CardHeader, CardTitle, Badge } from "@mcs-erp/ui";
import type { Assignment } from "@mcs-erp/api-client";

interface Conflict {
  type: "teacher" | "room";
  label: string;
  assignments: Assignment[];
}

interface Props {
  assignments: Assignment[];
  subjectMap: Map<string, { code: string; name: string }>;
  teacherMap: Map<string, { name: string }>;
  roomMap: Map<string, { name: string }>;
}

const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

function detectConflicts(
  assignments: Assignment[],
  teacherMap: Map<string, { name: string }>,
  roomMap: Map<string, { name: string }>,
): Conflict[] {
  const conflicts: Conflict[] = [];
  const slotKey = (a: Assignment) => `${a.day}-${a.period}`;

  // Group by teacher+slot
  const teacherSlots = new Map<string, Assignment[]>();
  for (const a of assignments) {
    const key = `${a.teacher_id}:${slotKey(a)}`;
    const arr = teacherSlots.get(key) ?? [];
    arr.push(a);
    teacherSlots.set(key, arr);
  }
  for (const [key, arr] of teacherSlots) {
    if (arr.length > 1) {
      const tid = key.split(":")[0];
      conflicts.push({
        type: "teacher",
        label: `Teacher ${teacherMap.get(tid)?.name ?? tid} double-booked on ${DAYS[arr[0].day]} P${arr[0].period}`,
        assignments: arr,
      });
    }
  }

  // Group by room+slot
  const roomSlots = new Map<string, Assignment[]>();
  for (const a of assignments) {
    const key = `${a.room_id}:${slotKey(a)}`;
    const arr = roomSlots.get(key) ?? [];
    arr.push(a);
    roomSlots.set(key, arr);
  }
  for (const [key, arr] of roomSlots) {
    if (arr.length > 1) {
      const rid = key.split(":")[0];
      conflicts.push({
        type: "room",
        label: `Room ${roomMap.get(rid)?.name ?? rid} double-booked on ${DAYS[arr[0].day]} P${arr[0].period}`,
        assignments: arr,
      });
    }
  }

  return conflicts;
}

export function ConflictPanel({ assignments, subjectMap, teacherMap, roomMap }: Props) {
  const conflicts = detectConflicts(assignments, teacherMap, roomMap);

  if (conflicts.length === 0) {
    return (
      <Card>
        <CardContent className="py-6 text-center text-sm text-muted-foreground">
          No conflicts detected.
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-lg flex items-center gap-2">
          Conflicts
          <Badge variant="destructive">{conflicts.length}</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ul className="space-y-2">
          {conflicts.map((c, i) => (
            <li key={i} className="text-sm border-l-2 border-destructive pl-3">
              <div className="font-medium">{c.label}</div>
              <div className="text-muted-foreground">
                {c.assignments.map((a) => subjectMap.get(a.subject_id)?.code ?? a.subject_id).join(", ")}
              </div>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}

export { detectConflicts, type Conflict };
