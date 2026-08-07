"use client";

import Link from "next/link";
import { ActivityEntry, ActivitySubjectType } from "@/lib/api/activity";
import { Icon } from "@/components/ui/v2/icon";

interface Entry {
  entry: ActivityEntry;
}

function actionLabel(subject: ActivitySubjectType, action: string, meta?: Record<string, unknown>): string {
  if (subject === "file") {
    if (action === "upload_complete") return "Uploaded";
    if (action === "delete") return "Moved to Trash";
    if (action === "restore") return "Restored";
    if (action === "purge") return "Deleted forever";
    if (action === "rename") {
      const name = meta && typeof meta.name === "string" ? meta.name : null;
      return name ? `Renamed to “${name}”` : "Renamed";
    }
    if (action === "update") return "Updated";
  }
  if (subject === "folder") {
    if (action === "create") return "Created";
    if (action === "rename") {
      const name = meta && typeof meta.name === "string" ? meta.name : null;
      return name ? `Renamed to “${name}”` : "Renamed";
    }
    if (action === "move") return "Moved";
    if (action === "delete") return "Deleted";
  }
  if (subject === "share") {
    if (action === "create") return "Share link created";
    if (action === "revoke") return "Share revoked";
  }
  return action.replace(/_/g, " ");
}

function subjectIcon(subject: ActivitySubjectType): string {
  switch (subject) {
    case "file":
      return "File";
    case "folder":
      return "Folder";
    case "tag":
      return "Tag";
    case "storage":
      return "Cloud";
    case "share":
      return "Share2";
    default:
      return "Activity";
  }
}

function subjectTone(action: string): string {
  if (action === "delete" || action === "purge" || action === "revoke")
    return "text-danger";
  if (action === "restore") return "text-accent-2";
  return "text-text-3";
}

function subjectLink(entry: ActivityEntry): string | null {
  if (entry.subject_type === "file") {
    return `/dashboard/files/view/${entry.subject_id}`;
  }
  if (entry.subject_type === "folder") {
    return `/dashboard/files/${entry.subject_id}`;
  }
  if (entry.subject_type === "share") {
    return `/dashboard/shares`;
  }
  return null;
}

function relTime(iso: string) {
  const d = new Date(iso);
  const diff = (Date.now() - d.getTime()) / 1000;
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 86400 * 7) return `${Math.floor(diff / 86400)}d ago`;
  return d.toLocaleDateString();
}

export function ActivityRow({ entry }: Entry) {
  const label = actionLabel(entry.subject_type, entry.action, entry.metadata);
  const icon = subjectIcon(entry.subject_type);
  const tone = subjectTone(entry.action);
  const href = subjectLink(entry);
  const meta = entry.metadata ?? {};
  const title =
    typeof meta.name === "string"
      ? (meta.name as string)
      : entry.subject_type === "file" || entry.subject_type === "folder"
        ? entry.subject_id
        : entry.subject_type;
  const content = (
    <div className="flex items-center gap-2.5 text-[12px]">
      <span className={`inline-flex h-7 w-7 items-center justify-center rounded bg-surface-2 ${tone}`}>
        <Icon name={icon} size={12} />
      </span>
      <div className="min-w-0 flex-1">
        <div className="truncate text-text">
          <span className="font-medium">{label}</span>
          {title && entry.subject_type !== "share" && (
            <span className="text-text-3"> · {title}</span>
          )}
        </div>
        <div className="text-[11px] text-text-3">{relTime(entry.created_at)}</div>
      </div>
    </div>
  );
  if (href) {
    return (
      <Link
        href={href}
        className="block rounded px-1 py-1 -mx-1 hover:bg-surface-2"
      >
        {content}
      </Link>
    );
  }
  return <div className="px-1 py-1">{content}</div>;
}

export interface ActivityFeedProps {
  entries: ActivityEntry[];
  emptyTitle?: string;
  emptyDescription?: string;
}

export function ActivityFeed({ entries, emptyTitle, emptyDescription }: ActivityFeedProps) {
  if (entries.length === 0) {
    return (
      <div className="py-6 text-center">
        <div className="text-[12px] font-medium text-text-2">
          {emptyTitle ?? "No activity yet"}
        </div>
        {emptyDescription && (
          <div className="text-[11px] text-text-3 mt-1">{emptyDescription}</div>
        )}
      </div>
    );
  }
  return (
    <ul className="space-y-1">
      {entries.map((e) => (
        <li key={e.id}>
          <ActivityRow entry={e} />
        </li>
      ))}
    </ul>
  );
}
