"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { filesApi, fetchFileBlob, FileItem, CustomField } from "@/lib/api/files";
import { storagesApi, Storage } from "@/lib/api/storages";
import { foldersApi, Folder } from "@/lib/api/folders";
import { tagsApi, Tag } from "@/lib/api/tags";
import { activityApi, ActivityEntry } from "@/lib/api/activity";
import { ActivityFeed } from "@/components/activity/activity-feed";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { FileIcon } from "@/components/ui/v2/file-icon";
import { Icon } from "@/components/ui/v2/icon";
import { IconButton } from "@/components/ui/v2/icon-button";
import { Input } from "@/components/ui/v2/input";
import { Pill, PillColor } from "@/components/ui/v2/pill";
import { Select } from "@/components/ui/v2/select";
import { Skeleton } from "@/components/ui/v2/skeleton";
import { Tabs } from "@/components/ui/v2/tabs";
import { ImageViewer } from "@/components/viewer/image-viewer";
import { ShareDialog } from "@/components/share/share-dialog";
import { AiChatPanel } from "./ai-chat-panel";

const PDFViewer = dynamic(
  () => import("@/components/viewer/pdf-viewer").then((m) => m.PDFViewer),
  {
    ssr: false,
    loading: () => (
      <div className="p-6 text-[12px] text-text-3">Loading PDF viewer…</div>
    ),
  },
);

type TabKey = "preview" | "details" | "versions" | "activity" | "comments";

export interface FileDetailProps {
  fileID: string;
}

