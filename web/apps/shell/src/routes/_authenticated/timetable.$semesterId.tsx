import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { SemesterDetailPage } from "@mcs-erp/module-timetable";

export const semesterDetailRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/timetable/$semesterId",
  component: SemesterDetailPage,
});
