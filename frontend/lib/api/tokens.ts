"use client";

import { apiFetch } from "./client";

export interface ApiToken {
  id: string;
  name: string;
  prefix: string;
  scopes: string[];
  created_at: string;
  last_used_at?: string | null;
  revoked_at?: string | null;
}

export interface CreateApiTokenResponse {
  token: ApiToken;
  plaintext: string;
}

export const tokensApi = {
  list: () => apiFetch<ApiToken[]>("/auth/tokens"),
  create: (input: { name: string; scopes?: string[] }) =>
    apiFetch<CreateApiTokenResponse>("/auth/tokens", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  revoke: (id: string) =>
    apiFetch<void>(`/auth/tokens/${id}`, { method: "DELETE" }),
  scopes: () =>
    apiFetch<{ scopes: string[] }>("/auth/token-scopes").then((r) => r.scopes),
};
