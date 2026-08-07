"use client";

import { apiFetch } from "./client";

export type ActivitySubjectType = "file" | "folder" | "tag" | "storage" | "share";

export interface ActivityEntry {
  id: string;
  subject_type: ActivitySubjectType;
  subject_id: string;
  action: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export const activityApi = {
  listForSubject: (subjectID: string, limit = 50) => {
    const qs = new URLSearchParams({
      subject_id: subjectID,
      limit: String(limit),
    }).toString();
    return apiFetch<ActivityEntry[]>(`/activity?${qs}`);
  },
  recent: (limit = 20) => {
    const qs = new URLSearchParams({ limit: String(limit) }).toString();
    return apiFetch<ActivityEntry[]>(`/activity/recent?${qs}`);
  },
};
