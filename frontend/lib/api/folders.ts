"use client";

import { apiFetch } from "./client";

export interface Folder {
  id: string;
  name: string;
  parent_id?: string;
  depth: number;
  tag_ids: string[];
  created_at: string;
  updated_at: string;
}

export interface FolderView {
  folder?: Folder;
  breadcrumbs?: Folder[];
  children: Folder[];
}

export const foldersApi = {
  list: (parentID?: string | null) => {
    const qs = parentID ? `?parent_id=${parentID}` : "";
    return apiFetch<Folder[]>(`/folders${qs}`);
  },
  get: (id: string) => apiFetch<FolderView>(`/folders/${id}`),
  create: (name: string, parentID?: string | null) =>
    apiFetch<Folder>("/folders", {
      method: "POST",
      body: JSON.stringify({ name, parent_id: parentID ?? null }),
    }),
  rename: (id: string, name: string) =>
    apiFetch<void>(`/folders/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ name }),
    }),
  move: (id: string, newParentID?: string | null) =>
    apiFetch<void>(`/folders/${id}/move`, {
      method: "POST",
      body: JSON.stringify({ new_parent_id: newParentID ?? null }),
    }),
  remove: (id: string, recursive = false) =>
    apiFetch<void>(`/folders/${id}${recursive ? "?recursive=true" : ""}`, {
      method: "DELETE",
    }),
};
