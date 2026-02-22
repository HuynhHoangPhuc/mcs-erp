// Category list page with CRUD operations using DataTable.
import { useState, useMemo } from "react";
import type { ColumnDef } from "@tanstack/react-table";
import { Button, DataTable, ConfirmDialog, LoadingSpinner } from "@mcs-erp/ui";
import { useCategories, useDeleteCategory, type Category } from "@mcs-erp/api-client";
import { CategoryFormDialog } from "./category-form-dialog";
import { Pencil, Trash2 } from "lucide-react";

export function CategoryListPage() {
  const { data: categoriesData, isLoading } = useCategories();
  const deleteCategory = useDeleteCategory();

  const [formOpen, setFormOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Category | undefined>();
  const [deleteTarget, setDeleteTarget] = useState<Category | undefined>();

  const columns: ColumnDef<Category>[] = useMemo(
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
          <span className="text-muted-foreground text-sm line-clamp-1">
            {row.original.description || "—"}
          </span>
        ),
      },
      {
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => (
          <span className="text-sm">
            {new Date(row.original.created_at).toLocaleDateString()}
          </span>
        ),
      },
      {
        id: "actions",
        header: "",
        cell: ({ row }) => (
          <div className="flex items-center gap-1 justify-end">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => { setEditTarget(row.original); setFormOpen(true); }}
            >
              <Pencil className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setDeleteTarget(row.original)}
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        ),
      },
    ],
    []
  );

  function handleAddNew() {
    setEditTarget(undefined);
    setFormOpen(true);
  }

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Categories</h1>
          <p className="text-muted-foreground text-sm">Manage subject categories.</p>
        </div>
        <Button onClick={handleAddNew}>Add Category</Button>
      </div>

      {isLoading ? (
        <LoadingSpinner />
      ) : (
        <DataTable
          columns={columns}
          data={categoriesData?.items ?? []}
          total={categoriesData?.total ?? 0}
        />
      )}

      <CategoryFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        initialData={editTarget}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(undefined)}
        title="Delete Category"
        description={`Are you sure you want to delete "${deleteTarget?.name}"? This cannot be undone.`}
        confirmLabel="Delete"
        isConfirming={deleteCategory.isPending}
        onConfirm={async () => {
          if (deleteTarget) {
            await deleteCategory.mutateAsync(deleteTarget.id);
            setDeleteTarget(undefined);
          }
        }}
      />
    </div>
  );
}
