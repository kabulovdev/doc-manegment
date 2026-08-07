"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { filesApi, FileItem } from "@/lib/api/files";
import { foldersApi, Folder } from "@/lib/api/folders";
import { storagesApi, Storage } from "@/lib/api/storages";
import { tagsApi, Tag } from "@/lib/api/tags";
import {
  PaletteScope,
  useCommandPalette,
} from "@/lib/stores/command-palette-store";
import { cn } from "@/lib/utils";
import { FileIcon } from "@/components/ui/v2/file-icon";
import { Icon } from "@/components/ui/v2/icon";
import { Kbd } from "@/components/ui/v2/kbd";

interface ResultItem {
  id: string;
  kind: PaletteScope;
  label: string;
  secondary?: string;
  icon: React.ReactNode;
  onSelect: () => void;
}

const scopes: { value: PaletteScope; label: string; icon: string }[] = [
  { value: "all", label: "All", icon: "Sparkles" },
  { value: "files", label: "Files", icon: "File" },
  { value: "folders", label: "Folders", icon: "Folder" },
  { value: "tags", label: "Tags", icon: "Tag" },
  { value: "storages", label: "Storages", icon: "Cloud" },
];

function formatBytes(n: number) {
  if (!n) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
}

function Highlight({ text, query }: { text: string; query: string }) {
  if (!query) return <>{text}</>;
  const idx = text.toLowerCase().indexOf(query.toLowerCase());
  if (idx < 0) return <>{text}</>;
  return (
    <>
      {text.slice(0, idx)}
      <mark className="bg-accent-soft text-accent-2 rounded px-0.5">
        {text.slice(idx, idx + query.length)}
      </mark>
      {text.slice(idx + query.length)}
    </>
  );
}

