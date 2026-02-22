import { useState } from "react";
import { Button, LoadingSpinner } from "@mcs-erp/ui";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
  Label,
} from "@mcs-erp/ui";
import { useUpdateAssignment } from "@mcs-erp/api-client";
import type { Assignment } from "@mcs-erp/api-client";

const DAYS = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

interface Props {
  assignment: Assignment;
  teachers: { id: string; name: string }[];
  rooms: { id: string; name: string }[];
  subjectName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AssignmentEditDialog({ assignment, teachers, rooms, subjectName, open, onOpenChange }: Props) {
  const [teacherId, setTeacherId] = useState(assignment.teacher_id);
  const [roomId, setRoomId] = useState(assignment.room_id);
  const [day, setDay] = useState(String(assignment.day));
  const [period, setPeriod] = useState(String(assignment.period));

  const mutation = useUpdateAssignment(assignment.id, assignment.semester_id);

  const handleSave = async () => {
    await mutation.mutateAsync({
      teacher_id: teacherId,
      room_id: roomId,
      day: Number(day),
      period: Number(period),
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Assignment: {subjectName}</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label>Teacher</Label>
            <Select value={teacherId} onValueChange={setTeacherId}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {teachers.map((t) => (
                  <SelectItem key={t.id} value={t.id}>{t.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-2">
            <Label>Room</Label>
            <Select value={roomId} onValueChange={setRoomId}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                {rooms.map((r) => (
                  <SelectItem key={r.id} value={r.id}>{r.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label>Day</Label>
              <Select value={day} onValueChange={setDay}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {DAYS.map((d, i) => (
                    <SelectItem key={i} value={String(i)}>{d}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label>Period</Label>
              <Select value={period} onValueChange={setPeriod}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {Array.from({ length: 10 }, (_, i) => (
                    <SelectItem key={i + 1} value={String(i + 1)}>Period {i + 1}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={handleSave} disabled={mutation.isPending}>
            {mutation.isPending ? <LoadingSpinner size="sm" /> : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
