import { useState } from "react";
import { Button, Checkbox, Card, CardContent, CardHeader, CardTitle, LoadingSpinner } from "@mcs-erp/ui";
import {
  useSubjects, useSemesterSubjects, useSetSemesterSubjects,
} from "@mcs-erp/api-client";

interface Props {
  semesterId: string;
}

// Step 1: Select subjects for semester
export function SemesterSetupStep({ semesterId }: Props) {
  const { data: allSubjects, isLoading: loadingSubjects } = useSubjects({ limit: 200 });
  const { data: semesterSubjects } = useSemesterSubjects(semesterId);
  const setSubjectsMutation = useSetSemesterSubjects(semesterId);

  const existingIds = new Set(semesterSubjects?.map((s) => s.subject_id) ?? []);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(existingIds);

  // Sync when data loads
  if (semesterSubjects && selectedIds.size === 0 && existingIds.size > 0) {
    setSelectedIds(new Set(existingIds));
  }

  const toggle = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  };

  const handleSave = async () => {
    await setSubjectsMutation.mutateAsync({ subject_ids: Array.from(selectedIds) });
  };

  if (loadingSubjects) {
    return <div className="flex justify-center py-8"><LoadingSpinner /></div>;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Select Subjects</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid gap-2 max-h-96 overflow-y-auto">
          {allSubjects?.items.map((subject) => (
            <label
              key={subject.id}
              className="flex items-center gap-3 p-2 rounded-md hover:bg-muted cursor-pointer"
            >
              <Checkbox
                checked={selectedIds.has(subject.id)}
                onCheckedChange={() => toggle(subject.id)}
              />
              <span className="font-mono text-sm text-muted-foreground">{subject.code}</span>
              <span>{subject.name}</span>
              <span className="ml-auto text-sm text-muted-foreground">{subject.credits} credits</span>
            </label>
          ))}
        </div>
        <div className="mt-4 flex justify-between items-center">
          <p className="text-sm text-muted-foreground">{selectedIds.size} subjects selected</p>
          <Button onClick={handleSave} disabled={setSubjectsMutation.isPending}>
            {setSubjectsMutation.isPending ? "Saving..." : "Save Selection"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
