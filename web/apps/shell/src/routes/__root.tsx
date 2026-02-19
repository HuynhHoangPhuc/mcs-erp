import { createRootRoute, Outlet } from "@tanstack/react-router";
import { AuthProvider } from "../providers/auth-provider";

export const rootRoute = createRootRoute({
  component: () => (
    <AuthProvider>
      <div>
        <header style={{ padding: "1rem", borderBottom: "1px solid #eee" }}>
          <h1>MCS-ERP</h1>
        </header>
        <main style={{ padding: "1rem" }}>
          <Outlet />
        </main>
      </div>
    </AuthProvider>
  ),
});
