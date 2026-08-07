"use client";

import { apiFetch } from "./client";

export interface Storage {
  id: string;
  display_name: string;
  provider: "s3" | "r2" | "minio";
  endpoint: string;
  region: string;
  bucket: string;
  force_path_style: boolean;
  used_bytes: number;
  object_count: number;
  last_sync_at?: string;
  last_error?: string;
  created_at: string;
}

export interface CreateStorageInput {
  display_name: string;
  provider: "s3" | "r2" | "minio";
  endpoint: string;
  region?: string;
  bucket: string;
  access_key: string;
  secret_key: string;
  force_path_style?: boolean;
}

export const storagesApi = {
  list: () => apiFetch<Storage[]>("/storages"),
  create: (input: CreateStorageInput) =>
    apiFetch<Storage>("/storages", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  remove: (id: string, force = false) =>
    apiFetch<void>(`/storages/${id}${force ? "?force=true" : ""}`, {
      method: "DELETE",
    }),
  test: (id: string) =>
    apiFetch<{ ok: boolean; error?: string }>(`/storages/${id}/test`, {
      method: "POST",
    }),
  resync: (id: string) =>
    apiFetch<Storage>(`/storages/${id}/resync`, { method: "POST" }),
};
