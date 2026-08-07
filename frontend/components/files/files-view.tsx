"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { useSelectionStore } from "@/lib/stores/selection-store";
import { filesApi, FileItem, fetchFileBlob } from "@/lib/api/files";
import { foldersApi, FolderView, Folder } from "@/lib/api/folders";
import { storagesApi, Storage } from "@/lib/api/storages";
import { tagsApi, Tag } from "@/lib/api/tags";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { Checkbox } from "@/components/ui/v2/checkbox";
import { ContextMenu } from "@/components/ui/v2/context-menu";
import { Dialog } from "@/components/ui/v2/dialog";
import { DropdownMenu } from "@/components/ui/v2/dropdown-menu";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { FileIcon, kindFromMime } from "@/components/ui/v2/file-icon";
import { Icon } from "@/components/ui/v2/icon";
import { IconButton } from "@/components/ui/v2/icon-button";
import { Input } from "@/components/ui/v2/input";
import { Pill, PillColor } from "@/components/ui/v2/pill";
import { Select } from "@/components/ui/v2/select";
import { Table, THead, THCell, TRow, TCell } from "@/components/ui/v2/table";
import { Tooltip } from "@/components/ui/v2/tooltip";
import { FolderTree } from "@/components/folders/folder-tree";
import { NewFolderDialog } from "@/components/folders/new-folder-dialog";
import { UploadDialog } from "@/components/upload/upload-dialog";
import { ShareDialog } from "@/components/share/share-dialog";

interface Props {
  folderID: string | null;
}

type ViewMode = "list" | "grid";
type SortKey = "name-asc" | "name-desc" | "size-desc" | "size-asc" | "date-desc" | "date-asc";
type TypeFilter = "all" | "pdf" | "image" | "docs" | "other";
type DateFilter = "all" | "today" | "7d" | "30d";

function formatBytes(n: number): string {
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

function useDebounced<T>(value: T, delay = 300) {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(t);
  }, [value, delay]);
  return debounced;
}

const pillColorFor = (hex: string): PillColor => {
  const h = hex.toLowerCase();
  if (h.startsWith("#10") || h.includes("emerald") || h === "#10b981") return "emerald";
  if (h.startsWith("#63") || h.startsWith("#4f") || h === "#6366f1") return "indigo";
  if (h.startsWith("#f5") || h.startsWith("#d9") || h === "#f59e0b") return "amber";
  if (h.startsWith("#f4") || h.startsWith("#e1") || h === "#f43f5e") return "rose";
  if (h.startsWith("#0e") || h === "#0ea5e9") return "sky";
  if (h.startsWith("#8b") || h === "#8b5cf6") return "violet";
  return "slate";
};

function matchesType(f: FileItem, t: TypeFilter) {
  if (t === "all") return true;
  const k = kindFromMime(f.mime_type, f.name);
  if (t === "pdf") return k === "pdf";
  if (t === "image") return k === "image";
  if (t === "docs") return ["word", "excel", "powerpoint", "text"].includes(k);
  return !["pdf", "image", "word", "excel", "powerpoint", "text"].includes(k);
}

function matchesDate(f: FileItem, d: DateFilter) {
  if (d === "all") return true;
  const when = f.uploaded_at ?? f.created_at;
  if (!when) return false;
  const t = new Date(when).getTime();
  const now = Date.now();
  if (d === "today") return now - t < 86_400_000;
  if (d === "7d") return now - t < 7 * 86_400_000;
  return now - t < 30 * 86_400_000;
}

function sortFiles(files: FileItem[], key: SortKey) {
  const arr = [...files];
  const cmp: Record<SortKey, (a: FileItem, b: FileItem) => number> = {
    "name-asc": (a, b) => a.name.localeCompare(b.name),
    "name-desc": (a, b) => b.name.localeCompare(a.name),
    "size-desc": (a, b) => b.size_bytes - a.size_bytes,
    "size-asc": (a, b) => a.size_bytes - b.size_bytes,
    "date-desc": (a, b) =>
      new Date(b.uploaded_at ?? b.created_at).getTime() -
      new Date(a.uploaded_at ?? a.created_at).getTime(),
    "date-asc": (a, b) =>
      new Date(a.uploaded_at ?? a.created_at).getTime() -
      new Date(b.uploaded_at ?? b.created_at).getTime(),
  };
  arr.sort(cmp[key]);
  return arr;
}

