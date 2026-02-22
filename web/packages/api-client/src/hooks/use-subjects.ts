import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../lib/api-client";
import { queryKeys } from "../query-keys";
import type { Subject, CreateSubjectRequest, UpdateSubjectRequest, SubjectFilter, Prerequisite } from "../types/subject";
import type { ListResponse, PaginationParams } from "../types/common";

export function useSubjects(filters?: SubjectFilter & PaginationParams) {
  const params = new URLSearchParams();
  if (filters?.offset) params.set("offset", String(filters.offset));
  if (filters?.limit) params.set("limit", String(filters.limit));
  if (filters?.category_id) params.set("category_id", filters.category_id);
  if (filters?.search) params.set("search", filters.search);
  const qs = params.toString();

  return useQuery({
    queryKey: queryKeys.subjects.list(filters),
    queryFn: () => apiFetch<ListResponse<Subject>>(`/subjects${qs ? `?${qs}` : ""}`),
  });
}

export function useSubject(id: string) {
  return useQuery({
    queryKey: queryKeys.subjects.detail(id),
    queryFn: () => apiFetch<Subject>(`/subjects/${id}`),
    enabled: !!id,
  });
}

export function useCreateSubject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSubjectRequest) =>
      apiFetch<Subject>("/subjects", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.subjects.all }),
  });
}

export function useUpdateSubject(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateSubjectRequest) =>
      apiFetch<Subject>(`/subjects/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.subjects.all });
      qc.invalidateQueries({ queryKey: queryKeys.subjects.detail(id) });
    },
  });
}

export function useSubjectPrerequisites(id: string) {
  return useQuery({
    queryKey: queryKeys.subjects.prerequisites(id),
    queryFn: () => apiFetch<Prerequisite[]>(`/subjects/${id}/prerequisites`),
    enabled: !!id,
  });
}

export function useSubjectPrerequisiteChain(id: string) {
  return useQuery({
    queryKey: queryKeys.subjects.prerequisiteChain(id),
    queryFn: () => apiFetch<string[]>(`/subjects/${id}/prerequisite-chain`),
    enabled: !!id,
  });
}

export function useAddPrerequisite(subjectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (prerequisiteId: string) =>
      apiFetch(`/subjects/${subjectId}/prerequisites`, {
        method: "POST",
        body: JSON.stringify({ prerequisite_id: prerequisiteId }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.subjects.prerequisites(subjectId) });
      qc.invalidateQueries({ queryKey: queryKeys.subjects.prerequisiteChain(subjectId) });
    },
  });
}

export function useDeletePrerequisite(subjectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (prerequisiteId: string) =>
      apiFetch(`/subjects/${subjectId}/prerequisites/${prerequisiteId}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.subjects.prerequisites(subjectId) });
      qc.invalidateQueries({ queryKey: queryKeys.subjects.prerequisiteChain(subjectId) });
    },
  });
}
