// Auth API types matching backend auth_handler.go

export interface LoginRequest {
  email: string;
  password: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface AuthUser {
  email: string;
  permissions: string[];
}
