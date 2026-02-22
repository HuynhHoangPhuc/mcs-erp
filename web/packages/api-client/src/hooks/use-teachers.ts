import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../lib/api-client";
import { queryKeys } from "../query-keys";
import type { Teacher, CreateTeacherRequest, UpdateTeacherRequest, TeacherFilter, AvailabilitySlot } from "../types/hr";
import type { ListResponse, PaginationParams } from "../types/common";

export function useTeachers(filters?: TeacherFilter & PaginationParams) {
  const params = new URLSearchParams();
  if (filters?.offset) params.set("offset", String(filters.offset));
  if (filters?.limit) params.set("limit", String(filters.limit));
  if (filters?.department_id) params.set("department_id", filters.department_id);
  if (filters?.status) params.set("status", filters.status);
  if (filters?.qualification) params.set("qualification", filters.qualification);
  const qs = params.toString();

  return useQuery({
    queryKey: queryKeys.teachers.list(filters),
    queryFn: () => apiFetch<ListResponse<Teacher>>(`/teachers${qs ? `?${qs}` : ""}`),
  });
}

export function useTeacher(id: string) {
  return useQuery({
    queryKey: queryKeys.teachers.detail(id),
    queryFn: () => apiFetch<Teacher>(`/teachers/${id}`),
    enabled: !!id,
  });
}

export function useCreateTeacher() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateTeacherRequest) =>
      apiFetch<Teacher>("/teachers", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.teachers.all }),
  });
}

export function useUpdateTeacher(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateTeacherRequest) =>
      apiFetch<Teacher>(`/teachers/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.teachers.all });
      qc.invalidateQueries({ queryKey: queryKeys.teachers.detail(id) });
    },
  });
}

export function useTeacherAvailability(id: string) {
  return useQuery({
    queryKey: queryKeys.teachers.availability(id),
    queryFn: () => apiFetch<AvailabilitySlot[]>(`/teachers/${id}/availability`),
    enabled: !!id,
  });
}

export function useUpdateTeacherAvailability(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (slots: AvailabilitySlot[]) =>
      apiFetch(`/teachers/${id}/availability`, { method: "PUT", body: JSON.stringify({ slots }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.teachers.availability(id) }),
  });
}
