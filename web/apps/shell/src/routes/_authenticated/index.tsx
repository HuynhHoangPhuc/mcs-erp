import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { useAuth } from "../../hooks/use-auth";

export const dashboardRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/",
  component: DashboardPage,
});

function DashboardPage() {
  const { user, logout } = useAuth();

  return (
    <div>
      <h2>Dashboard</h2>
      <p>Welcome, {user?.email}</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
