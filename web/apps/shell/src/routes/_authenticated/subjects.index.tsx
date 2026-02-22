import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { SubjectListPage } from "@mcs-erp/module-subject";

export const subjectsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/subjects",
  component: SubjectListPage,
});
