import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { Card, CardContent, CardHeader, CardTitle } from "@mcs-erp/ui";
import { useTeachers, useSubjects, useRooms, useSemesters } from "@mcs-erp/api-client";
import { Users, BookOpen, DoorOpen, Calendar } from "lucide-react";

export const dashboardRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/",
  component: DashboardPage,
});

function DashboardPage() {
  const { data: teachers } = useTeachers({ limit: 1 });
  const { data: subjects } = useSubjects({ limit: 1 });
  const { data: rooms } = useRooms({ limit: 1 });
  const { data: semesters } = useSemesters({ limit: 1 });

  const stats = [
    { label: "Teachers", value: teachers?.total ?? 0, icon: Users },
    { label: "Subjects", value: subjects?.total ?? 0, icon: BookOpen },
    { label: "Rooms", value: rooms?.total ?? 0, icon: DoorOpen },
    { label: "Semesters", value: semesters?.total ?? 0, icon: Calendar },
  ];

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Dashboard</h2>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <Card key={stat.label}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">{stat.label}</CardTitle>
              <stat.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
