// Dialog for adding a prerequisite relationship between two subjects.
import { useState } from "react";
import { FormDialog, Label } from "@mcs-erp/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@mcs-erp/ui";
import { useSubjects, useAddPrerequisite } from "@mcs-erp/api-client";

interface AddPrerequisiteDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function AddPrerequisiteDialog({ open, onOpenChange, onSuccess }: AddPrerequisiteDialogProps) {
  const [subjectId, setSubjectId] = useState("");
  const [prerequisiteId, setPrerequisiteId] = useState("");

  const { data: subjectsData } = useSubjects({ limit: 1000 });
  const addPrerequisite = useAddPrerequisite(subjectId);

  async function handleSubmit() {
    if (!subjectId || !prerequisiteId || subjectId === prerequisiteId) return;
    await addPrerequisite.mutateAsync(prerequisiteId);
    setSubjectId("");
    setPrerequisiteId("");
    onSuccess?.();
    onOpenChange(false);
  }

  const subjects = subjectsData?.items ?? [];

  return (
    <FormDialog
      open={open}
      onOpenChange={(v) => {
        if (!v) { setSubjectId(""); setPrerequisiteId(""); }
        onOpenChange(v);
      }}
      title="Add Prerequisite"
      description="Select a subject and the prerequisite it requires."
      onSubmit={handleSubmit}
      submitLabel="Add"
      isSubmitting={addPrerequisite.isPending}
    >
      <div className="space-y-3">
        <div className="space-y-1">
          <Label>Subject</Label>
          <Select value={subjectId} onValueChange={setSubjectId}>
            <SelectTrigger>
              <SelectValue placeholder="Select subject..." />
            </SelectTrigger>
            <SelectContent>
              {subjects.map((s) => (
                <SelectItem key={s.id} value={s.id}>
                  {s.code} — {s.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-1">
          <Label>Prerequisite</Label>
          <Select value={prerequisiteId} onValueChange={setPrerequisiteId}>
            <SelectTrigger>
              <SelectValue placeholder="Select prerequisite..." />
            </SelectTrigger>
            <SelectContent>
              {subjects
                .filter((s) => s.id !== subjectId)
                .map((s) => (
                  <SelectItem key={s.id} value={s.id}>
                    {s.code} — {s.name}
                  </SelectItem>
                ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </FormDialog>
  );
}
