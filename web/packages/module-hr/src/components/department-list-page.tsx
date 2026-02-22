// Department list page: DataTable with inline create/edit/delete actions.
import { useState, useMemo, useCallback } from "react";
import type { ColumnDef } from "@tanstack/react-table";
import { Button, DataTable, Badge, ConfirmDialog } from "@mcs-erp/ui";
import { useDepartments, useDeleteDepartment } from "@mcs-erp/api-client";
import type { Department } from "@mcs-erp/api-client";
import { DepartmentFormDialog } from "./department-form-dialog";

export function DepartmentListPage() {
  const { data, isLoading } = useDepartments();
  const deleteDepartment = useDeleteDepartment();

  const [formOpen, setFormOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Department | undefined>(undefined);
  const [deleteTarget, setDeleteTarget] = useState<Department | undefined>(undefined);

  const handleAdd = useCallback(() => {
    setEditTarget(undefined);
    setFormOpen(true);
  }, []);

  const handleEdit = useCallback((dept: Department) => {
    setEditTarget(dept);
    setFormOpen(true);
  }, []);

  const handleDelete = useCallback((dept: Department) => {
    setDeleteTarget(dept);
  }, []);

  const handleConfirmDelete = useCallback(async () => {
    if (!deleteTarget) return;
    await deleteDepartment.mutateAsync(deleteTarget.id);
    setDeleteTarget(undefined);
  }, [deleteTarget, deleteDepartment]);

  const columns = useMemo<ColumnDef<Department>[]>(
    () => [
      {
        accessorKey: "name",
        header: "Name",
        cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
      },
      {
        accessorKey: "description",
        header: "Description",
        cell: ({ row }) => (
          <span className="text-muted-foreground">
            {row.original.description || "—"}
          </span>
        ),
      },
      {
        accessorKey: "head_teacher_id",
        header: "Head Teacher",
        cell: ({ row }) =>
          row.original.head_teacher_id ? (
            <Badge variant="secondary">Assigned</Badge>
          ) : (
            <span className="text-muted-foreground text-sm">None</span>
          ),
      },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => (
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleEdit(row.original)}
            >
              Edit
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="text-destructive hover:text-destructive"
              onClick={() => handleDelete(row.original)}
            >
              Delete
            </Button>
          </div>
        ),
      },
    ],
    [handleEdit, handleDelete]
  );

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Departments</h1>
        <Button onClick={handleAdd}>Add Department</Button>
      </div>

      <DataTable
        columns={columns}
        data={data?.items ?? []}
        total={data?.total ?? 0}
        isLoading={isLoading}
      />

      <DepartmentFormDialog
        open={formOpen}
        onOpenChange={(open) => {
          setFormOpen(open);
          if (!open) setEditTarget(undefined);
        }}
        department={editTarget}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => { if (!open) setDeleteTarget(undefined); }}
        title="Delete Department"
        description={`Are you sure you want to delete "${deleteTarget?.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={handleConfirmDelete}
        isConfirming={deleteDepartment.isPending}
        variant="destructive"
      />
    </div>
  );
}