async function downloadBlob(file: FileItem) {
  const { blob } = await fetchFileBlob(file.id);
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = file.name;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

export function FilesView({ folderID }: Props) {
  const qc = useQueryClient();
  const searchParams = useSearchParams();
  const smartMode = searchParams?.get("mode") ?? null;
  const starredOnly = smartMode === "starred";
  const trashOnly = smartMode === "trash";
  const { selectedFileIDs, toggle, clear, setMany } = useSelectionStore();

  const [uploadOpen, setUploadOpen] = useState(false);
  const [newFolderOpen, setNewFolderOpen] = useState(false);
  const [shareTarget, setShareTarget] =
    useState<{ type: "file" | "folder"; id: string; name: string } | null>(null);
  const [tagDialogOpen, setTagDialogOpen] = useState(false);

  const [view, setView] = useState<ViewMode>("list");
  const [sortKey, setSortKey] = useState<SortKey>("date-desc");
  const [typeFilter, setTypeFilter] = useState<TypeFilter>("all");
  const [dateFilter, setDateFilter] = useState<DateFilter>("all");
  const [tagFilter, setTagFilter] = useState<string>("");
  const [searchInput, setSearchInput] = useState("");
  const search = useDebounced(searchInput, 300);

  const { data: storages = [] } = useQuery<Storage[]>({
    queryKey: ["storages"],
    queryFn: () => storagesApi.list(),
  });

  const { data: folderView } = useQuery<FolderView | null>({
    queryKey: ["folder-view", folderID],
    queryFn: () => (folderID ? foldersApi.get(folderID) : Promise.resolve(null)),
    enabled: folderID !== null && !starredOnly && !trashOnly,
  });

  const { data: rootChildren = [] } = useQuery<Folder[]>({
    queryKey: ["folders", null],
    queryFn: () => foldersApi.list(null),
    enabled: folderID === null && !starredOnly && !trashOnly,
  });

  const children = folderID ? folderView?.children ?? [] : rootChildren;

  const { data: files = [], isLoading } = useQuery<FileItem[]>({
    queryKey: ["files", folderID, tagFilter, search, smartMode],
    queryFn: () =>
      filesApi.list({
        folder_id: starredOnly || trashOnly ? undefined : folderID ?? undefined,
        tag_id: tagFilter || undefined,
        q: search || undefined,
        starred: starredOnly ? "true" : undefined,
        status: trashOnly ? "deleted" : undefined,
      }),
  });

  const { data: tags = [] } = useQuery<Tag[]>({
    queryKey: ["tags"],
    queryFn: () => tagsApi.list(),
  });

  usePageTitle(
    starredOnly ? "Starred" : trashOnly ? "Trash" : folderView?.folder?.name ?? "Files",
  );

  const filtered = useMemo(() => {
    const base = files.filter((f) => matchesType(f, typeFilter) && matchesDate(f, dateFilter));
    return sortFiles(base, sortKey);
  }, [files, typeFilter, dateFilter, sortKey]);

  const totalSize = filtered.reduce((a, f) => a + f.size_bytes, 0);
  const tagByID = useMemo(() => Object.fromEntries(tags.map((t) => [t.id, t])), [tags]);

  const deleteFile = useMutation({
    mutationFn: (id: string) => filesApi.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["files"] });
      qc.invalidateQueries({ queryKey: ["storages"] });
    },
  });

  const createFolder = useMutation({
    mutationFn: (name: string) => foldersApi.create(name, folderID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["folders"] });
      qc.invalidateQueries({ queryKey: ["folder-view"] });
    },
  });

  const deleteFolder = useMutation({
    mutationFn: (id: string) => foldersApi.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["folders"] });
      qc.invalidateQueries({ queryKey: ["folder-view"] });
    },
  });

  const forceDeleteFolder = useMutation({
    mutationFn: (id: string) => foldersApi.remove(id, true),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["folders"] });
      qc.invalidateQueries({ queryKey: ["folder-view"] });
      qc.invalidateQueries({ queryKey: ["files"] });
    },
  });

  const updateFileTags = useMutation({
    mutationFn: ({ id, tagIDs }: { id: string; tagIDs: string[] }) =>
      filesApi.updateTags(id, tagIDs),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["files"] }),
  });

  const toggleStar = useMutation({
    mutationFn: ({ id, starred }: { id: string; starred: boolean }) =>
      filesApi.toggleStar(id, starred),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["files"] }),
  });

  const restoreFile = useMutation({
    mutationFn: (id: string) => filesApi.restore(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["files"] });
      qc.invalidateQueries({ queryKey: ["storages"] });
    },
  });

  const purgeFile = useMutation({
    mutationFn: (id: string) => filesApi.purge(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["files"] }),
  });

  useEffect(() => {
    clear();
  }, [folderID, clear]);

  const canUpload = storages.length > 0;
  const selectedCount = selectedFileIDs.size;
  const selectedFiles = useMemo(
    () => filtered.filter((f) => selectedFileIDs.has(f.id)),
    [filtered, selectedFileIDs],
  );

  const allSelected = filtered.length > 0 && filtered.every((f) => selectedFileIDs.has(f.id));

  async function onFolderDelete(id: string) {
    try {
      await deleteFolder.mutateAsync(id);
    } catch (e) {
      const err = e as { status?: number; message?: string };
      if (err.status === 409) {
        if (
          confirm(
            "This folder is not empty. Recursive delete will soft-delete all files inside. Continue?",
          )
        ) {
          await forceDeleteFolder.mutateAsync(id);
        }
      } else {
        alert(err.message ?? "Delete failed");
      }
    }
  }

  async function onBulkDelete() {
    if (!confirm(`Delete ${selectedCount} file${selectedCount === 1 ? "" : "s"}?`)) return;
    await Promise.all(selectedFiles.map((f) => filesApi.remove(f.id)));
    qc.invalidateQueries({ queryKey: ["files"] });
    qc.invalidateQueries({ queryKey: ["storages"] });
    clear();
  }

  async function onBulkDownload() {
    for (const f of selectedFiles) {
      try {
        await downloadBlob(f);
      } catch (e) {
        console.error("download failed", f.id, e);
      }
    }
  }

  function onBulkShare() {
    const first = selectedFiles[0];
    if (!first) return;
    setShareTarget({ type: "file", id: first.id, name: first.name });
  }

  function onBulkMove() {
    alert("Move support arrives in Phase E.");
  }

  return (
    <div className="flex flex-col gap-4 md:flex-row md:gap-6 h-full">
      <aside className="md:w-60 md:shrink-0">
        <FolderTree activeID={folderID} />
      </aside>

      <div className="flex-1 min-w-0 space-y-4">
        <Breadcrumbs folderView={folderView} />

        <div className="flex items-start justify-between gap-3 flex-wrap">
          <div className="min-w-0">
            <h1 className="text-lg font-semibold text-text truncate">
              {starredOnly
                ? "Starred"
                : trashOnly
                  ? "Trash"
                  : folderView?.folder?.name ?? "Files"}
            </h1>
            <div className="text-[12px] text-text-3 mt-0.5">
              {filtered.length} file{filtered.length === 1 ? "" : "s"} · {formatBytes(totalSize)}
              {children.length > 0 && ` · ${children.length} folder${children.length === 1 ? "" : "s"}`}
            </div>
          </div>
          {!trashOnly && !starredOnly && (
            <div className="flex items-center gap-2">
              <Button
                variant="secondary"
                size="md"
                onClick={() => setNewFolderOpen(true)}
              >
                <Icon name="FolderPlus" size={14} /> New folder
              </Button>
              <Button
                variant="accent"
                size="md"
                disabled={!canUpload}
                onClick={() => setUploadOpen(true)}
              >
                <Icon name="Upload" size={14} /> Upload
              </Button>
            </div>
          )}
        </div>

        {trashOnly && (
          <div className="flex items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-[12px] text-amber-900">
            <Icon name="AlertTriangle" size={14} className="shrink-0 mt-0.5" />
            <span>
              Items in Trash are permanently deleted after 30 days. Restore what
              you want to keep before then.
            </span>
          </div>
        )}

        <Card padding="sm">
          <div className="flex flex-wrap items-center gap-2">
            <div className="relative min-w-[200px] flex-1">
              <Icon
                name="Search"
                size={14}
                className="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-3"
              />
              <Input
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                placeholder="Search files…"
                className="pl-8"
              />
            </div>

            <Select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value as TypeFilter)}
              className="max-w-[140px]"
            >
              <option value="all">All types</option>
              <option value="pdf">PDF</option>
              <option value="image">Images</option>
              <option value="docs">Docs</option>
              <option value="other">Other</option>
            </Select>

            <Select
              value={tagFilter}
              onChange={(e) => setTagFilter(e.target.value)}
              className="max-w-[160px]"
            >
              <option value="">All tags</option>
              {tags.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </Select>

            <Select
              value={dateFilter}
              onChange={(e) => setDateFilter(e.target.value as DateFilter)}
              className="max-w-[140px]"
            >
              <option value="all">Any date</option>
              <option value="today">Today</option>
              <option value="7d">Last 7 days</option>
              <option value="30d">Last 30 days</option>
            </Select>

            <div className="ml-auto flex items-center gap-1">
              <Tooltip content="List view" side="top">
                <IconButton
                  active={view === "list"}
                  onClick={() => setView("list")}
                  aria-label="List view"
                >
                  <Icon name="List" size={14} />
                </IconButton>
              </Tooltip>
              <Tooltip content="Grid view" side="top">
                <IconButton
                  active={view === "grid"}
                  onClick={() => setView("grid")}
                  aria-label="Grid view"
                >
                  <Icon name="Grid" size={14} />
                </IconButton>
              </Tooltip>

              <Select
                value={sortKey}
                onChange={(e) => setSortKey(e.target.value as SortKey)}
                className="w-[160px]"
              >
                <option value="date-desc">Newest first</option>
                <option value="date-asc">Oldest first</option>
                <option value="name-asc">Name (A–Z)</option>
                <option value="name-desc">Name (Z–A)</option>
                <option value="size-desc">Largest first</option>
                <option value="size-asc">Smallest first</option>
              </Select>
            </div>
          </div>
        </Card>

        {selectedCount > 0 && (
          <div className="flex items-center gap-2 rounded-lg border border-accent-border bg-accent-soft px-3 py-2">
            <span className="text-[12px] font-medium text-accent-2">
              {selectedCount} selected
            </span>
            <span className="text-text-3">·</span>
            <Button size="sm" variant="ghost" onClick={onBulkDownload}>
              <Icon name="Download" size={12} /> Download
            </Button>
            <Button size="sm" variant="ghost" onClick={onBulkShare}>
              <Icon name="Share2" size={12} /> Share
            </Button>
            <Button size="sm" variant="ghost" onClick={() => setTagDialogOpen(true)}>
              <Icon name="Tag" size={12} /> Tag
            </Button>
            <Button size="sm" variant="ghost" onClick={onBulkMove}>
              <Icon name="Move" size={12} /> Move
            </Button>
            <Button size="sm" variant="ghost" onClick={onBulkDelete} className="text-danger">
              <Icon name="Trash2" size={12} /> Delete
            </Button>
            <span className="ml-auto">
              <Button size="sm" variant="ghost" onClick={clear}>
                Clear
              </Button>
            </span>
          </div>
        )}

        {children.length > 0 && (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
            {children.map((f) => (
              <ContextMenu
                key={f.id}
                items={[
                  {
                    label: "Share",
                    icon: <Icon name="Share2" size={12} />,
                    onSelect: () =>
                      setShareTarget({ type: "folder", id: f.id, name: f.name }),
                  },
                  {
                    label: "Delete",
                    icon: <Icon name="Trash2" size={12} />,
                    onSelect: () => onFolderDelete(f.id),
                    danger: true,
                  },
                ]}
              >
                <Link
                  href={`/dashboard/files/${f.id}`}
                  className="group flex items-center gap-2 rounded-lg border border-border bg-surface px-3 py-2 hover:border-border-strong"
                >
                  <Icon name="Folder" size={16} className="text-text-3" />
                  <span className="truncate text-[13px] text-text flex-1">{f.name}</span>
                  <Icon name="ChevronRight" size={12} className="text-text-3" />
                </Link>
              </ContextMenu>
            ))}
          </div>
        )}

        {isLoading ? (
          <div className="text-[12px] text-text-3 py-8 text-center">Loading files…</div>
        ) : filtered.length === 0 ? (
          children.length === 0 ? (
            <Card>
              <EmptyState
                icon={<Icon name="Files" size={18} />}
                title={canUpload ? "Empty folder" : "No storage connected"}
                description={
                  canUpload
                    ? "Upload a file or create a sub-folder to get started."
                    : "Add a storage first to upload files."
                }
                action={
                  canUpload ? (
                    <Button variant="accent" onClick={() => setUploadOpen(true)}>
                      <Icon name="Upload" size={14} /> Upload
                    </Button>
                  ) : (
                    <Link
                      href="/dashboard/storages"
                      className="inline-flex items-center gap-1.5 h-8 px-3 rounded bg-accent text-white text-[13px] font-medium hover:bg-accent-2"
                    >
                      <Icon name="Plus" size={14} /> Add storage
                    </Link>
                  )
                }
              />
            </Card>
          ) : null
        ) : view === "list" ? (
          <FilesListView
            files={filtered}
            tagByID={tagByID}
            trashMode={trashOnly}
            allSelected={allSelected}
            onToggleAll={() => {
              if (allSelected) clear();
              else setMany(filtered.map((f) => f.id));
            }}
            onToggle={toggle}
            isSelected={(id) => selectedFileIDs.has(id)}
            onDelete={(id) => deleteFile.mutate(id)}
            onShare={(f) => setShareTarget({ type: "file", id: f.id, name: f.name })}
            onToggleStar={(f) =>
              toggleStar.mutate({ id: f.id, starred: !f.starred })
            }
            onRestore={(id) => restoreFile.mutate(id)}
            onPurge={(id) => {
              if (confirm("Permanently delete this file? This cannot be undone.")) {
                purgeFile.mutate(id);
              }
            }}
          />
        ) : (
          <FilesGridView
            files={filtered}
            tagByID={tagByID}
            trashMode={trashOnly}
            onToggle={toggle}
            isSelected={(id) => selectedFileIDs.has(id)}
            onShare={(f) => setShareTarget({ type: "file", id: f.id, name: f.name })}
            onDelete={(id) => deleteFile.mutate(id)}
            onToggleStar={(f) =>
              toggleStar.mutate({ id: f.id, starred: !f.starred })
            }
            onRestore={(id) => restoreFile.mutate(id)}
            onPurge={(id) => {
              if (confirm("Permanently delete this file? This cannot be undone.")) {
                purgeFile.mutate(id);
              }
            }}
          />
        )}
      </div>

      <UploadDialog
        open={uploadOpen}
        onClose={() => setUploadOpen(false)}
        storages={storages}
        folderID={folderID}
        onUploaded={() => {
          qc.invalidateQueries({ queryKey: ["files"] });
          qc.invalidateQueries({ queryKey: ["storages"] });
        }}
      />
      <NewFolderDialog
        open={newFolderOpen}
        onClose={() => setNewFolderOpen(false)}
        onCreate={(name) => createFolder.mutateAsync(name).then(() => undefined)}
      />
      {shareTarget && (
        <ShareDialog
          open={true}
          onClose={() => setShareTarget(null)}
          targetType={shareTarget.type}
          targetID={shareTarget.id}
          targetName={shareTarget.name}
        />
      )}
      <BulkTagDialog
        open={tagDialogOpen}
        onClose={() => setTagDialogOpen(false)}
        tags={tags}
        onApply={async (tagIDs) => {
          await Promise.all(
            selectedFiles.map((f) =>
              updateFileTags.mutateAsync({ id: f.id, tagIDs }),
            ),
          );
          setTagDialogOpen(false);
          clear();
        }}
      />
    </div>
  );
}

