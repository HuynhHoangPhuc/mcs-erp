import { useMatches } from "@tanstack/react-router";
import { useAuth } from "../hooks/use-auth";
import { Button } from "@mcs-erp/ui";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger,
} from "@mcs-erp/ui";
import { LogOut, User, ChevronRight } from "lucide-react";

export function AppHeader() {
  const { user, logout } = useAuth();
  const matches = useMatches();

  // Build breadcrumbs from route matches (skip root + authenticated layout)
  const crumbs = matches
    .filter((m) => m.pathname !== "/" || matches.length <= 3)
    .slice(2) // skip __root and _authenticated
    .map((m) => {
      const segments = m.pathname.split("/").filter(Boolean);
      const label = segments[segments.length - 1] ?? "Dashboard";
      return {
        label: label.charAt(0).toUpperCase() + label.slice(1),
        path: m.pathname,
      };
    });

  if (crumbs.length === 0) {
    crumbs.push({ label: "Dashboard", path: "/" });
  }

  return (
    <header className="flex h-14 items-center justify-between border-b bg-background px-6">
      {/* Breadcrumbs */}
      <nav className="flex items-center gap-1 text-sm">
        {crumbs.map((crumb, i) => (
          <span key={crumb.path} className="flex items-center gap-1">
            {i > 0 && <ChevronRight className="h-3 w-3 text-muted-foreground" />}
            <span className={i === crumbs.length - 1 ? "font-medium" : "text-muted-foreground"}>
              {crumb.label}
            </span>
          </span>
        ))}
      </nav>

      {/* User menu */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" className="gap-2">
            <User className="h-4 w-4" />
            <span className="hidden sm:inline">{user?.email}</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuLabel>{user?.email}</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={logout}>
            <LogOut className="mr-2 h-4 w-4" />
            Logout
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </header>
  );
}
