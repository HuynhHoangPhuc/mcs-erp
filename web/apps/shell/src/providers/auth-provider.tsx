import { createContext, useCallback, useEffect, useState } from "react";
import type { ReactNode } from "react";
import { apiFetch, setAccessToken } from "../lib/api-client";

interface AuthUser {
  email: string;
  permissions: string[];
}

interface AuthContextType {
  user: AuthUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextType | null>(null);

interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Try restore session from stored refresh token on mount
  useEffect(() => {
    const refreshToken = localStorage.getItem("refresh_token");
    if (!refreshToken) {
      setIsLoading(false);
      return;
    }

    fetch("/api/v1/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
      .then((res) => {
        if (!res.ok) throw new Error("refresh failed");
        return res.json();
      })
      .then((data: TokenResponse) => {
        setAccessToken(data.access_token);
        localStorage.setItem("refresh_token", data.refresh_token);
        // Decode user from JWT payload (base64 middle segment)
        const payload = JSON.parse(atob(data.access_token.split(".")[1]));
        setUser({ email: payload.email, permissions: payload.permissions || [] });
      })
      .catch(() => {
        localStorage.removeItem("refresh_token");
        setAccessToken(null);
      })
      .finally(() => setIsLoading(false));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const data = await apiFetch<TokenResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });

    setAccessToken(data.access_token);
    localStorage.setItem("refresh_token", data.refresh_token);

    const payload = JSON.parse(atob(data.access_token.split(".")[1]));
    setUser({ email: payload.email, permissions: payload.permissions || [] });
  }, []);

  const logout = useCallback(() => {
    setAccessToken(null);
    localStorage.removeItem("refresh_token");
    setUser(null);
    apiFetch("/auth/logout", { method: "POST" }).catch(() => {});
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, isAuthenticated: !!user, isLoading, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}
