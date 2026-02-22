// HR module types matching backend teacher/department/availability handlers.

export interface Teacher {
  id: string;
  name: string;
  email: string;
  department_id: string | null;
  qualifications: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateTeacherRequest {
  name: string;
  email: string;
  department_id?: string;
  qualifications: string[];
}

export interface UpdateTeacherRequest {
  name: string;
  email: string;
  department_id?: string;
  qualifications: string[];
  is_active: boolean;
}

export interface TeacherFilter {
  department_id?: string;
  status?: "active" | "inactive";
  qualification?: string;
}

export interface Department {
  id: string;
  name: string;
  description: string;
  head_teacher_id: string | null;
  created_at: string;
}

export interface CreateDepartmentRequest {
  name: string;
  description: string;
}

export interface UpdateDepartmentRequest {
  name: string;
  description: string;
}

export interface AvailabilitySlot {
  day: number;       // 0-6 (Mon-Sun)
  period: number;    // 1-10
  is_available: boolean;
}
