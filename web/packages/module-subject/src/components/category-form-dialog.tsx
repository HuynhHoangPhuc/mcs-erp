// Form dialog for creating and editing subject categories.
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { FormDialog, Input, Label } from "@mcs-erp/ui";
import { useCreateCategory, useUpdateCategory, type Category } from "@mcs-erp/api-client";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string(),
});

type FormValues = z.infer<typeof schema>;

interface CategoryFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialData?: Category;
  onSuccess?: () => void;
}

export function CategoryFormDialog({
  open, onOpenChange, initialData, onSuccess,
}: CategoryFormDialogProps) {
  const createCategory = useCreateCategory();
  const updateCategory = useUpdateCategory(initialData?.id ?? "");

  const { register, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", description: "" },
  });

  useEffect(() => {
    if (open) {
      reset(initialData
        ? { name: initialData.name, description: initialData.description }
        : { name: "", description: "" }
      );
    }
  }, [open, initialData, reset]);

  const isSubmitting = createCategory.isPending || updateCategory.isPending;

  const onSubmit = handleSubmit(async (values) => {
    if (initialData) {
      await updateCategory.mutateAsync(values);
    } else {
      await createCategory.mutateAsync(values);
    }
    onSuccess?.();
    onOpenChange(false);
  });

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={initialData ? "Edit Category" : "Add Category"}
      onSubmit={onSubmit}
      isSubmitting={isSubmitting}
    >
      <div className="space-y-3">
        <div className="space-y-1">
          <Label htmlFor="cat-name">Name</Label>
          <Input id="cat-name" {...register("name")} placeholder="e.g. Computer Science" />
          {errors.name && <p className="text-destructive text-xs">{errors.name.message}</p>}
        </div>
        <div className="space-y-1">
          <Label htmlFor="cat-description">Description</Label>
          <textarea
            id="cat-description"
            {...register("description")}
            placeholder="Category description..."
            rows={3}
            className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex min-h-[80px] w-full rounded-md border px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
          />
        </div>
      </div>
    </FormDialog>
  );
}
