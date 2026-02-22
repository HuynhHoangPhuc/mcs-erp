import { useParams } from "@tanstack/react-router";
import { useSemester } from "@mcs-erp/api-client";
import { Tabs, TabsContent, TabsList, TabsTrigger, Badge, LoadingSpinner } from "@mcs-erp/ui";
import { SemesterSetupStep } from "./semester-setup-step";
import { SemesterAssignStep } from "./semester-assign-step";
import { SemesterReviewStep } from "./semester-review-step";

export function SemesterDetailPage() {
  const { semesterId } = useParams({ strict: false }) as { semesterId: string };
  const { data: semester, isLoading } = useSemester(semesterId);

  if (isLoading) {
    return <div className="flex justify-center py-12"><LoadingSpinner size="lg" /></div>;
  }

  if (!semester) {
    return <p className="text-muted-foreground">Semester not found.</p>;
  }

  const isReadOnly = semester.status === "approved";
  const defaultTab = semester.status === "review" || semester.status === "approved" ? "review" : "setup";

  return (
    <div>
      <div className="flex items-center gap-3 mb-6">
        <h2 className="text-2xl font-bold">{semester.name}</h2>
        <Badge variant={semester.status === "approved" ? "default" : "outline"}>
          {semester.status.toUpperCase()}
        </Badge>
      </div>

      <Tabs defaultValue={defaultTab}>
        <TabsList>
          <TabsTrigger value="setup" disabled={isReadOnly}>Setup</TabsTrigger>
          <TabsTrigger value="assign" disabled={isReadOnly}>Assign Teachers</TabsTrigger>
          <TabsTrigger value="review">Review & Approve</TabsTrigger>
        </TabsList>

        <TabsContent value="setup" className="mt-4">
          <SemesterSetupStep semesterId={semesterId} />
        </TabsContent>

        <TabsContent value="assign" className="mt-4">
          <SemesterAssignStep semesterId={semesterId} />
        </TabsContent>

        <TabsContent value="review" className="mt-4">
          <SemesterReviewStep semesterId={semesterId} semester={semester} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
