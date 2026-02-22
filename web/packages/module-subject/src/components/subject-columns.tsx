// TanStack Table column definitions for Subject list view.
import type { ColumnDef } from "@tanstack/react-table";
import { Badge, Button, DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@mcs-erp/ui";
import { MoreHorizontal } from "lucide-react";
import type { Subject } from "@mcs-erp/api-client";

interface SubjectColumnsOptions {
  categoryMap: Record<string, string>;
  onEdit: (subject: Subject) => void;
  onViewDetail: (subject: Subject) => void;
}

export function buildSubjectColumns({
  categoryMap,
  onEdit,
  onViewDetail,
}: SubjectColumnsOptions): ColumnDef<Subject>[] {
  return [
    {
      accessorKey: "code",
      header: "Code",
      cell: ({ row }) => (
        <span className="font-mono text-sm font-medium">{row.original.code}</span>
      ),
    },
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <button
          className="text-left hover:underline font-medium"
          onClick={() => onViewDetail(row.original)}
        >
          {row.original.name}
        </button>
      ),
    },
    {
      accessorKey: "category_id",
      header: "Category",
      cell: ({ row }) => {
        const catId = row.original.category_id;
        return catId ? (
          <span className="text-sm">{categoryMap[catId] ?? "Unknown"}</span>
        ) : (
          <span className="text-muted-foreground text-sm">—</span>
        );
      },
    },
    {
      accessorKey: "credits",
      header: "Credits",
      cell: ({ row }) => <span className="text-sm">{row.original.credits}</span>,
    },
    {
      accessorKey: "hours_per_week",
      header: "Hrs/Week",
      cell: ({ row }) => <span className="text-sm">{row.original.hours_per_week}</span>,
    },
    {
      accessorKey: "is_active",
      header: "Status",
      cell: ({ row }) =>
        row.original.is_active ? (
          <Badge variant="default">Active</Badge>
        ) : (
          <Badge variant="secondary">Inactive</Badge>
        ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => onViewDetail(row.original)}>
              View Detail
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onEdit(row.original)}>
              Edit
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
    },
  ];
}
