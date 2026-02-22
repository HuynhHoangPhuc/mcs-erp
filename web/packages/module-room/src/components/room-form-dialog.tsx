// Create/edit room form dialog using react-hook-form + zod validation.
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { FormDialog, Input, Label } from "@mcs-erp/ui";
import { useCreateRoom, useUpdateRoom } from "@mcs-erp/api-client";
import type { Room } from "@mcs-erp/api-client";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  code: z.string().min(1, "Code is required"),
  building: z.string().min(1, "Building is required"),
  floor: z.number().int("Floor must be an integer"),
  capacity: z.number().int().min(1, "Capacity must be at least 1"),
  equipment: z.string(),
});

type FormData = z.infer<typeof schema>;

interface RoomFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  room?: Room;
}

// Converts equipment comma-separated string to/from string array.
function equipmentToString(eq: string[]): string {
  return eq.join(", ");
}

function stringToEquipment(s: string): string[] {
  return s
    .split(",")
    .map((e) => e.trim())
    .filter(Boolean);
}

export function RoomFormDialog({ open, onOpenChange, room }: RoomFormDialogProps) {
  const createMutation = useCreateRoom();
  const updateMutation = useUpdateRoom(room?.id ?? "");
  const isEdit = !!room;

  const { register, handleSubmit, reset, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      code: "",
      building: "",
      floor: 1,
      capacity: 30,
      equipment: "",
    },
  });

  useEffect(() => {
    if (open && room) {
      reset({
        name: room.name,
        code: room.code,
        building: room.building,
        floor: room.floor,
        capacity: room.capacity,
        equipment: equipmentToString(room.equipment),
      });
    } else if (open && !room) {
      reset({ name: "", code: "", building: "", floor: 1, capacity: 30, equipment: "" });
    }
  }, [open, room, reset]);

  const onSubmit = handleSubmit(async (data) => {
    const payload = {
      name: data.name,
      code: data.code,
      building: data.building,
      floor: data.floor,
      capacity: data.capacity,
      equipment: stringToEquipment(data.equipment),
    };
    if (isEdit) {
      await updateMutation.mutateAsync({ ...payload, is_active: room.is_active });
    } else {
      await createMutation.mutateAsync(payload);
    }
    reset();
    onOpenChange(false);
  });

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit Room" : "Add Room"}
      onSubmit={onSubmit}
      submitLabel={isEdit ? "Update" : "Create"}
      isSubmitting={isPending}
    >
      <div className="space-y-2">
        <Label htmlFor="room-code">Code</Label>
        <Input id="room-code" {...register("code")} placeholder="e.g. A101" />
        {errors.code && <p className="text-sm text-destructive">{errors.code.message}</p>}
      </div>
      <div className="space-y-2">
        <Label htmlFor="room-name">Name</Label>
        <Input id="room-name" {...register("name")} placeholder="e.g. Lecture Hall A" />
        {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="room-building">Building</Label>
          <Input id="room-building" {...register("building")} placeholder="e.g. A" />
          {errors.building && <p className="text-sm text-destructive">{errors.building.message}</p>}
        </div>
        <div className="space-y-2">
          <Label htmlFor="room-floor">Floor</Label>
          <Input id="room-floor" type="number" {...register("floor")} />
          {errors.floor && <p className="text-sm text-destructive">{errors.floor.message}</p>}
        </div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="room-capacity">Capacity</Label>
        <Input id="room-capacity" type="number" {...register("capacity")} />
        {errors.capacity && <p className="text-sm text-destructive">{errors.capacity.message}</p>}
      </div>
      <div className="space-y-2">
        <Label htmlFor="room-equipment">Equipment</Label>
        <Input id="room-equipment" {...register("equipment")} placeholder="projector, whiteboard, AC" />
        <p className="text-xs text-muted-foreground">Separate multiple items with commas</p>
        {errors.equipment && <p className="text-sm text-destructive">{errors.equipment.message}</p>}
      </div>
    </FormDialog>
  );
}
