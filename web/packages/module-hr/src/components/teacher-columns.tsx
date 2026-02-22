// TanStack Table column definitions for the Teacher entity.
import type { ColumnDef } from "@tanstack/react-table";
import { Badge } from "@mcs-erp/ui";
import { Button } from "@mcs-erp/ui";
import type { Teacher } from "@mcs-erp/api-client";

interface TeacherColumnsOptions {
  departments: Record<string, string>; // id → name map
  onEdit: (teacher: Teacher) => void;
  onView: (teacherId: string) => void;
}

export function createTeacherColumns({
  departments,
  onEdit,
  onView,
}: TeacherColumnsOptions): ColumnDef<Teacher>[] {
  return [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.name}</span>
      ),
    },
    {
      accessorKey: "email",
      header: "Email",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.email}</span>
      ),
    },
    {
      accessorKey: "department_id",
      header: "Department",
      cell: ({ row }) => {
        const deptId = row.original.department_id;
        return (
          <span>{deptId ? (departments[deptId] ?? deptId) : "—"}</span>
        );
      },
    },
    {
      accessorKey: "qualifications",
      header: "Qualifications",
      cell: ({ row }) => {
        const quals = row.original.qualifications;
        if (!quals?.length) return <span className="text-muted-foreground">—</span>;
        return (
          <div className="flex flex-wrap gap-1">
            {quals.map((q) => (
              <Badge key={q} variant="secondary" className="text-xs">
                {q}
              </Badge>
            ))}
          </div>
        );
      },
    },
    {
      accessorKey: "is_active",
      header: "Status",
      cell: ({ row }) =>
        row.original.is_active ? (
          <Badge variant="default">Active</Badge>
        ) : (
          <Badge variant="outline">Inactive</Badge>
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
            onClick={() => onView(row.original.id)}
          >
            View
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onEdit(row.original)}
          >
            Edit
          </Button>
        </div>
      ),
    },
  ];
}
