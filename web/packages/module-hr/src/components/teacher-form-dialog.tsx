// Create/edit teacher form dialog using react-hook-form + zod validation.
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { FormDialog, Input, Label, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@mcs-erp/ui";
import { useCreateTeacher, useUpdateTeacher, useDepartments } from "@mcs-erp/api-client";
import type { Teacher } from "@mcs-erp/api-client";

const teacherSchema = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Invalid email address"),
  department_id: z.string().optional(),
  qualifications: z.string(), // comma-separated, split on submit
  is_active: z.boolean().optional(),
});

type TeacherFormValues = z.infer<typeof teacherSchema>;

interface TeacherFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  teacher?: Teacher; // if provided → edit mode
  onSuccess?: () => void;
}

export function TeacherFormDialog({
  open,
  onOpenChange,
  teacher,
  onSuccess,
}: TeacherFormDialogProps) {
  const isEdit = !!teacher;
  const { data: deptData } = useDepartments();
  const departments = deptData?.items ?? [];

  const createTeacher = useCreateTeacher();
  const updateTeacher = useUpdateTeacher(teacher?.id ?? "");

  const form = useForm<TeacherFormValues>({
    resolver: zodResolver(teacherSchema),
    defaultValues: {
      name: "",
      email: "",
      department_id: undefined,
      qualifications: "",
      is_active: true,
    },
  });

  // Populate form when editing an existing teacher.
  useEffect(() => {
    if (open && teacher) {
      form.reset({
        name: teacher.name,
        email: teacher.email,
        department_id: teacher.department_id ?? undefined,
        qualifications: teacher.qualifications.join(", "),
        is_active: teacher.is_active,
      });
    } else if (open && !teacher) {
      form.reset({ name: "", email: "", department_id: undefined, qualifications: "", is_active: true });
    }
  }, [open, teacher, form]);

  const onSubmit = form.handleSubmit(async (values) => {
    const qualifications = values.qualifications
      .split(",")
      .map((q) => q.trim())
      .filter(Boolean);

    if (isEdit && teacher) {
      await updateTeacher.mutateAsync({
        name: values.name,
        email: values.email,
        department_id: values.department_id,
        qualifications,
        is_active: values.is_active ?? true,
      });
    } else {
      await createTeacher.mutateAsync({
        name: values.name,
        email: values.email,
        department_id: values.department_id,
        qualifications,
      });
    }
    onOpenChange(false);
    onSuccess?.();
  });

  const isPending = createTeacher.isPending || updateTeacher.isPending;

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit Teacher" : "Add Teacher"}
      description={isEdit ? "Update teacher information." : "Fill in the details to add a new teacher."}
      onSubmit={onSubmit}
      submitLabel={isEdit ? "Update" : "Create"}
      isSubmitting={isPending}
    >
      <div className="space-y-3">
        <div>
          <Label htmlFor="teacher-name">Name</Label>
          <Input
            id="teacher-name"
            {...form.register("name")}
            placeholder="Full name"
          />
          {form.formState.errors.name && (
            <p className="text-xs text-destructive mt-1">{form.formState.errors.name.message}</p>
          )}
        </div>

        <div>
          <Label htmlFor="teacher-email">Email</Label>
          <Input
            id="teacher-email"
            type="email"
            {...form.register("email")}
            placeholder="email@example.com"
          />
          {form.formState.errors.email && (
            <p className="text-xs text-destructive mt-1">{form.formState.errors.email.message}</p>
          )}
        </div>

        <div>
          <Label htmlFor="teacher-dept">Department</Label>
          <Select
            value={form.watch("department_id") ?? "none"}
            onValueChange={(v) =>
              form.setValue("department_id", v === "none" ? undefined : v)
            }
          >
            <SelectTrigger id="teacher-dept">
              <SelectValue placeholder="Select department" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">No department</SelectItem>
              {departments.map((dept) => (
                <SelectItem key={dept.id} value={dept.id}>
                  {dept.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div>
          <Label htmlFor="teacher-quals">Qualifications</Label>
          <Input
            id="teacher-quals"
            {...form.register("qualifications")}
            placeholder="e.g. PhD, Mathematics, Computer Science"
          />
          <p className="text-xs text-muted-foreground mt-1">Comma-separated list</p>
        </div>
      </div>
    </FormDialog>
  );
}
