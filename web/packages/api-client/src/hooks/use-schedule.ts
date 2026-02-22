import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../lib/api-client";
import { queryKeys } from "../query-keys";
import type { Schedule, Assignment, UpdateAssignmentRequest } from "../types/timetable";

export function useSchedule(semesterId: string) {
  return useQuery({
    queryKey: queryKeys.semesters.schedule(semesterId),
    queryFn: () => apiFetch<Schedule>(`/timetable/semesters/${semesterId}/schedule`),
    enabled: !!semesterId,
  });
}

export function useGenerateSchedule(semesterId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      apiFetch<Schedule>(`/timetable/semesters/${semesterId}/generate`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.semesters.schedule(semesterId) });
      qc.invalidateQueries({ queryKey: queryKeys.semesters.detail(semesterId) });
    },
  });
}

export function useApproveSemester(semesterId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      apiFetch(`/timetable/semesters/${semesterId}/approve`, { method: "POST" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.semesters.detail(semesterId) });
      qc.invalidateQueries({ queryKey: queryKeys.semesters.all });
    },
  });
}

export function useUpdateAssignment(assignmentId: string, semesterId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateAssignmentRequest) =>
      apiFetch<Assignment>(`/timetable/assignments/${assignmentId}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.semesters.schedule(semesterId) }),
  });
}