function formatBytes(n: number): string {
  if (!n) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(n) / Math.log(1024));
  return `${(n / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
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

export function FileDetail({ fileID }: FileDetailProps) {
  const router = useRouter();
  const qc = useQueryClient();
  const [tab, setTab] = useState<TabKey>("preview");
  const [shareOpen, setShareOpen] = useState(false);

  const { data: file, isLoading, error } = useQuery<FileItem>({
    queryKey: ["file", fileID],
    queryFn: () => filesApi.get(fileID),
  });

  const starMut = useMutation({
    mutationFn: ({ id, starred }: { id: string; starred: boolean }) =>
      filesApi.toggleStar(id, starred),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["file", fileID] });
      qc.invalidateQueries({ queryKey: ["files"] });
    },
  });

  usePageTitle(file?.name ?? "File");

  return (
    <div className="h-full">
      <div className="flex items-center gap-2 mb-3">
        <IconButton size="sm" onClick={() => router.back()} aria-label="Back">
          <Icon name="ArrowLeft" size={14} />
        </IconButton>
        <Link
          href="/dashboard/files"
          className="text-[12px] text-text-3 hover:text-text"
        >
          Files
        </Link>
        <Icon name="ChevronRight" size={10} className="text-text-3" />
        <span className="text-[12px] text-text font-medium truncate">
          {file?.name ?? "…"}
        </span>
      </div>

      {isLoading && (
        <div className="space-y-3">
          <Skeleton className="h-9 w-1/2" />
          <Skeleton className="h-64 w-full" />
        </div>
      )}
      {error && (
        <Card>
          <div className="text-[12px] text-danger">
            {(error as Error).message || "Failed to load file"}
          </div>
        </Card>
      )}

      {file && (
        <div className="grid gap-4 lg:grid-cols-10">
          <div className="lg:col-span-7 min-w-0 space-y-4">
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-start gap-3 min-w-0">
                <FileIcon mime={file.mime_type} name={file.name} size="lg" />
                <div className="min-w-0">
                  <h1 className="text-lg font-semibold text-text truncate">
                    {file.name}
                  </h1>
                  <div className="text-[12px] text-text-3 mt-0.5">
                    {file.mime_type} · {formatBytes(file.size_bytes)}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <Button
                  variant="secondary"
                  onClick={() =>
                    starMut.mutate({ id: file.id, starred: !file.starred })
                  }
                >
                  <Icon
                    name="Star"
                    size={14}
                    className={file.starred ? "text-warn fill-current" : undefined}
                  />
                  {file.starred ? "Starred" : "Star"}
                </Button>
                <Button variant="secondary" onClick={() => setShareOpen(true)}>
                  <Icon name="Share2" size={14} /> Share
                </Button>
              </div>
            </div>

            <Tabs
              value={tab}
              onChange={(v) => setTab(v as TabKey)}
              items={[
                { value: "preview", label: "Preview", icon: <Icon name="Eye" size={12} /> },
                { value: "details", label: "Details", icon: <Icon name="Info" size={12} /> },
                { value: "versions", label: "Versions", icon: <Icon name="History" size={12} /> },
                { value: "activity", label: "Activity", icon: <Icon name="Activity" size={12} /> },
                { value: "comments", label: "Comments", icon: <Icon name="MessageSquare" size={12} /> },
              ]}
            />

            {tab === "preview" && <PreviewTab file={file} />}
            {tab === "details" && (
              <DetailsTab
                file={file}
                onSaved={() => qc.invalidateQueries({ queryKey: ["file", fileID] })}
              />
            )}
            {tab === "versions" && (
              <Card>
                <EmptyState
                  icon={<Icon name="History" size={18} />}
                  title="Version history coming soon"
                  description="Future versions will let you revert to earlier uploads of this file."
                />
              </Card>
            )}
            {tab === "activity" && <ActivityTab fileID={fileID} />}
            {tab === "comments" && (
              <Card>
                <EmptyState
                  icon={<Icon name="MessageSquare" size={18} />}
                  title="Comments will appear here once enabled"
                  description="Threaded comments are on the roadmap."
                />
              </Card>
            )}
          </div>

          <div className="lg:col-span-3 min-w-0">
            <AiChatPanel file={file} />
          </div>
        </div>
      )}

      {file && (
        <ShareDialog
          open={shareOpen}
          onClose={() => setShareOpen(false)}
          targetType="file"
          targetID={file.id}
          targetName={file.name}
        />
      )}
    </div>
  );
}

function PreviewTab({ file }: { file: FileItem }) {
  const [blob, setBlob] = useState<Blob | null>(null);
  const [imageURL, setImageURL] = useState<string | null>(null);
  const [textBody, setTextBody] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    let tmpURL: string | null = null;
    setLoading(true);
    setErr(null);
    setBlob(null);
    setImageURL(null);
    setTextBody(null);
    (async () => {
      try {
        const { blob, mime } = await fetchFileBlob(file.id);
        if (cancelled) return;
        setBlob(blob);
        if (mime.startsWith("image/") || file.mime_type.startsWith("image/")) {
          tmpURL = URL.createObjectURL(blob);
          setImageURL(tmpURL);
        }
        if (mime.startsWith("text/") || file.mime_type.startsWith("text/")) {
          setTextBody(await blob.text());
        }
      } catch (e) {
        if (!cancelled) setErr(e instanceof Error ? e.message : "Failed to load");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
      if (tmpURL) URL.revokeObjectURL(tmpURL);
    };
  }, [file.id, file.mime_type]);

  return (
    <Card padding="none">
      <div className="border-b border-border bg-accent-soft/60 px-4 py-2 text-[11px] text-accent-2 flex items-center gap-2">
        <Icon name="Lock" size={12} />
        View-only mode. The file stays on your bucket; no download link is exposed.
      </div>
      <div className="p-4">
        {loading && <div className="text-[12px] text-text-3">Loading preview…</div>}
        {err && <div className="text-[12px] text-danger">{err}</div>}
        {!loading && !err && (
          <PreviewRenderer
            mime={file.mime_type}
            name={file.name}
            blob={blob}
            imageURL={imageURL}
            textBody={textBody}
          />
        )}
      </div>
    </Card>
  );
}

function PreviewRenderer({
  mime,
  name,
  blob,
  imageURL,
  textBody,
}: {
  mime: string;
  name: string;
  blob: Blob | null;
  imageURL: string | null;
  textBody: string | null;
}) {
  if (mime === "application/pdf" && blob) return <PDFViewer file={blob} />;
  if (mime.startsWith("image/") && imageURL)
    return <ImageViewer url={imageURL} alt={name} />;
  if (mime.startsWith("text/") && textBody !== null)
    return (
      <pre
        className="whitespace-pre-wrap rounded border border-border bg-surface-2 p-4 text-[12px] font-mono max-h-[70vh] overflow-auto select-none"
        onContextMenu={(e) => e.preventDefault()}
      >
        {textBody}
      </pre>
    );
  return (
    <EmptyState
      icon={<Icon name="FileX" size={18} />}
      title="Preview not available"
      description={`Only PDF, images, and plain text can be previewed (mime: ${mime}).`}
    />
  );
}

function DetailsTab({
  file,
  onSaved,
}: {
  file: FileItem;
  onSaved: () => void;
}) {
  const { data: storages = [] } = useQuery<Storage[]>({
    queryKey: ["storages"],
    queryFn: () => storagesApi.list(),
  });
  const { data: folders = [] } = useQuery<Folder[]>({
    queryKey: ["folders-flat"],
    queryFn: () => foldersApi.list(),
  });
  const { data: tags = [] } = useQuery<Tag[]>({
    queryKey: ["tags"],
    queryFn: () => tagsApi.list(),
  });

  const [name, setName] = useState(file.name);
  const [tagIDs, setTagIDs] = useState<string[]>(file.tag_ids);
  const [fields, setFields] = useState<CustomField[]>(file.custom_fields);

  useEffect(() => {
    setName(file.name);
    setTagIDs(file.tag_ids);
    setFields(file.custom_fields);
  }, [file.id, file.name, file.tag_ids, file.custom_fields]);

  const storage = storages.find((s) => s.id === file.storage_id);
  const folder = folders.find((f) => f.id === file.folder_id);
  const dirty = useMemo(() => {
    if (name.trim() !== file.name) return true;
    if (JSON.stringify([...tagIDs].sort()) !== JSON.stringify([...file.tag_ids].sort()))
      return true;
    if (JSON.stringify(fields) !== JSON.stringify(file.custom_fields)) return true;
    return false;
  }, [name, tagIDs, fields, file]);

  const save = useMutation({
    mutationFn: () =>
      filesApi.update(file.id, {
        name: name.trim() !== file.name ? name.trim() : undefined,
        tag_ids: tagIDs,
        custom_fields: fields,
      }),
    onSuccess: () => onSaved(),
  });

  return (
    <div className="space-y-4">
      <Card title="Metadata">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-3">
          <Field label="Name">
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <Field label="Type">
            <span className="text-[13px] text-text font-mono">{file.mime_type}</span>
          </Field>
          <Field label="Size">
            <span className="text-[13px] text-text tabular-nums">
              {formatBytes(file.size_bytes)}
            </span>
          </Field>
          <Field label="Status">
            <span className="text-[13px] capitalize text-text">{file.status}</span>
          </Field>
          <Field label="Folder">
            {folder ? (
              <Link
                href={`/dashboard/files/${folder.id}`}
                className="text-[13px] text-accent-2 hover:underline"
              >
                {folder.name}
              </Link>
            ) : (
              <span className="text-[13px] text-text-3">Root</span>
            )}
          </Field>
          <Field label="Storage">
            {storage ? (
              <Link
                href="/dashboard/storages"
                className="text-[13px] text-accent-2 hover:underline"
              >
                {storage.display_name}
              </Link>
            ) : (
              <span className="text-[13px] text-text-3">—</span>
            )}
          </Field>
          <Field label="Uploaded">
            <span className="text-[13px] text-text">
              {file.uploaded_at
                ? new Date(file.uploaded_at).toLocaleString()
                : "—"}
            </span>
          </Field>
          <Field label="Created">
            <span className="text-[13px] text-text">
              {new Date(file.created_at).toLocaleString()}
            </span>
          </Field>
          <Field label="File ID">
            <span className="text-[12px] font-mono text-text-2 break-all">{file.id}</span>
          </Field>
        </div>
      </Card>

      <Card
        title="Custom fields"
        action={
          <Button
            size="sm"
            variant="secondary"
            onClick={() =>
              setFields((cur) => [...cur, { key: "", value: "", type: "text" }])
            }
          >
            <Icon name="Plus" size={12} /> Add field
          </Button>
        }
      >
        {fields.length === 0 ? (
          <div className="text-[12px] text-text-3">No custom fields.</div>
        ) : (
          <div className="space-y-2">
            {fields.map((f, i) => (
              <div key={i} className="flex items-start gap-2">
                <Input
                  placeholder="key"
                  value={f.key}
                  onChange={(e) =>
                    setFields((cur) =>
                      cur.map((x, j) => (i === j ? { ...x, key: e.target.value } : x)),
                    )
                  }
                  className="max-w-[180px]"
                />
                <Input
                  placeholder="value"
                  value={f.value}
                  onChange={(e) =>
                    setFields((cur) =>
                      cur.map((x, j) => (i === j ? { ...x, value: e.target.value } : x)),
                    )
                  }
                  className="flex-1"
                />
                <Select
                  value={f.type}
                  onChange={(e) =>
                    setFields((cur) =>
                      cur.map((x, j) =>
                        i === j
                          ? { ...x, type: e.target.value as CustomField["type"] }
                          : x,
                      ),
                    )
                  }
                  className="max-w-[120px]"
                >
                  <option value="text">text</option>
                  <option value="number">number</option>
                  <option value="date">date</option>
                  <option value="boolean">boolean</option>
                </Select>
                <IconButton
                  size="md"
                  aria-label="Remove field"
                  onClick={() =>
                    setFields((cur) => cur.filter((_, j) => j !== i))
                  }
                >
                  <Icon name="Trash2" size={12} />
                </IconButton>
              </div>
            ))}
          </div>
        )}
      </Card>

      <Card title="Tags">
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
                active={tagIDs.includes(t.id)}
                onClick={() =>
                  setTagIDs((cur) =>
                    cur.includes(t.id)
                      ? cur.filter((id) => id !== t.id)
                      : [...cur, t.id],
                  )
                }
              />
            ))}
          </div>
        )}
      </Card>

      <div
        className={cn(
          "sticky bottom-0 flex items-center justify-end gap-2 border-t border-border bg-surface/95 backdrop-blur -mx-1 px-1 py-3",
          !dirty && "opacity-0 pointer-events-none",
        )}
      >
        <span className="text-[11px] text-text-3 mr-auto">
          {dirty ? "Unsaved changes" : "All changes saved"}
        </span>
        <Button
          variant="ghost"
          onClick={() => {
            setName(file.name);
            setTagIDs(file.tag_ids);
            setFields(file.custom_fields);
          }}
          disabled={!dirty || save.isPending}
        >
          Discard
        </Button>
        <Button
          variant="accent"
          onClick={() => save.mutate()}
          disabled={!dirty || save.isPending}
        >
          {save.isPending ? "Saving…" : "Save"}
        </Button>
      </div>
      {save.isError && (
        <div className="text-[12px] text-danger">
          {(save.error as Error).message || "Save failed"}
        </div>
      )}
    </div>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="min-w-0">
      <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3 mb-1">
        {label}
      </div>
      <div className="min-w-0">{children}</div>
    </div>
  );
}

function ActivityTab({ fileID }: { fileID: string }) {
  const { data = [], isLoading } = useQuery<ActivityEntry[]>({
    queryKey: ["activity", "file", fileID],
    queryFn: () => activityApi.listForSubject(fileID, 50),
  });
  return (
    <Card title="Activity">
      {isLoading ? (
        <div className="text-[12px] text-text-3">Loading…</div>
      ) : (
        <ActivityFeed
          entries={data}
          emptyTitle="No activity recorded yet"
          emptyDescription="Uploads, edits, shares, and restores will appear here."
        />
      )}
    </Card>
  );
}
