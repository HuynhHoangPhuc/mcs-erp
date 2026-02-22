import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../lib/api-client";
import { queryKeys } from "../query-keys";
import type { Room, CreateRoomRequest, UpdateRoomRequest, RoomFilter, RoomAvailabilitySlot } from "../types/room";
import type { ListResponse, PaginationParams } from "../types/common";

export function useRooms(filters?: RoomFilter & PaginationParams) {
  const params = new URLSearchParams();
  if (filters?.offset) params.set("offset", String(filters.offset));
  if (filters?.limit) params.set("limit", String(filters.limit));
  if (filters?.building) params.set("building", filters.building);
  if (filters?.min_capacity) params.set("min_capacity", String(filters.min_capacity));
  if (filters?.equipment) params.set("equipment", filters.equipment);
  const qs = params.toString();

  return useQuery({
    queryKey: queryKeys.rooms.list(filters),
    queryFn: () => apiFetch<ListResponse<Room>>(`/rooms${qs ? `?${qs}` : ""}`),
  });
}

export function useRoom(id: string) {
  return useQuery({
    queryKey: queryKeys.rooms.detail(id),
    queryFn: () => apiFetch<Room>(`/rooms/${id}`),
    enabled: !!id,
  });
}

export function useCreateRoom() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateRoomRequest) =>
      apiFetch<Room>("/rooms", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.rooms.all }),
  });
}

export function useUpdateRoom(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateRoomRequest) =>
      apiFetch<Room>(`/rooms/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.rooms.all });
      qc.invalidateQueries({ queryKey: queryKeys.rooms.detail(id) });
    },
  });
}

export function useRoomAvailability(id: string) {
  return useQuery({
    queryKey: queryKeys.rooms.availability(id),
    queryFn: () => apiFetch<RoomAvailabilitySlot[]>(`/rooms/${id}/availability`),
    enabled: !!id,
  });
}

export function useUpdateRoomAvailability(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (slots: RoomAvailabilitySlot[]) =>
      apiFetch(`/rooms/${id}/availability`, { method: "PUT", body: JSON.stringify({ slots }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.rooms.availability(id) }),
  });
}