export function CommandPalette() {
  const router = useRouter();
  const { open, scope, closePalette, setScope } = useCommandPalette();
  const [query, setQuery] = useState("");
  const [focus, setFocus] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        useCommandPalette.getState().togglePalette();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  useEffect(() => {
    if (!open) {
      setQuery("");
      setFocus(0);
      return;
    }
    setTimeout(() => inputRef.current?.focus(), 10);
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") closePalette();
    };
    document.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = prev;
      document.removeEventListener("keydown", onKey);
    };
  }, [open, closePalette]);

  const files = useQuery({
    queryKey: ["palette", "files", query],
    queryFn: () => filesApi.list(query ? { q: query } : {}),
    enabled: open && (scope === "all" || scope === "files"),
    staleTime: 15_000,
  });

  const folders = useQuery({
    queryKey: ["palette", "folders"],
    queryFn: () => foldersApi.list(),
    enabled: open && (scope === "all" || scope === "folders"),
    staleTime: 60_000,
  });

  const tags = useQuery({
    queryKey: ["palette", "tags"],
    queryFn: () => tagsApi.list(),
    enabled: open && (scope === "all" || scope === "tags"),
    staleTime: 60_000,
  });

  const storages = useQuery({
    queryKey: ["palette", "storages"],
    queryFn: () => storagesApi.list(),
    enabled: open && (scope === "all" || scope === "storages"),
    staleTime: 60_000,
  });

  const groups = useMemo(() => {
    const q = query.trim().toLowerCase();
    const matches = (s: string) => !q || s.toLowerCase().includes(q);

    const fileItems: ResultItem[] = (files.data ?? [])
      .filter((f: FileItem) => matches(f.name))
      .slice(0, 5)
      .map((f) => ({
        id: f.id,
        kind: "files" as const,
        label: f.name,
        secondary: `${formatBytes(f.size_bytes)} · ${f.mime_type}`,
        icon: <FileIcon mime={f.mime_type} name={f.name} size="sm" />,
        onSelect: () => {
          closePalette();
          const path = f.folder_id
            ? `/dashboard/files/${f.folder_id}`
            : "/dashboard/files";
          router.push(path);
        },
      }));

    const folderItems: ResultItem[] = (folders.data ?? [])
      .filter((f: Folder) => matches(f.name))
      .slice(0, 5)
      .map((f) => ({
        id: f.id,
        kind: "folders" as const,
        label: f.name,
        secondary: `Depth ${f.depth}`,
        icon: <Icon name="Folder" size={14} className="text-text-3" />,
        onSelect: () => {
          closePalette();
          router.push(`/dashboard/files/${f.id}`);
        },
      }));

    const tagItems: ResultItem[] = (tags.data ?? [])
      .filter((t: Tag) => matches(t.name))
      .slice(0, 5)
      .map((t) => ({
        id: t.id,
        kind: "tags" as const,
        label: t.name,
        secondary: t.color || undefined,
        icon: (
          <span
            className="h-3 w-3 rounded-full border border-border"
            style={{ background: t.color || "var(--text-3)" }}
          />
        ),
        onSelect: () => {
          closePalette();
          router.push("/dashboard/tags");
        },
      }));

    const storageItems: ResultItem[] = (storages.data ?? [])
      .filter((s: Storage) => matches(s.display_name) || matches(s.bucket))
      .slice(0, 5)
      .map((s) => ({
        id: s.id,
        kind: "storages" as const,
        label: s.display_name,
        secondary: `${s.provider} · ${s.bucket}`,
        icon: <Icon name="Cloud" size={14} className="text-text-3" />,
        onSelect: () => {
          closePalette();
          router.push("/dashboard/storages");
        },
      }));

    const active: [string, ResultItem[]][] = [];
    if (scope === "all" || scope === "files") active.push(["Files", fileItems]);
    if (scope === "all" || scope === "folders") active.push(["Folders", folderItems]);
    if (scope === "all" || scope === "tags") active.push(["Tags", tagItems]);
    if (scope === "all" || scope === "storages") active.push(["Storages", storageItems]);
    return active.filter(([, items]) => items.length > 0);
  }, [files.data, folders.data, tags.data, storages.data, query, scope, closePalette, router]);

  const flat = useMemo(() => groups.flatMap(([, items]) => items), [groups]);
  const total = flat.length;

  useEffect(() => {
    if (focus >= total) setFocus(Math.max(0, total - 1));
  }, [focus, total]);

  if (!open) return null;

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setFocus((f) => Math.min(total - 1, f + 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setFocus((f) => Math.max(0, f - 1));
    } else if (e.key === "Enter") {
      e.preventDefault();
      flat[focus]?.onSelect();
    }
  };

  let running = 0;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center px-4 pt-[10vh] bg-black/40 backdrop-blur-sm"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) closePalette();
      }}
    >
      <div
        role="dialog"
        aria-modal
        aria-label="Command palette"
        className="w-full max-w-[640px] max-h-[80vh] flex flex-col bg-surface border border-border rounded-lg shadow-lg overflow-hidden"
        onKeyDown={onKeyDown}
      >
        <div className="flex items-center gap-2 px-3 h-11 border-b border-border">
          <Icon name="Search" size={16} className="text-text-3" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setFocus(0);
            }}
            placeholder="Search files, tags, folders, storages…"
            className="flex-1 bg-transparent text-[13px] text-text placeholder:text-text-3 focus:outline-none"
          />
          <Kbd>Esc</Kbd>
        </div>

        <div className="flex items-center gap-1 px-3 py-2 border-b border-border overflow-x-auto">
          {scopes.map((s) => (
            <button
              key={s.value}
              onClick={() => setScope(s.value)}
              className={cn(
                "inline-flex items-center gap-1.5 h-6 rounded-full border px-2 text-[11px] transition-colors",
                scope === s.value
                  ? "bg-accent-soft border-accent-border text-accent-2 font-medium"
                  : "bg-surface border-border text-text-2 hover:text-text",
              )}
            >
              <Icon name={s.icon} size={11} />
              {s.label}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto">
          {total === 0 ? (
            <div className="px-6 py-12 text-center text-[12px] text-text-3">
              {query ? "No results." : "Type to search across your workspace."}
            </div>
          ) : (
            groups.map(([label, items]) => (
              <div key={label} className="py-1">
                <div className="px-3 py-1 text-[10px] font-semibold uppercase tracking-wide text-text-3">
                  {label}
                </div>
                {items.map((it) => {
                  const i = running++;
                  const active = i === focus;
                  return (
                    <button
                      key={`${it.kind}:${it.id}`}
                      onMouseEnter={() => setFocus(i)}
                      onClick={() => it.onSelect()}
                      className={cn(
                        "flex w-full items-center gap-3 px-3 py-1.5 text-left",
                        active ? "bg-surface-2" : "hover:bg-surface-2",
                      )}
                    >
                      <span className="shrink-0">{it.icon}</span>
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-[13px] text-text">
                          <Highlight text={it.label} query={query} />
                        </span>
                        {it.secondary && (
                          <span className="block truncate text-[11px] text-text-3">
                            {it.secondary}
                          </span>
                        )}
                      </span>
                      {active && (
                        <Icon name="CornerDownLeft" size={12} className="text-text-3" />
                      )}
                    </button>
                  );
                })}
              </div>
            ))
          )}
        </div>

        <div className="flex items-center gap-3 px-3 py-2 border-t border-border text-[11px] text-text-3">
          <span className="flex items-center gap-1">
            <Kbd>↑</Kbd>
            <Kbd>↓</Kbd>
            Navigate
          </span>
          <span className="flex items-center gap-1">
            <Kbd>↵</Kbd>
            Open
          </span>
          <span className="flex items-center gap-1">
            <Kbd>Esc</Kbd>
            Close
          </span>
        </div>
      </div>
    </div>
  );
}
