import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { PrerequisiteDagPage } from "@mcs-erp/module-subject";

export const prerequisitesRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/subjects/prerequisites",
  component: PrerequisiteDagPage,
});
