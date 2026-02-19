// Fetch wrapper with JWT token management and auto-refresh on 401.

let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken(): string | null {
  return accessToken;
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");

  if (accessToken) {
    headers.set("Authorization", `Bearer ${accessToken}`);
  }

  const res = await fetch(`/api/v1${path}`, { ...options, headers });

  if (res.status === 401 && accessToken) {
    // Try refresh
    const refreshed = await tryRefresh();
    if (refreshed) {
      headers.set("Authorization", `Bearer ${accessToken}`);
      const retry = await fetch(`/api/v1${path}`, { ...options, headers });
      if (!retry.ok) {
        throw new ApiError(retry.status, await retry.text());
      }
      return retry.json();
    }
    // Refresh failed — clear token, redirect to login
    setAccessToken(null);
    window.location.href = "/login";
    throw new ApiError(401, "session expired");
  }

  if (!res.ok) {
    throw new ApiError(res.status, await res.text());
  }

  return res.json();
}

async function tryRefresh(): Promise<boolean> {
  const refreshToken = localStorage.getItem("refresh_token");
  if (!refreshToken) return false;

  try {
    const res = await fetch("/api/v1/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!res.ok) return false;

    const data = await res.json();
    setAccessToken(data.access_token);
    localStorage.setItem("refresh_token", data.refresh_token);
    return true;
  } catch {
    return false;
  }
}

export class ApiError extends Error {
  constructor(
    public status: number,
    public body: string,
  ) {
    super(`API error ${status}: ${body}`);
  }
}
