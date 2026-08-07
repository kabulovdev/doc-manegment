"use client";

import { apiFetch, authApi } from "./client";
import { useAuthStore } from "../stores/auth-store";

export interface CustomField {
  key: string;
  value: string;
  type: "text" | "number" | "date" | "boolean";
}

export interface FileItem {
  id: string;
  storage_id: string;
  folder_id?: string;
  name: string;
  size_bytes: number;
  mime_type: string;
  status: "uploading" | "ready" | "deleted";
  starred?: boolean;
  tag_ids: string[];
  custom_fields: CustomField[];
  ai_status?: "none" | "pending" | "ready" | "failed";
  uploaded_at?: string;
  created_at: string;
}

export interface InitUploadInput {
  storage_id: string;
  folder_id?: string | null;
  name: string;
  size: number;
  mime: string;
  custom_fields?: CustomField[];
  tag_ids?: string[];
}

export interface InitUploadResponse {
  file_id: string;
  upload_id?: string;
  object_key: string;
  part_size: number;
  multipart: boolean;
}

export const filesApi = {
  list: (params: Record<string, string | undefined> = {}) => {
    const qs = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v !== undefined && v !== "") as [string, string][],
    ).toString();
    return apiFetch<FileItem[]>(`/files${qs ? `?${qs}` : ""}`);
  },
  updateTags: (fileID: string, tagIDs: string[]) =>
    apiFetch<FileItem>(`/files/${fileID}`, {
      method: "PATCH",
      body: JSON.stringify({ tag_ids: tagIDs }),
    }),
  update: (
    fileID: string,
    patch: {
      name?: string;
      custom_fields?: CustomField[];
      tag_ids?: string[];
      folder_id?: string | null;
    },
  ) =>
    apiFetch<FileItem>(`/files/${fileID}`, {
      method: "PATCH",
      body: JSON.stringify(patch),
    }),
  get: (id: string) => apiFetch<FileItem>(`/files/${id}`),
  toggleStar: (id: string, starred: boolean) =>
    apiFetch<FileItem>(`/files/${id}/star`, {
      method: "POST",
      body: JSON.stringify({ starred }),
    }),
  restore: (id: string) =>
    apiFetch<FileItem>(`/files/${id}/restore`, { method: "POST" }),
  purge: (id: string) =>
    apiFetch<void>(`/files/${id}/purge`, { method: "POST" }),
  initUpload: (input: InitUploadInput) =>
    apiFetch<InitUploadResponse>("/files/upload/init", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  abortUpload: (fileID: string) =>
    apiFetch<void>(`/files/upload/${fileID}/abort`, { method: "POST" }),
  remove: (id: string) => apiFetch<void>(`/files/${id}`, { method: "DELETE" }),
};

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080/api/v1";

// Fetches file content as a Blob using the Bearer token (with auto-refresh on
// 401). Returns both the Blob (for passing directly to react-pdf) and the
// response Content-Type header.
export async function fetchFileBlob(
  fileID: string,
): Promise<{ blob: Blob; mime: string }> {
  const doFetch = () => {
    const token = useAuthStore.getState().accessToken;
    const headers: Record<string, string> = {};
    if (token) headers["Authorization"] = `Bearer ${token}`;
    return fetch(`${API_BASE}/files/${fileID}/content`, {
      headers,
      credentials: "include",
    });
  };
  let res = await doFetch();
  if (res.status === 401) {
    const refreshed = await authApi.refresh();
    if (refreshed) {
      res = await doFetch();
    }
  }
  if (!res.ok) {
    throw new Error(`Failed to fetch file: HTTP ${res.status}`);
  }
  const mime = res.headers.get("Content-Type") ?? "application/octet-stream";
  const blob = await res.blob();
  return { blob, mime };
}
