// Timetable module types matching backend semester/schedule handlers.

export type SemesterStatus = "draft" | "scheduling" | "review" | "approved" | "rejected";

export interface Semester {
  id: string;
  name: string;
  start_date: string;
  end_date: string;
  status: SemesterStatus;
  created_at: string;
  updated_at: string;
}

export interface CreateSemesterRequest {
  name: string;
  start_date: string;  // RFC3339
  end_date: string;
}

export interface SemesterSubject {
  semester_id: string;
  subject_id: string;
  teacher_id: string | null;
}

export interface SetSubjectsRequest {
  subject_ids: string[];
}

export interface AssignTeacherRequest {
  teacher_id: string;
}

export interface Assignment {
  id: string;
  semester_id: string;
  subject_id: string;
  teacher_id: string;
  room_id: string;
  day: number;     // 0-5 (Mon-Sat)
  period: number;  // 1-10
  version: number;
}

export interface Schedule {
  semester_id: string;
  version: number;
  assignments: Assignment[];
  hard_violations: number;
  soft_penalty: number;
  generated_at: string;
}

export interface UpdateAssignmentRequest {
  teacher_id?: string;
  room_id?: string;
  day?: number;
  period?: number;
}
