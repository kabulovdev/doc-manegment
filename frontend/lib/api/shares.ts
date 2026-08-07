"use client";

import { apiFetch } from "./client";

export interface Share {
  id: string;
  token: string;
  target_type: "file" | "folder";
  target_id: string;
  expires_at?: string | null;
  has_password: boolean;
  one_time_use: boolean;
  consumed: boolean;
  revoked: boolean;
  created_at: string;
}

export interface ShareAccessLogEntry {
  accessed_at: string;
  ip?: string;
  user_agent?: string;
  bytes_streamed?: number;
}

export interface ShareDetail {
  share: Share;
  access: ShareAccessLogEntry[];
}

export interface CreateShareInput {
  target_type: "file" | "folder";
  target_id: string;
  expires_at?: string | null;
  password?: string;
  one_time_use?: boolean;
}

export const sharesApi = {
  list: () => apiFetch<Share[]>("/shares"),
  get: (id: string) => apiFetch<ShareDetail>(`/shares/${id}`),
  create: (input: CreateShareInput) =>
    apiFetch<Share>("/shares", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  revoke: (id: string) => apiFetch<void>(`/shares/${id}`, { method: "DELETE" }),
};
