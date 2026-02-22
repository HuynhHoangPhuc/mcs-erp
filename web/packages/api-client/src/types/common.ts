// Shared API response types matching backend patterns.

export interface ListResponse<T> {
  items: T[];
  total: number;
}

export interface PaginationParams {
  offset?: number;
  limit?: number;
}

export interface ApiErrorResponse {
  error: string;
}
