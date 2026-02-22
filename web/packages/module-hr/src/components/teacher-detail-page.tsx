// Teacher detail page: info card + editable availability grid.
import { useState, useCallback } from "react";
import { useParams, useNavigate } from "@tanstack/react-router";
import {
  Button, Card, CardContent, CardHeader, CardTitle,
  Badge, AvailabilityGrid, LoadingSpinner, EmptyState,
} from "@mcs-erp/ui";
import { useTeacher, useTeacherAvailability, useUpdateTeacherAvailability, useDepartments } from "@mcs-erp/api-client";
import type { AvailabilitySlot } from "@mcs-erp/api-client";
import { TeacherFormDialog } from "./teacher-form-dialog";

export function TeacherDetailPage() {
  const { teacherId } = useParams({ strict: false }) as { teacherId: string };
  const navigate = useNavigate();

  const { data: teacher, isLoading: teacherLoading } = useTeacher(teacherId);
  const { data: availabilityData, isLoading: availLoading } = useTeacherAvailability(teacherId);
  const { data: deptData } = useDepartments();
  const updateAvailability = useUpdateTeacherAvailability(teacherId);

  const [slots, setSlots] = useState<AvailabilitySlot[] | null>(null);
  const [editOpen, setEditOpen] = useState(false);

  // Use local slot state if modified, otherwise fall back to fetched data.
  const currentSlots = slots ?? availabilityData ?? [];

  const handleSave = useCallback(async () => {
    await updateAvailability.mutateAsync(currentSlots);
    setSlots(null);
  }, [updateAvailability, currentSlots]);

  const handleSlotsChange = useCallback((updated: AvailabilitySlot[]) => {
    setSlots(updated);
  }, []);

  if (teacherLoading) {
    return (
      <div className="flex justify-center items-center h-48">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (!teacher) {
    return (
      <EmptyState
        title="Teacher not found"
        description="This teacher does not exist or has been removed."
        action={
          <Button variant="outline" onClick={() => navigate({ to: "/teachers" })}>
            Back to list
          </Button>
        }
      />
    );
  }

  const departmentName =
    teacher.department_id
      ? (deptData?.items.find((d) => d.id === teacher.department_id)?.name ?? teacher.department_id)
      : "—";

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="outline" size="sm" onClick={() => navigate({ to: "/teachers" })}>
            Back
          </Button>
          <h1 className="text-2xl font-bold">{teacher.name}</h1>
          {teacher.is_active ? (
            <Badge variant="default">Active</Badge>
          ) : (
            <Badge variant="outline">Inactive</Badge>
          )}
        </div>
        <Button variant="outline" onClick={() => setEditOpen(true)}>
          Edit
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Teacher Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Email</p>
              <p className="font-medium">{teacher.email}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Department</p>
              <p className="font-medium">{departmentName}</p>
            </div>
          </div>
          {teacher.qualifications?.length > 0 && (
            <div>
              <p className="text-muted-foreground text-sm mb-1">Qualifications</p>
              <div className="flex flex-wrap gap-1">
                {teacher.qualifications.map((q) => (
                  <Badge key={q} variant="secondary">{q}</Badge>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Availability Schedule</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {availLoading ? (
            <div className="flex justify-center py-6">
              <LoadingSpinner />
            </div>
          ) : (
            <AvailabilityGrid slots={currentSlots} onChange={handleSlotsChange} />
          )}
          <div className="flex justify-end">
            <Button
              onClick={handleSave}
              disabled={updateAvailability.isPending || slots === null}
            >
              {updateAvailability.isPending ? "Saving..." : "Save Availability"}
            </Button>
          </div>
        </CardContent>
      </Card>

      <TeacherFormDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        teacher={teacher}
      />
    </div>
  );
}
