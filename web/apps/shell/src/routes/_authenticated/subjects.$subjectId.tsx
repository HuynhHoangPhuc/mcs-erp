import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { SubjectDetailPage } from "@mcs-erp/module-subject";

export const subjectDetailRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/subjects/$subjectId",
  component: SubjectDetailPage,
});
