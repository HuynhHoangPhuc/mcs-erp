// Subject module types matching backend subject/category/prerequisite handlers.

export interface Subject {
  id: string;
  name: string;
  code: string;
  description: string;
  category_id: string | null;
  credits: number;
  hours_per_week: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateSubjectRequest {
  name: string;
  code: string;
  description: string;
  category_id?: string;
  credits: number;
  hours_per_week: number;
}

export interface UpdateSubjectRequest {
  name: string;
  code: string;
  description: string;
  category_id?: string;
  credits: number;
  hours_per_week: number;
  is_active: boolean;
}

export interface SubjectFilter {
  category_id?: string;
  search?: string;
}

export interface Category {
  id: string;
  name: string;
  description: string;
  created_at: string;
}

export interface CreateCategoryRequest {
  name: string;
  description: string;
}

export interface Prerequisite {
  subject_id: string;
  prerequisite_id: string;
}
