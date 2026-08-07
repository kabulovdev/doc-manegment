"use client";

import { useAuthStore } from "../stores/auth-store";
import type { ApiErrorBody, AuthResponse, User } from "../types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080/api/v1";

export class ApiError extends Error {
  status: number;
  code?: string;
  constructor(status: number, body: ApiErrorBody) {
    super(body.message ?? body.error);
    this.status = status;
    this.code = body.code ?? body.error;
  }
}

let refreshInFlight: Promise<boolean> | null = null;

async function tryRefresh(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    try {
      const res = await fetch(`${API_BASE}/auth/refresh`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) return false;
      const data = (await res.json()) as AuthResponse;
      useAuthStore.getState().setAuth(data.access_token, data.user);
      return true;
    } catch {
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
  { retryOn401 = true }: { retryOn401?: boolean } = {},
): Promise<T> {
  const token = useAuthStore.getState().accessToken;
  const headers = new Headers(init.headers);
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
    credentials: "include",
  });

  if (res.status === 401 && retryOn401) {
    const ok = await tryRefresh();
    if (ok) return apiFetch<T>(path, init, { retryOn401: false });
    useAuthStore.getState().clear();
  }

  if (!res.ok) {
    let body: ApiErrorBody = { error: "http_error" };
    try {
      body = (await res.json()) as ApiErrorBody;
    } catch {
      /* ignore */
    }
    throw new ApiError(res.status, body);
  }

  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const authApi = {
  register: (input: { email: string; password: string; display_name?: string }) =>
    apiFetch<{ id: string; email: string; display_name: string }>(
      "/auth/register",
      { method: "POST", body: JSON.stringify(input) },
      { retryOn401: false },
    ),
  login: (input: { email: string; password: string }) =>
    apiFetch<AuthResponse>(
      "/auth/login",
      { method: "POST", body: JSON.stringify(input) },
      { retryOn401: false },
    ),
  logout: () =>
    apiFetch<void>("/auth/logout", { method: "POST" }, { retryOn401: false }),
  me: () => apiFetch<User>("/auth/me"),
  updateMe: (input: { display_name: string }) =>
    apiFetch<User>("/auth/me", {
      method: "PATCH",
      body: JSON.stringify(input),
    }),
  refresh: tryRefresh,
};
