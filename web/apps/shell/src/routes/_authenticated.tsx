import { createRoute, redirect } from "@tanstack/react-router";
import { rootRoute } from "./__root";
import { AppLayout } from "../components/app-layout";

export const authenticatedRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: "authenticated",
  beforeLoad: () => {
    const hasToken = !!localStorage.getItem("refresh_token");
    if (!hasToken) {
      throw redirect({ to: "/login" });
    }
  },
  component: AppLayout,
});
