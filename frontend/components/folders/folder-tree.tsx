"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { foldersApi, Folder } from "@/lib/api/folders";
import { cn } from "@/lib/utils";
import { Icon } from "@/components/ui/v2/icon";

interface Props {
  activeID?: string | null;
}

const smartViews = [
  { icon: "Star", label: "Starred", mode: "starred", enabled: true },
  { icon: "Clock", label: "Recent", mode: "recent", enabled: false },
  { icon: "Share2", label: "Shared with me", mode: "shared", enabled: false },
  { icon: "Trash2", label: "Trash", mode: "trash", enabled: true },
];

export function FolderTree({ activeID }: Props) {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const mode = searchParams?.get("mode") ?? null;
  const atFilesRoot = pathname === "/dashboard/files";
  const rootActive = !activeID && !mode;
  return (
    <div className="text-[13px]">
      <div className="px-2 mb-1 text-[10px] font-semibold uppercase tracking-wide text-text-3">
        Folders
      </div>
      <div className="space-y-0.5">
        <Link
          href="/dashboard/files"
          className={cn(
            "flex items-center gap-2 rounded px-2 py-1.5 transition-colors",
            rootActive
              ? "bg-accent-soft text-accent-2 font-medium"
              : "text-text-2 hover:bg-surface-2 hover:text-text",
          )}
        >
          <Icon name="Folder" size={14} />
          <span className="truncate">Root</span>
        </Link>
        <Branch parentID={null} depth={0} activeID={activeID ?? null} />
      </div>

      <div className="px-2 mt-5 mb-1 text-[10px] font-semibold uppercase tracking-wide text-text-3">
        Smart views
      </div>
      <div className="space-y-0.5">
        {smartViews.map((v) => {
          const active = atFilesRoot && mode === v.mode;
          if (!v.enabled) {
            return (
              <button
                key={v.label}
                type="button"
                disabled
                className="flex w-full items-center gap-2 rounded px-2 py-1.5 text-text-3 opacity-60 cursor-not-allowed"
                title="Coming soon"
              >
                <Icon name={v.icon} size={14} />
                <span className="truncate">{v.label}</span>
              </button>
            );
          }
          return (
            <Link
              key={v.label}
              href={`/dashboard/files?mode=${v.mode}`}
              className={cn(
                "flex items-center gap-2 rounded px-2 py-1.5 transition-colors",
                active
                  ? "bg-accent-soft text-accent-2 font-medium"
                  : "text-text-2 hover:bg-surface-2 hover:text-text",
              )}
            >
              <Icon name={v.icon} size={14} />
              <span className="truncate">{v.label}</span>
            </Link>
          );
        })}
      </div>
    </div>
  );
}

function Branch({
  parentID,
  depth,
  activeID,
}: {
  parentID: string | null;
  depth: number;
  activeID: string | null;
}) {
  const { data = [] } = useQuery<Folder[]>({
    queryKey: ["folders", parentID],
    queryFn: () => foldersApi.list(parentID),
  });
  return (
    <>
      {data.map((f) => (
        <Node key={f.id} folder={f} depth={depth} activeID={activeID} />
      ))}
    </>
  );
}

function Node({
  folder,
  depth,
  activeID,
}: {
  folder: Folder;
  depth: number;
  activeID: string | null;
}) {
  const [open, setOpen] = useState(false);
  const { data: children = [] } = useQuery<Folder[]>({
    queryKey: ["folders", folder.id],
    queryFn: () => foldersApi.list(folder.id),
    enabled: open,
    staleTime: 60_000,
  });
  const active = activeID === folder.id;
  return (
    <div>
      <div
        className={cn(
          "group flex items-center gap-1 rounded pr-1.5 transition-colors",
          active
            ? "bg-accent-soft text-accent-2"
            : "text-text-2 hover:bg-surface-2 hover:text-text",
        )}
        style={{ paddingLeft: 4 + depth * 10 }}
      >
        <button
          type="button"
          onClick={(e) => {
            e.preventDefault();
            setOpen((v) => !v);
          }}
          aria-label={open ? "Collapse" : "Expand"}
          className="flex h-6 w-5 items-center justify-center text-text-3"
        >
          <Icon name={open ? "ChevronDown" : "ChevronRight"} size={12} />
        </button>
        <Link
          href={`/dashboard/files/${folder.id}`}
          className="flex flex-1 min-w-0 items-center gap-1.5 py-1 truncate"
        >
          <Icon name="Folder" size={13} className={cn(active && "text-accent-2")} />
          <span className="truncate">{folder.name}</span>
        </Link>
      </div>
      {open && children.length > 0 && (
        <Branch parentID={folder.id} depth={depth + 1} activeID={activeID} />
      )}
    </div>
  );
}
