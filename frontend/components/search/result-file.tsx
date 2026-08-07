"use client";

import Link from "next/link";
import { FileItem } from "@/lib/api/files";
import { Tag } from "@/lib/api/tags";
import { FileIcon } from "@/components/ui/v2/file-icon";
import { Pill, PillColor } from "@/components/ui/v2/pill";
import { highlight } from "./highlight";

function formatBytes(n: number) {
  if (!n) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(n) / Math.log(1024));
  return `${(n / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

function relTime(iso?: string) {
  if (!iso) return "—";
  const d = new Date(iso);
  const diff = (Date.now() - d.getTime()) / 1000;
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 86400 * 7) return `${Math.floor(diff / 86400)}d ago`;
  return d.toLocaleDateString();
}

const pillColorFor = (hex: string): PillColor => {
  const h = hex.toLowerCase();
  if (h === "#10b981") return "emerald";
  if (h === "#6366f1") return "indigo";
  if (h === "#f59e0b") return "amber";
  if (h === "#f43f5e") return "rose";
  if (h === "#0ea5e9") return "sky";
  if (h === "#8b5cf6") return "violet";
  return "slate";
};

export interface ResultFileProps {
  file: FileItem;
  query: string;
  tagByID: Record<string, Tag>;
  folderPath?: string;
}

export function ResultFile({ file, query, tagByID, folderPath }: ResultFileProps) {
  const attached = file.tag_ids.map((id) => tagByID[id]).filter(Boolean);
  return (
    <Link
      href={`/dashboard/files/view/${file.id}`}
      className="flex items-center gap-3 rounded-lg border border-border bg-surface px-3 py-2.5 hover:border-border-strong transition-colors"
    >
      <FileIcon mime={file.mime_type} name={file.name} size="md" />
      <div className="min-w-0 flex-1">
        <div className="truncate text-[13px] font-medium text-text">
          {highlight(file.name, query)}
        </div>
        <div className="flex items-center gap-2 text-[11px] text-text-3 mt-0.5">
          <span>{formatBytes(file.size_bytes)}</span>
          <span>·</span>
          <span className="font-mono">{file.mime_type}</span>
          <span>·</span>
          <span>{relTime(file.uploaded_at ?? file.created_at)}</span>
          {folderPath && (
            <>
              <span>·</span>
              <span className="truncate">{folderPath}</span>
            </>
          )}
        </div>
      </div>
      {attached.length > 0 && (
        <div className="hidden md:flex items-center gap-1">
          {attached.slice(0, 3).map((t) => (
            <Pill key={t.id} name={t.name} color={pillColorFor(t.color)} />
          ))}
          {attached.length > 3 && (
            <span className="text-[10px] text-text-3">+{attached.length - 3}</span>
          )}
        </div>
      )}
    </Link>
  );
}
