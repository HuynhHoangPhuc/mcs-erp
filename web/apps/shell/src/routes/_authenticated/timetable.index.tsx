import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { SemesterListPage } from "@mcs-erp/module-timetable";

export const timetableRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/timetable",
  component: SemesterListPage,
});
