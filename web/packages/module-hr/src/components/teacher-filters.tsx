// Filter bar for teacher list: search input, department select, status toggle.
import { useCallback, useEffect, useRef, useState } from "react";
import { Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@mcs-erp/ui";
import { useDepartments } from "@mcs-erp/api-client";
import type { TeacherFilter } from "@mcs-erp/api-client";

interface TeacherFiltersProps {
  filters: TeacherFilter & { search?: string };
  onChange: (filters: TeacherFilter & { search?: string }) => void;
}

// Debounce hook for search input to avoid excessive API calls.
function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
  return debounced;
}

export function TeacherFilters({ filters, onChange }: TeacherFiltersProps) {
  const [searchInput, setSearchInput] = useState(filters.search ?? "");
  const debouncedSearch = useDebounce(searchInput, 300);
  const isFirstRender = useRef(true);

  const { data: deptData } = useDepartments();
  const departments = deptData?.items ?? [];

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }
    onChange({ ...filters, search: debouncedSearch || undefined });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedSearch]);

  const handleDepartment = useCallback(
    (value: string) => {
      onChange({ ...filters, department_id: value === "all" ? undefined : value });
    },
    [filters, onChange]
  );

  const handleStatus = useCallback(
    (value: string) => {
      const status =
        value === "active" ? "active" : value === "inactive" ? "inactive" : undefined;
      onChange({ ...filters, status });
    },
    [filters, onChange]
  );

  return (
    <div className="flex flex-wrap gap-3 items-center">
      <Input
        placeholder="Search teachers..."
        value={searchInput}
        onChange={(e) => setSearchInput(e.target.value)}
        className="w-64"
      />
      <Select
        value={filters.department_id ?? "all"}
        onValueChange={handleDepartment}
      >
        <SelectTrigger className="w-48">
          <SelectValue placeholder="All departments" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All departments</SelectItem>
          {departments.map((dept) => (
            <SelectItem key={dept.id} value={dept.id}>
              {dept.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select
        value={filters.status ?? "all"}
        onValueChange={handleStatus}
      >
        <SelectTrigger className="w-36">
          <SelectValue placeholder="All status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All status</SelectItem>
          <SelectItem value="active">Active</SelectItem>
          <SelectItem value="inactive">Inactive</SelectItem>
        </SelectContent>
      </Select>
    </div>
  );
}
