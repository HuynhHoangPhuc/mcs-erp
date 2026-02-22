import { Link, useLocation } from "@tanstack/react-router";
import { cn } from "@mcs-erp/ui";
import {
  LayoutDashboard, Users, Building2, BookOpen, Tag, GitBranch,
  DoorOpen, Calendar, MessageSquare, ChevronLeft,
} from "lucide-react";
import { Button } from "@mcs-erp/ui";
import { useState } from "react";

interface NavItem {
  label: string;
  to: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface NavGroup {
  title: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    title: "Overview",
    items: [
      { label: "Dashboard", to: "/", icon: LayoutDashboard },
    ],
  },
  {
    title: "HR",
    items: [
      { label: "Teachers", to: "/teachers", icon: Users },
      { label: "Departments", to: "/departments", icon: Building2 },
    ],
  },
  {
    title: "Academic",
    items: [
      { label: "Subjects", to: "/subjects", icon: BookOpen },
      { label: "Categories", to: "/categories", icon: Tag },
      { label: "Prerequisites", to: "/subjects/prerequisites", icon: GitBranch },
    ],
  },
  {
    title: "Resources",
    items: [
      { label: "Rooms", to: "/rooms", icon: DoorOpen },
    ],
  },
  {
    title: "Scheduling",
    items: [
      { label: "Timetable", to: "/timetable", icon: Calendar },
    ],
  },
  {
    title: "AI",
    items: [
      { label: "Chat", to: "/chat", icon: MessageSquare },
    ],
  },
];

export function AppSidebar() {
  const [collapsed, setCollapsed] = useState(() => {
    return localStorage.getItem("sidebar-collapsed") === "true";
  });
  const location = useLocation();

  const toggleCollapse = () => {
    const next = !collapsed;
    setCollapsed(next);
    localStorage.setItem("sidebar-collapsed", String(next));
  };

  return (
    <aside
      className={cn(
        "flex flex-col border-r bg-sidebar-background text-sidebar-foreground transition-all duration-200",
        collapsed ? "w-16" : "w-60"
      )}
    >
      {/* Logo / Brand */}
      <div className="flex h-14 items-center border-b px-4">
        {!collapsed && <span className="text-lg font-bold text-sidebar-primary">MCS-ERP</span>}
        <Button
          variant="ghost"
          size="icon"
          className={cn("ml-auto h-8 w-8", collapsed && "mx-auto")}
          onClick={toggleCollapse}
        >
          <ChevronLeft className={cn("h-4 w-4 transition-transform", collapsed && "rotate-180")} />
        </Button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-2">
        {navGroups.map((group) => (
          <div key={group.title} className="mb-2">
            {!collapsed && (
              <div className="px-4 py-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                {group.title}
              </div>
            )}
            {group.items.map((item) => {
              const isActive = location.pathname === item.to ||
                (item.to !== "/" && location.pathname.startsWith(item.to));
              return (
                <Link
                  key={item.to}
                  to={item.to}
                  className={cn(
                    "flex items-center gap-3 px-4 py-2 text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
                    isActive && "bg-sidebar-accent text-sidebar-primary font-medium",
                    collapsed && "justify-center px-2"
                  )}
                  title={collapsed ? item.label : undefined}
                >
                  <item.icon className="h-4 w-4 shrink-0" />
                  {!collapsed && <span>{item.label}</span>}
                </Link>
              );
            })}
          </div>
        ))}
      </nav>
    </aside>
  );
}
