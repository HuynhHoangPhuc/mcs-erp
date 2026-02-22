import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { CategoryListPage } from "@mcs-erp/module-subject";

export const categoriesRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/categories",
  component: CategoryListPage,
});
