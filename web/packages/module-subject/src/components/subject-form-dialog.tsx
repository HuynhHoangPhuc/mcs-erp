// Form dialog for creating and editing subjects using react-hook-form + zod.
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { FormDialog, Input, Label } from "@mcs-erp/ui";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@mcs-erp/ui";
import {
  useCreateSubject, useUpdateSubject, useCategories,
  type Subject,
} from "@mcs-erp/api-client";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  code: z.string().min(1, "Code is required"),
  description: z.string(),
  category_id: z.string().optional(),
  credits: z.number().min(1).max(20),
  hours_per_week: z.number().min(1).max(40),
  is_active: z.boolean().optional(),
});

type FormValues = z.infer<typeof schema>;

interface SubjectFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialData?: Subject;
  onSuccess?: () => void;
}

export function SubjectFormDialog({
  open, onOpenChange, initialData, onSuccess,
}: SubjectFormDialogProps) {
  const { data: categoriesData } = useCategories();
  const createSubject = useCreateSubject();
  const updateSubject = useUpdateSubject(initialData?.id ?? "");

  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "", code: "", description: "", credits: 3, hours_per_week: 3, is_active: true,
    },
  });

  useEffect(() => {
    if (open) {
      reset(initialData
        ? {
            name: initialData.name,
            code: initialData.code,
            description: initialData.description,
            category_id: initialData.category_id ?? undefined,
            credits: initialData.credits,
            hours_per_week: initialData.hours_per_week,
            is_active: initialData.is_active,
          }
        : { name: "", code: "", description: "", credits: 3, hours_per_week: 3, is_active: true }
      );
    }
  }, [open, initialData, reset]);

  const isSubmitting = createSubject.isPending || updateSubject.isPending;
  const categoryId = watch("category_id");

  const onSubmit = handleSubmit(async (values) => {
    const payload = {
      name: values.name,
      code: values.code,
      description: values.description,
      credits: values.credits,
      hours_per_week: values.hours_per_week,
      ...(values.category_id ? { category_id: values.category_id } : {}),
    };
    if (initialData) {
      await updateSubject.mutateAsync({ ...payload, is_active: values.is_active ?? true });
    } else {
      await createSubject.mutateAsync(payload);
    }
    onSuccess?.();
    onOpenChange(false);
  });

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={initialData ? "Edit Subject" : "Add Subject"}
      onSubmit={onSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="grid gap-3">
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1">
            <Label htmlFor="name">Name</Label>
            <Input id="name" {...register("name")} placeholder="e.g. Introduction to CS" />
            {errors.name && <p className="text-destructive text-xs">{errors.name.message}</p>}
          </div>
          <div className="space-y-1">
            <Label htmlFor="code">Code</Label>
            <Input id="code" {...register("code")} placeholder="e.g. CS101" />
            {errors.code && <p className="text-destructive text-xs">{errors.code.message}</p>}
          </div>
        </div>

        <div className="space-y-1">
          <Label htmlFor="description">Description</Label>
          <textarea
            id="description"
            {...register("description")}
            placeholder="Subject description..."
            rows={3}
            className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex min-h-[80px] w-full rounded-md border px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
          />
        </div>

        <div className="space-y-1">
          <Label>Category</Label>
          <Select
            value={categoryId ?? "none"}
            onValueChange={(v) => setValue("category_id", v === "none" ? undefined : v)}
          >
            <SelectTrigger>
              <SelectValue placeholder="Select category" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">No category</SelectItem>
              {categoriesData?.items.map((cat) => (
                <SelectItem key={cat.id} value={cat.id}>{cat.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1">
            <Label htmlFor="credits">Credits</Label>
            <Input id="credits" type="number" min={1} max={20} {...register("credits")} />
            {errors.credits && <p className="text-destructive text-xs">{errors.credits.message}</p>}
          </div>
          <div className="space-y-1">
            <Label htmlFor="hours_per_week">Hours/Week</Label>
            <Input id="hours_per_week" type="number" min={1} max={40} {...register("hours_per_week")} />
            {errors.hours_per_week && <p className="text-destructive text-xs">{errors.hours_per_week.message}</p>}
          </div>
        </div>
      </div>
    </FormDialog>
  );
}
