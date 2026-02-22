// Room list page: filterable, paginated table with create/edit actions.
import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Button, DataTable } from "@mcs-erp/ui";
import { useRooms } from "@mcs-erp/api-client";
import type { Room, RoomFilter } from "@mcs-erp/api-client";
import { Plus } from "lucide-react";
import { getRoomColumns } from "./room-columns";
import { RoomFilters } from "./room-filters";
import { RoomFormDialog } from "./room-form-dialog";

const LIMIT = 20;

export function RoomListPage() {
  const navigate = useNavigate();
  const [offset, setOffset] = useState(0);
  const [filters, setFilters] = useState<RoomFilter>({});
  const [showCreate, setShowCreate] = useState(false);
  const [editRoom, setEditRoom] = useState<Room | undefined>(undefined);

  const { data, isLoading } = useRooms({ ...filters, offset, limit: LIMIT });

  const handleFilterChange = (next: RoomFilter) => {
    setFilters(next);
    setOffset(0);
  };

  const columns = getRoomColumns({
    onView: (room) => navigate({ to: "/rooms/$roomId", params: { roomId: room.id } }),
    onEdit: (room) => setEditRoom(room),
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Rooms</h2>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="h-4 w-4 mr-1" /> Add Room
        </Button>
      </div>

      <RoomFilters filters={filters} onChange={handleFilterChange} />

      <DataTable
        columns={columns}
        data={data?.items ?? []}
        total={data?.total ?? 0}
        offset={offset}
        limit={LIMIT}
        onPaginationChange={setOffset}
        isLoading={isLoading}
      />

      <RoomFormDialog open={showCreate} onOpenChange={setShowCreate} />
      <RoomFormDialog
        open={!!editRoom}
        onOpenChange={(open) => { if (!open) setEditRoom(undefined); }}
        room={editRoom}
      />
    </div>
  );
}
