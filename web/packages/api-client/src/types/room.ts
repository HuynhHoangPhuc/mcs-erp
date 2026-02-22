// Room module types matching backend room/availability handlers.

export interface Room {
  id: string;
  name: string;
  code: string;
  building: string;
  floor: number;
  capacity: number;
  equipment: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateRoomRequest {
  name: string;
  code: string;
  building: string;
  floor: number;
  capacity: number;
  equipment: string[];
}

export interface UpdateRoomRequest {
  name: string;
  code: string;
  building: string;
  floor: number;
  capacity: number;
  equipment: string[];
  is_active: boolean;
}

export interface RoomFilter {
  building?: string;
  min_capacity?: number;
  equipment?: string;
}

export interface RoomAvailabilitySlot {
  day: number;
  period: number;
  is_available: boolean;
}
