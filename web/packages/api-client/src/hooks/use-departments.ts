import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../lib/api-client";
import { queryKeys } from "../query-keys";
import type { Department, CreateDepartmentRequest, UpdateDepartmentRequest } from "../types/hr";
import type { ListResponse } from "../types/common";

export function useDepartments() {
  return useQuery({
    queryKey: queryKeys.departments.list(),
    queryFn: () => apiFetch<ListResponse<Department>>("/departments"),
  });
}

export function useCreateDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateDepartmentRequest) =>
      apiFetch<Department>("/departments", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.departments.all }),
  });
}

export function useUpdateDepartment(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateDepartmentRequest) =>
      apiFetch<Department>(`/departments/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.departments.all }),
  });
}

export function useDeleteDepartment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiFetch(`/departments/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.departments.all }),
  });
}
