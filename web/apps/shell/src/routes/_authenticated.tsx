import { createRoute, Outlet, redirect } from "@tanstack/react-router";
import { rootRoute } from "./__root";

export const authenticatedRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: "authenticated",
  beforeLoad: () => {
    // Check if refresh token exists as a quick auth check.
    // Full validation happens server-side on API calls.
    const hasToken = !!localStorage.getItem("refresh_token");
    if (!hasToken) {
      throw redirect({ to: "/login" });
    }
  },
  component: () => <Outlet />,
});