function Breadcrumbs({ folderView }: { folderView: FolderView | null | undefined }) {
  const crumbs = folderView?.breadcrumbs ?? [];
  if (crumbs.length === 0 && !folderView) {
    return null;
  }
  return (
    <nav className="flex items-center gap-1 text-[12px] text-text-3 flex-wrap">
      <Link href="/dashboard/files" className="hover:text-text">
        Root
      </Link>
      {crumbs.map((b, i) => (
        <span key={b.id} className="flex items-center gap-1">
          <Icon name="ChevronRight" size={10} />
          {i === crumbs.length - 1 ? (
            <span className="text-text font-medium">{b.name}</span>
          ) : (
            <Link href={`/dashboard/files/${b.id}`} className="hover:text-text">
              {b.name}
            </Link>
          )}
        </span>
      ))}
    </nav>
  );
}

interface ListProps {
  files: FileItem[];
  tagByID: Record<string, Tag>;
  trashMode: boolean;
  allSelected: boolean;
  onToggleAll: () => void;
  onToggle: (id: string) => void;
  isSelected: (id: string) => boolean;
  onDelete: (id: string) => void;
  onShare: (f: FileItem) => void;
  onToggleStar: (f: FileItem) => void;
  onRestore: (id: string) => void;
  onPurge: (id: string) => void;
}

