import { useState } from "react";
import { Button, Card, CardContent, CardHeader, CardTitle, Badge, LoadingSpinner } from "@mcs-erp/ui";
import {
  useSchedule, useGenerateSchedule, useApproveSemester,
  useSubjects, useTeachers, useRooms,
  type Semester,
} from "@mcs-erp/api-client";
import { TimetableGrid } from "./timetable-grid";
import { ConflictPanel, detectConflicts } from "./conflict-panel";

interface Props {
  semesterId: string;
  semester: Semester;
}

export function SemesterReviewStep({ semesterId, semester }: Props) {
  const { data: schedule, isLoading } = useSchedule(semesterId);
  const generateMutation = useGenerateSchedule(semesterId);
  const approveMutation = useApproveSemester(semesterId);

  const { data: allSubjects } = useSubjects({ limit: 200 });
  const { data: allTeachers } = useTeachers({ limit: 200 });
  const { data: allRooms } = useRooms({ limit: 200 });

  const [generating, setGenerating] = useState(false);

  const subjectMap = new Map(
    allSubjects?.items.map((s) => [s.id, { id: s.id, code: s.code, name: s.name }]) ?? [],
  );
  const teacherMap = new Map(
    allTeachers?.items.map((t) => [t.id, { id: t.id, name: t.name }]) ?? [],
  );
  const roomMap = new Map(
    allRooms?.items.map((r) => [r.id, { id: r.id, name: r.name }]) ?? [],
  );

  const assignments = schedule?.assignments ?? [];
  const conflicts = detectConflicts(assignments, teacherMap, roomMap);
  const isReadOnly = semester.status === "approved";

  const handleGenerate = async () => {
    setGenerating(true);
    try {
      await generateMutation.mutateAsync();
    } finally {
      setGenerating(false);
    }
  };

  const handleApprove = async () => {
    await approveMutation.mutateAsync();
  };

  if (isLoading) {
    return <div className="flex justify-center py-8"><LoadingSpinner /></div>;
  }

  // No schedule yet
  if (!schedule || assignments.length === 0) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground mb-4">No schedule generated yet.</p>
          <Button onClick={handleGenerate} disabled={generating}>
            {generating ? "Generating..." : "Generate Schedule"}
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {/* Stats bar */}
      <div className="flex items-center gap-4 flex-wrap">
        <Badge variant="outline">Version {schedule.version}</Badge>
        <Badge variant={schedule.hard_violations > 0 ? "destructive" : "default"}>
          {schedule.hard_violations} hard violations
        </Badge>
        <Badge variant="outline">Soft penalty: {schedule.soft_penalty}</Badge>
        <span className="text-xs text-muted-foreground">
          Generated {new Date(schedule.generated_at).toLocaleString()}
        </span>
        <div className="ml-auto flex gap-2">
          {!isReadOnly && (
            <>
              <Button variant="outline" onClick={handleGenerate} disabled={generating}>
                {generating ? "Regenerating..." : "Regenerate"}
              </Button>
              <Button
                onClick={handleApprove}
                disabled={approveMutation.isPending || conflicts.length > 0}
              >
                {approveMutation.isPending ? "Approving..." : "Approve Schedule"}
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Conflict panel */}
      {conflicts.length > 0 && (
        <ConflictPanel
          assignments={assignments}
          subjectMap={subjectMap}
          teacherMap={teacherMap}
          roomMap={roomMap}
        />
      )}

      {/* Grid */}
      <TimetableGrid
        assignments={assignments}
        conflicts={conflicts}
        subjectMap={subjectMap}
        teacherMap={teacherMap}
        roomMap={roomMap}
        readOnly={isReadOnly}
      />
    </div>
  );
}
