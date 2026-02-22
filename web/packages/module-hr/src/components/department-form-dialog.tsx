// Create/edit department form dialog using react-hook-form + zod.
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { FormDialog, Input, Label } from "@mcs-erp/ui";
import { useCreateDepartment, useUpdateDepartment } from "@mcs-erp/api-client";
import type { Department } from "@mcs-erp/api-client";

const departmentSchema = z.object({
  name: z.string().min(1, "Name is required"),
  description: z.string(),
});

type DepartmentFormValues = z.infer<typeof departmentSchema>;

interface DepartmentFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  department?: Department; // if provided → edit mode
  onSuccess?: () => void;
}

export function DepartmentFormDialog({
  open,
  onOpenChange,
  department,
  onSuccess,
}: DepartmentFormDialogProps) {
  const isEdit = !!department;
  const createDepartment = useCreateDepartment();
  const updateDepartment = useUpdateDepartment(department?.id ?? "");

  const form = useForm<DepartmentFormValues>({
    resolver: zodResolver(departmentSchema),
    defaultValues: { name: "", description: "" },
  });

  useEffect(() => {
    if (open && department) {
      form.reset({ name: department.name, description: department.description });
    } else if (open && !department) {
      form.reset({ name: "", description: "" });
    }
  }, [open, department, form]);

  const onSubmit = form.handleSubmit(async (values) => {
    if (isEdit) {
      await updateDepartment.mutateAsync(values);
    } else {
      await createDepartment.mutateAsync(values);
    }
    onOpenChange(false);
    onSuccess?.();
  });

  const isPending = createDepartment.isPending || updateDepartment.isPending;

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit Department" : "Add Department"}
      description={isEdit ? "Update department details." : "Create a new department."}
      onSubmit={onSubmit}
      submitLabel={isEdit ? "Update" : "Create"}
      isSubmitting={isPending}
    >
      <div className="space-y-3">
        <div>
          <Label htmlFor="dept-name">Name</Label>
          <Input
            id="dept-name"
            {...form.register("name")}
            placeholder="Department name"
          />
          {form.formState.errors.name && (
            <p className="text-xs text-destructive mt-1">{form.formState.errors.name.message}</p>
          )}
        </div>

        <div>
          <Label htmlFor="dept-description">Description</Label>
          <Input
            id="dept-description"
            {...form.register("description")}
            placeholder="Brief description (optional)"
          />
        </div>
      </div>
    </FormDialog>
  );
}