function FilesListView({
  files,
  tagByID,
  trashMode,
  allSelected,
  onToggleAll,
  onToggle,
  isSelected,
  onDelete,
  onShare,
  onToggleStar,
  onRestore,
  onPurge,
}: ListProps) {
  return (
    <Card padding="none" className="overflow-hidden">
      <div className="overflow-x-auto">
        <Table>
          <THead>
            <tr>
              <THCell className="w-8">
                <Checkbox
                  checked={allSelected}
                  onChange={onToggleAll}
                  aria-label="Select all"
                />
              </THCell>
              <THCell>Name</THCell>
              <THCell className="w-24">Size</THCell>
              <THCell className="w-48">Tags</THCell>
              <THCell className="w-32">Type</THCell>
              <THCell className="w-32">Uploaded</THCell>
              <THCell className="w-10" />
            </tr>
          </THead>
          <tbody>
            {files.map((f) => {
              const selected = isSelected(f.id);
              const attached = f.tag_ids
                .map((id) => tagByID[id])
                .filter(Boolean);
              return (
                <TRow key={f.id} selected={selected}>
                  <TCell>
                    <Checkbox
                      checked={selected}
                      onChange={() => onToggle(f.id)}
                      aria-label={`Select ${f.name}`}
                    />
                  </TCell>
                  <TCell>
                    <div className="flex items-center gap-2 min-w-0">
                      <button
                        type="button"
                        onClick={(e) => {
                          e.preventDefault();
                          onToggleStar(f);
                        }}
                        aria-label={f.starred ? "Unstar" : "Star"}
                        className={
                          f.starred
                            ? "text-warn hover:opacity-80"
                            : "text-text-3 hover:text-text"
                        }
                      >
                        <Icon
                          name="Star"
                          size={14}
                          className={f.starred ? "fill-current" : undefined}
                        />
                      </button>
                      <Link
                        href={`/dashboard/files/view/${f.id}`}
                        className="flex items-center gap-2 min-w-0 group/name"
                      >
                        <FileIcon mime={f.mime_type} name={f.name} size="sm" />
                        <span className="truncate font-medium text-text group-hover/name:underline">
                          {f.name}
                        </span>
                        {f.ai_status === "ready" && (
                          <Icon
                            name="Sparkles"
                            size={11}
                            className="shrink-0 text-accent-2"
                            aria-label="AI-processed"
                          />
                        )}
                      </Link>
                    </div>
                  </TCell>
                  <TCell className="tabular-nums text-text-2 whitespace-nowrap">
                    {formatBytes(f.size_bytes)}
                  </TCell>
                  <TCell>
                    {attached.length === 0 ? (
                      <span className="text-text-3 text-[11px]">—</span>
                    ) : (
                      <div className="flex flex-wrap items-center gap-1">
                        {attached.slice(0, 3).map((t) => (
                          <Pill
                            key={t.id}
                            name={t.name}
                            color={pillColorFor(t.color)}
                          />
                        ))}
                        {attached.length > 3 && (
                          <span className="text-[10px] text-text-3">
                            +{attached.length - 3}
                          </span>
                        )}
                      </div>
                    )}
                  </TCell>
                  <TCell className="text-text-2 whitespace-nowrap">
                    {f.mime_type}
                  </TCell>
                  <TCell className="text-text-3 whitespace-nowrap">
                    {relTime(f.uploaded_at ?? f.created_at)}
                  </TCell>
                  <TCell className="text-right">
                    <DropdownMenu
                      align="end"
                      trigger={({ toggle }) => (
                        <IconButton
                          size="sm"
                          onClick={toggle}
                          aria-label="Actions"
                        >
                          <Icon name="MoreHorizontal" size={14} />
                        </IconButton>
                      )}
                      items={
                        trashMode
                          ? [
                              {
                                label: "Restore",
                                icon: <Icon name="RotateCcw" size={12} />,
                                onSelect: () => onRestore(f.id),
                              },
                              { separator: true, label: "" },
                              {
                                label: "Delete forever",
                                icon: <Icon name="Trash2" size={12} />,
                                onSelect: () => onPurge(f.id),
                                danger: true,
                              },
                            ]
                          : [
                              {
                                label: "View",
                                icon: <Icon name="Eye" size={12} />,
                                onSelect: () => {
                                  window.location.href = `/dashboard/files/view/${f.id}`;
                                },
                              },
                              {
                                label: "Share",
                                icon: <Icon name="Share2" size={12} />,
                                onSelect: () => onShare(f),
                              },
                              {
                                label: "Download",
                                icon: <Icon name="Download" size={12} />,
                                onSelect: () => {
                                  downloadBlob(f).catch((e) => alert(e.message));
                                },
                              },
                              { separator: true, label: "" },
                              {
                                label: "Delete",
                                icon: <Icon name="Trash2" size={12} />,
                                onSelect: () => onDelete(f.id),
                                danger: true,
                              },
                            ]
                      }
                    />
                  </TCell>
                </TRow>
              );
            })}
          </tbody>
        </Table>
      </div>
    </Card>
  );
}

