import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../lib/api-client";
import { queryKeys } from "../query-keys";
import type {
  Semester, CreateSemesterRequest, SemesterSubject,
  SetSubjectsRequest, AssignTeacherRequest,
} from "../types/timetable";
import type { ListResponse, PaginationParams } from "../types/common";

export function useSemesters(params?: PaginationParams) {
  const qs = new URLSearchParams();
  if (params?.offset) qs.set("offset", String(params.offset));
  if (params?.limit) qs.set("limit", String(params.limit));
  const q = qs.toString();

  return useQuery({
    queryKey: queryKeys.semesters.list(),
    queryFn: () => apiFetch<ListResponse<Semester>>(`/timetable/semesters${q ? `?${q}` : ""}`),
  });
}

export function useSemester(id: string) {
  return useQuery({
    queryKey: queryKeys.semesters.detail(id),
    queryFn: () => apiFetch<Semester>(`/timetable/semesters/${id}`),
    enabled: !!id,
  });
}

export function useCreateSemester() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSemesterRequest) =>
      apiFetch<Semester>("/timetable/semesters", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.semesters.all }),
  });
}

export function useSemesterSubjects(id: string) {
  return useQuery({
    queryKey: queryKeys.semesters.subjects(id),
    queryFn: () => apiFetch<SemesterSubject[]>(`/timetable/semesters/${id}/subjects`),
    enabled: !!id,
  });
}

export function useSetSemesterSubjects(semesterId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: SetSubjectsRequest) =>
      apiFetch(`/timetable/semesters/${semesterId}/subjects`, { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.semesters.subjects(semesterId) }),
  });
}

export function useAssignTeacher(semesterId: string, subjectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: AssignTeacherRequest) =>
      apiFetch(`/timetable/semesters/${semesterId}/subjects/${subjectId}/teacher`, {
        method: "POST",
        body: JSON.stringify(data),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.semesters.subjects(semesterId) }),
  });
}
