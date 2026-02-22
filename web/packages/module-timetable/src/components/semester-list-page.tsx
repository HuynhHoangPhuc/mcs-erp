import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { DataTable, Button, Badge } from "@mcs-erp/ui";
import { useSemesters, type Semester, type SemesterStatus } from "@mcs-erp/api-client";
import { Plus, Eye } from "lucide-react";
import { SemesterFormDialog } from "./semester-form-dialog";

const statusColors: Record<SemesterStatus, "default" | "secondary" | "destructive" | "outline"> = {
  draft: "outline",
  scheduling: "secondary",
  review: "default",
  approved: "default",
  rejected: "destructive",
};

const columns: ColumnDef<Semester, unknown>[] = [
  { accessorKey: "name", header: "Name" },
  {
    accessorKey: "start_date",
    header: "Start",
    cell: ({ getValue }) => new Date(getValue() as string).toLocaleDateString(),
  },
  {
    accessorKey: "end_date",
    header: "End",
    cell: ({ getValue }) => new Date(getValue() as string).toLocaleDateString(),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ getValue }) => {
      const status = getValue() as SemesterStatus;
      return <Badge variant={statusColors[status]}>{status.toUpperCase()}</Badge>;
    },
  },
];

export function SemesterListPage() {
  const [offset, setOffset] = useState(0);
  const [showCreate, setShowCreate] = useState(false);
  const navigate = useNavigate();
  const { data, isLoading } = useSemesters({ offset, limit: 20 });

  const columnsWithActions: ColumnDef<Semester, unknown>[] = [
    ...columns,
    {
      id: "actions",
      header: "Actions",
      cell: ({ row }) => (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => navigate({ to: "/timetable/$semesterId", params: { semesterId: row.original.id } })}
        >
          <Eye className="h-4 w-4 mr-1" /> View
        </Button>
      ),
    },
  ];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Timetable - Semesters</h2>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4 mr-1" /> New Semester
        </Button>
      </div>
      <DataTable
        columns={columnsWithActions}
        data={data?.items ?? []}
        total={data?.total ?? 0}
        offset={offset}
        limit={20}
        onPaginationChange={setOffset}
        isLoading={isLoading}
      />
      <SemesterFormDialog open={showCreate} onOpenChange={setShowCreate} />
    </div>
  );
}
