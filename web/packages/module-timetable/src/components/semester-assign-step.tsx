import { Card, CardContent, CardHeader, CardTitle, Button, LoadingSpinner } from "@mcs-erp/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@mcs-erp/ui";
import {
  useSemesterSubjects, useSubjects, useTeachers, useAssignTeacher,
} from "@mcs-erp/api-client";

interface Props {
  semesterId: string;
}

// Step 2: Assign teachers to semester subjects
export function SemesterAssignStep({ semesterId }: Props) {
  const { data: semSubjects, isLoading } = useSemesterSubjects(semesterId);
  const { data: allSubjects } = useSubjects({ limit: 200 });
  const { data: allTeachers } = useTeachers({ limit: 200 });

  // Build lookup maps
  const subjectMap = new Map(allSubjects?.items.map((s) => [s.id, s]) ?? []);
  const teacherMap = new Map(allTeachers?.items.map((t) => [t.id, t]) ?? []);

  if (isLoading) {
    return <div className="flex justify-center py-8"><LoadingSpinner /></div>;
  }

  if (!semSubjects?.length) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          No subjects added yet. Go to Setup tab first.
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Assign Teachers</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {semSubjects.map((ss) => (
            <TeacherAssignmentRow
              key={ss.subject_id}
              semesterId={semesterId}
              subjectId={ss.subject_id}
              subjectName={subjectMap.get(ss.subject_id)?.name ?? ss.subject_id}
              subjectCode={subjectMap.get(ss.subject_id)?.code ?? ""}
              currentTeacherId={ss.teacher_id}
              teachers={allTeachers?.items ?? []}
              teacherMap={teacherMap}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function TeacherAssignmentRow({
  semesterId, subjectId, subjectName, subjectCode,
  currentTeacherId, teachers, teacherMap,
}: {
  semesterId: string;
  subjectId: string;
  subjectName: string;
  subjectCode: string;
  currentTeacherId: string | null;
  teachers: { id: string; name: string }[];
  teacherMap: Map<string, { id: string; name: string }>;
}) {
  const assignMutation = useAssignTeacher(semesterId, subjectId);

  const handleAssign = async (teacherId: string) => {
    await assignMutation.mutateAsync({ teacher_id: teacherId });
  };

  return (
    <div className="flex items-center gap-4 p-3 rounded-md border">
      <div className="flex-1">
        <span className="font-mono text-sm text-muted-foreground mr-2">{subjectCode}</span>
        <span className="font-medium">{subjectName}</span>
      </div>
      <Select
        value={currentTeacherId ?? ""}
        onValueChange={handleAssign}
      >
        <SelectTrigger className="w-60">
          <SelectValue placeholder="Select teacher...">
            {currentTeacherId ? teacherMap.get(currentTeacherId)?.name ?? "Unknown" : "Select teacher..."}
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          {teachers.map((t) => (
            <SelectItem key={t.id} value={t.id}>{t.name}</SelectItem>
          ))}
        </SelectContent>
      </Select>
      {assignMutation.isPending && <LoadingSpinner size="sm" />}
    </div>
  );
}
