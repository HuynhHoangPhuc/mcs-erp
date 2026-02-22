// Teacher list page: filterable, paginated table with create dialog.
import { useState, useMemo, useCallback } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Button, DataTable } from "@mcs-erp/ui";
import { useTeachers, useDepartments } from "@mcs-erp/api-client";
import type { Teacher, TeacherFilter } from "@mcs-erp/api-client";
import { TeacherFilters } from "./teacher-filters";
import { TeacherFormDialog } from "./teacher-form-dialog";
import { createTeacherColumns } from "./teacher-columns";

const PAGE_LIMIT = 20;

export function TeacherListPage() {
  const navigate = useNavigate();
  const [filters, setFilters] = useState<TeacherFilter & { search?: string }>({});
  const [offset, setOffset] = useState(0);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Teacher | undefined>(undefined);

  const { data, isLoading } = useTeachers({ ...filters, offset, limit: PAGE_LIMIT });
  const { data: deptData } = useDepartments();

  const departmentMap = useMemo(() => {
    const map: Record<string, string> = {};
    for (const dept of deptData?.items ?? []) {
      map[dept.id] = dept.name;
    }
    return map;
  }, [deptData]);

  const handleFiltersChange = useCallback(
    (next: TeacherFilter & { search?: string }) => {
      setFilters(next);
      setOffset(0);
    },
    []
  );

  const handleEdit = useCallback((teacher: Teacher) => {
    setEditTarget(teacher);
    setDialogOpen(true);
  }, []);

  const handleView = useCallback(
    (teacherId: string) => {
      navigate({ to: "/teachers/$teacherId", params: { teacherId } });
    },
    [navigate]
  );

  const handleAddTeacher = useCallback(() => {
    setEditTarget(undefined);
    setDialogOpen(true);
  }, []);

  const columns = useMemo(
    () => createTeacherColumns({ departments: departmentMap, onEdit: handleEdit, onView: handleView }),
    [departmentMap, handleEdit, handleView]
  );

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Teachers</h1>
        <Button onClick={handleAddTeacher}>Add Teacher</Button>
      </div>

      <TeacherFilters filters={filters} onChange={handleFiltersChange} />

      <DataTable
        columns={columns}
        data={data?.items ?? []}
        total={data?.total ?? 0}
        offset={offset}
        limit={PAGE_LIMIT}
        onPaginationChange={setOffset}
        isLoading={isLoading}
      />

      <TeacherFormDialog
        open={dialogOpen}
        onOpenChange={(open) => {
          setDialogOpen(open);
          if (!open) setEditTarget(undefined);
        }}
        teacher={editTarget}
      />
    </div>
  );
}
