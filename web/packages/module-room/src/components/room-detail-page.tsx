// Room detail page: info card + availability grid with save/edit actions.
import { useState } from "react";
import { useParams } from "@tanstack/react-router";
import {
  Button, Card, CardHeader, CardTitle, CardContent,
  Badge, AvailabilityGrid, LoadingSpinner,
} from "@mcs-erp/ui";
import { useRoom, useRoomAvailability, useUpdateRoomAvailability } from "@mcs-erp/api-client";
import type { RoomAvailabilitySlot } from "@mcs-erp/api-client";
import { RoomFormDialog } from "./room-form-dialog";

export function RoomDetailPage() {
  const { roomId } = useParams({ strict: false }) as { roomId: string };
  const [showEdit, setShowEdit] = useState(false);
  const [localSlots, setLocalSlots] = useState<RoomAvailabilitySlot[] | null>(null);

  const { data: room, isLoading: roomLoading } = useRoom(roomId);
  const { data: availability, isLoading: availLoading } = useRoomAvailability(roomId);
  const updateAvailability = useUpdateRoomAvailability(roomId);

  const slots = localSlots ?? availability ?? [];

  const handleSaveAvailability = async () => {
    await updateAvailability.mutateAsync(slots);
    setLocalSlots(null);
  };

  if (roomLoading) {
    return (
      <div className="flex justify-center py-12">
        <LoadingSpinner />
      </div>
    );
  }

  if (!room) {
    return <p className="text-muted-foreground">Room not found.</p>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">{room.name}</h2>
        <Button variant="outline" onClick={() => setShowEdit(true)}>Edit Room</Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Room Information</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-2 gap-x-8 gap-y-3 text-sm sm:grid-cols-3">
            <div>
              <dt className="text-muted-foreground">Code</dt>
              <dd className="font-medium">{room.code}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Building</dt>
              <dd className="font-medium">{room.building}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Floor</dt>
              <dd className="font-medium">{room.floor}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Capacity</dt>
              <dd className="font-medium">{room.capacity}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Status</dt>
              <dd>
                <Badge variant={room.is_active ? "default" : "outline"}>
                  {room.is_active ? "Active" : "Inactive"}
                </Badge>
              </dd>
            </div>
            <div className="col-span-2 sm:col-span-3">
              <dt className="text-muted-foreground mb-1">Equipment</dt>
              <dd className="flex flex-wrap gap-1">
                {room.equipment.length > 0
                  ? room.equipment.map((eq) => (
                      <Badge key={eq} variant="secondary">{eq}</Badge>
                    ))
                  : <span className="text-muted-foreground">None</span>
                }
              </dd>
            </div>
          </dl>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Availability</CardTitle>
            <Button
              size="sm"
              onClick={handleSaveAvailability}
              disabled={updateAvailability.isPending || !localSlots}
            >
              {updateAvailability.isPending ? "Saving..." : "Save Availability"}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {availLoading ? (
            <div className="flex justify-center py-6"><LoadingSpinner /></div>
          ) : (
            <AvailabilityGrid
              slots={slots}
              onChange={(updated) => setLocalSlots(updated)}
            />
          )}
        </CardContent>
      </Card>

      <RoomFormDialog
        open={showEdit}
        onOpenChange={setShowEdit}
        room={room}
      />
    </div>
  );
}
