"use client";

import { apiFetch } from "./client";

export interface Tag {
  id: string;
  name: string;
  color: string;
  created_at: string;
}

export const tagsApi = {
  list: () => apiFetch<Tag[]>("/tags"),
  create: (name: string, color?: string) =>
    apiFetch<Tag>("/tags", {
      method: "POST",
      body: JSON.stringify({ name, color: color ?? "" }),
    }),
  update: (id: string, input: { name?: string; color?: string }) =>
    apiFetch<void>(`/tags/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ name: input.name ?? "", color: input.color ?? "" }),
    }),
  remove: (id: string) => apiFetch<void>(`/tags/${id}`, { method: "DELETE" }),
  attach: (tagID: string, targetType: "file" | "folder", targetID: string) =>
    apiFetch<void>(`/tags/${tagID}/attach`, {
      method: "POST",
      body: JSON.stringify({ target_type: targetType, target_id: targetID }),
    }),
  detach: (tagID: string, targetType: "file" | "folder", targetID: string) =>
    apiFetch<void>(`/tags/${tagID}/detach`, {
      method: "POST",
      body: JSON.stringify({ target_type: targetType, target_id: targetID }),
    }),
};
