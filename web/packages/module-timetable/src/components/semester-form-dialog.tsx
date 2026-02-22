import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { FormDialog, Input, Label } from "@mcs-erp/ui";
import { useCreateSemester } from "@mcs-erp/api-client";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  start_date: z.string().min(1, "Start date is required"),
  end_date: z.string().min(1, "End date is required"),
}).refine((d) => new Date(d.end_date) > new Date(d.start_date), {
  message: "End date must be after start date",
  path: ["end_date"],
});

type FormData = z.infer<typeof schema>;

interface SemesterFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function SemesterFormDialog({ open, onOpenChange }: SemesterFormDialogProps) {
  const createMutation = useCreateSemester();
  const { register, handleSubmit, reset, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  });

  const onSubmit = handleSubmit(async (data) => {
    await createMutation.mutateAsync({
      name: data.name,
      start_date: new Date(data.start_date).toISOString(),
      end_date: new Date(data.end_date).toISOString(),
    });
    reset();
    onOpenChange(false);
  });

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Create Semester"
      onSubmit={onSubmit}
      isSubmitting={createMutation.isPending}
    >
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input id="name" {...register("name")} placeholder="Fall 2026" />
        {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
      </div>
      <div className="space-y-2">
        <Label htmlFor="start_date">Start Date</Label>
        <Input id="start_date" type="date" {...register("start_date")} />
        {errors.start_date && <p className="text-sm text-destructive">{errors.start_date.message}</p>}
      </div>
      <div className="space-y-2">
        <Label htmlFor="end_date">End Date</Label>
        <Input id="end_date" type="date" {...register("end_date")} />
        {errors.end_date && <p className="text-sm text-destructive">{errors.end_date.message}</p>}
      </div>
    </FormDialog>
  );
}
