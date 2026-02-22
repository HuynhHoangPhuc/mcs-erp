import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { DepartmentListPage } from "@mcs-erp/module-hr";

export const departmentsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/departments",
  component: DepartmentListPage,
});