interface GridProps {
  files: FileItem[];
  tagByID: Record<string, Tag>;
  trashMode: boolean;
  onToggle: (id: string) => void;
  isSelected: (id: string) => boolean;
  onShare: (f: FileItem) => void;
  onDelete: (id: string) => void;
  onToggleStar: (f: FileItem) => void;
  onRestore: (id: string) => void;
  onPurge: (id: string) => void;
}

function FilesGridView({
  files,
  tagByID,
  trashMode,
  onToggle,
  isSelected,
  onShare,
  onDelete,
  onToggleStar,
  onRestore,
  onPurge,
}: GridProps) {
  return (
    <div
      className="grid gap-3"
      style={{ gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))" }}
    >
      {files.map((f) => {
        const selected = isSelected(f.id);
        const attached = f.tag_ids.map((id) => tagByID[id]).filter(Boolean);
        return (
          <div
            key={f.id}
            className={cn(
              "group relative rounded-lg border bg-surface p-3 transition-colors",
              selected ? "border-accent bg-accent-soft" : "border-border hover:border-border-strong",
            )}
          >
            <div className="absolute left-2 top-2 z-10">
              <Checkbox
                checked={selected}
                onChange={() => onToggle(f.id)}
                aria-label={`Select ${f.name}`}
              />
            </div>
            <div className="absolute right-2 top-2 z-10 flex items-center gap-1">
              {!trashMode && (
                <IconButton
                  size="sm"
                  onClick={() => onToggleStar(f)}
                  aria-label={f.starred ? "Unstar" : "Star"}
                  className={f.starred ? "text-warn" : undefined}
                >
                  <Icon
                    name="Star"
                    size={12}
                    className={f.starred ? "fill-current" : undefined}
                  />
                </IconButton>
              )}
              <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                {trashMode ? (
                  <>
                    <IconButton
                      size="sm"
                      onClick={() => onRestore(f.id)}
                      aria-label="Restore"
                    >
                      <Icon name="RotateCcw" size={12} />
                    </IconButton>
                    <IconButton
                      size="sm"
                      onClick={() => onPurge(f.id)}
                      aria-label="Delete forever"
                    >
                      <Icon name="Trash2" size={12} />
                    </IconButton>
                  </>
                ) : (
                  <>
                    <IconButton size="sm" onClick={() => onShare(f)} aria-label="Share">
                      <Icon name="Share2" size={12} />
                    </IconButton>
                    <IconButton size="sm" onClick={() => onDelete(f.id)} aria-label="Delete">
                      <Icon name="Trash2" size={12} />
                    </IconButton>
                  </>
                )}
              </div>
            </div>
            <Link
              href={`/dashboard/files/view/${f.id}`}
              className="flex h-28 items-center justify-center rounded bg-surface-2 mb-2"
            >
              <FileIcon mime={f.mime_type} name={f.name} size="lg" />
            </Link>
            <Link
              href={`/dashboard/files/view/${f.id}`}
              className="flex items-center gap-1.5 text-[13px] font-medium text-text truncate hover:underline"
            >
              <span className="truncate">{f.name}</span>
              {f.ai_status === "ready" && (
                <Icon
                  name="Sparkles"
                  size={11}
                  className="shrink-0 text-accent-2"
                  aria-label="AI-processed"
                />
              )}
            </Link>
            <div className="flex items-center justify-between text-[11px] text-text-3 mt-1">
              <span>{formatBytes(f.size_bytes)}</span>
              <span>{relTime(f.uploaded_at ?? f.created_at)}</span>
            </div>
            {attached.length > 0 && (
              <div className="flex flex-wrap items-center gap-1 mt-2">
                {attached.slice(0, 2).map((t) => (
                  <Pill key={t.id} name={t.name} color={pillColorFor(t.color)} />
                ))}
                {attached.length > 2 && (
                  <span className="text-[10px] text-text-3">
                    +{attached.length - 2}
                  </span>
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

interface BulkTagDialogProps {
  open: boolean;
  onClose: () => void;
  tags: Tag[];
  onApply: (tagIDs: string[]) => Promise<void> | void;
}

function BulkTagDialog({ open, onClose, tags, onApply }: BulkTagDialogProps) {
  const [selected, setSelected] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (open) setSelected([]);
  }, [open]);

  function toggle(id: string) {
    setSelected((cur) => (cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id]));
  }

  async function apply() {
    setLoading(true);
    try {
      await onApply(selected);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Apply tags"
      description="Selected tags will be set on every selected file."
      footer={
        <>
          <Button type="button" variant="secondary" onClick={onClose} disabled={loading}>
            Cancel
          </Button>
          <Button
            type="button"
            variant="accent"
            onClick={apply}
            disabled={loading || selected.length === 0}
          >
            Apply
          </Button>
        </>
      }
    >
      {tags.length === 0 ? (
        <div className="text-[12px] text-text-3">
          No tags yet. Create tags from the Tags page.
        </div>
      ) : (
        <div className="flex flex-wrap gap-1.5">
          {tags.map((t) => (
            <Pill
              key={t.id}
              name={t.name}
              color={pillColorFor(t.color)}
              active={selected.includes(t.id)}
              onClick={() => toggle(t.id)}
            />
          ))}
        </div>
      )}
    </Dialog>
  );
}

