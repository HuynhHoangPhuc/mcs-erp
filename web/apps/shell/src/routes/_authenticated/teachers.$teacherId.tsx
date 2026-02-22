import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { TeacherDetailPage } from "@mcs-erp/module-hr";

export const teacherDetailRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/teachers/$teacherId",
  component: TeacherDetailPage,
});
