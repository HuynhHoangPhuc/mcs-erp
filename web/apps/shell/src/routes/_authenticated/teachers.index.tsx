import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { TeacherListPage } from "@mcs-erp/module-hr";

export const teachersRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/teachers",
  component: TeacherListPage,
});
