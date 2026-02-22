// TanStack Table column definitions for Room entity.
import type { ColumnDef } from "@tanstack/react-table";
import { Badge, Button } from "@mcs-erp/ui";
import { Eye } from "lucide-react";
import type { Room } from "@mcs-erp/api-client";

interface RoomColumnsOptions {
  onView: (room: Room) => void;
  onEdit: (room: Room) => void;
}

export function getRoomColumns({ onView, onEdit }: RoomColumnsOptions): ColumnDef<Room, unknown>[] {
  return [
    { accessorKey: "code", header: "Code" },
    { accessorKey: "name", header: "Name" },
    { accessorKey: "building", header: "Building" },
    { accessorKey: "floor", header: "Floor" },
    { accessorKey: "capacity", header: "Capacity" },
    {
      accessorKey: "equipment",
      header: "Equipment",
      cell: ({ getValue }) => {
        const items = getValue() as string[];
        if (!items || items.length === 0) return <span className="text-muted-foreground">—</span>;
        return (
          <div className="flex flex-wrap gap-1">
            {items.map((eq) => (
              <Badge key={eq} variant="secondary" className="text-xs">{eq}</Badge>
            ))}
          </div>
        );
      },
    },
    {
      accessorKey: "is_active",
      header: "Status",
      cell: ({ getValue }) => {
        const active = getValue() as boolean;
        return <Badge variant={active ? "default" : "outline"}>{active ? "Active" : "Inactive"}</Badge>;
      },
    },
    {
      id: "actions",
      header: "Actions",
      cell: ({ row }) => (
        <div className="flex gap-1">
          <Button variant="ghost" size="sm" onClick={() => onView(row.original)}>
            <Eye className="h-4 w-4 mr-1" /> View
          </Button>
          <Button variant="ghost" size="sm" onClick={() => onEdit(row.original)}>
            Edit
          </Button>
        </div>
      ),
    },
  ];
}
