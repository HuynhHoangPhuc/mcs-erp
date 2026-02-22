// Subject detail page showing full subject info with edit capability.
import { useState } from "react";
import { useParams, useNavigate } from "@tanstack/react-router";
import { Button, Card, CardContent, CardHeader, CardTitle, Badge, LoadingSpinner } from "@mcs-erp/ui";
import { useSubject, useCategories } from "@mcs-erp/api-client";
import { SubjectFormDialog } from "./subject-form-dialog";
import { ArrowLeft } from "lucide-react";

export function SubjectDetailPage() {
  const { subjectId } = useParams({ strict: false }) as { subjectId: string };
  const navigate = useNavigate();
  const [editOpen, setEditOpen] = useState(false);

  const { data: subject, isLoading } = useSubject(subjectId);
  const { data: categoriesData } = useCategories();

  const categoryName = subject?.category_id
    ? categoriesData?.items.find((c) => c.id === subject.category_id)?.name
    : null;

  if (isLoading) return <LoadingSpinner />;
  if (!subject) return <p className="p-6 text-muted-foreground">Subject not found.</p>;

  return (
    <div className="space-y-6 p-6 max-w-2xl">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="sm" onClick={() => navigate({ to: "/subjects" })}>
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back
        </Button>
      </div>

      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{subject.name}</h1>
          <p className="text-muted-foreground font-mono text-sm mt-1">{subject.code}</p>
        </div>
        <div className="flex items-center gap-2">
          {subject.is_active ? (
            <Badge variant="default">Active</Badge>
          ) : (
            <Badge variant="secondary">Inactive</Badge>
          )}
          <Button onClick={() => setEditOpen(true)}>Edit</Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Details</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-muted-foreground">Category</p>
            <p className="font-medium">{categoryName ?? "—"}</p>
          </div>
          <div>
            <p className="text-muted-foreground">Credits</p>
            <p className="font-medium">{subject.credits}</p>
          </div>
          <div>
            <p className="text-muted-foreground">Hours / Week</p>
            <p className="font-medium">{subject.hours_per_week}</p>
          </div>
          <div>
            <p className="text-muted-foreground">Created</p>
            <p className="font-medium">{new Date(subject.created_at).toLocaleDateString()}</p>
          </div>
          {subject.description && (
            <div className="col-span-2">
              <p className="text-muted-foreground">Description</p>
              <p className="font-medium whitespace-pre-line">{subject.description}</p>
            </div>
          )}
        </CardContent>
      </Card>

      <SubjectFormDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        initialData={subject}
      />
    </div>
  );
}
