"use client";

import { DragEvent, FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Storage } from "@/lib/api/storages";
import { tagsApi, Tag } from "@/lib/api/tags";
import { uploadFile, UploadProgress } from "@/lib/upload/multipart-uploader";
import { AiProcessModal } from "@/components/files/ai-process-modal";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/v2/button";
import { Dialog } from "@/components/ui/v2/dialog";
import { FileIcon } from "@/components/ui/v2/file-icon";
import { Icon } from "@/components/ui/v2/icon";
import { IconButton } from "@/components/ui/v2/icon-button";
import { Input } from "@/components/ui/v2/input";
import { Pill, PillColor } from "@/components/ui/v2/pill";
import { Progress } from "@/components/ui/v2/progress";
import { Select } from "@/components/ui/v2/select";

interface Props {
  open: boolean;
  onClose: () => void;
  storages: Storage[];
  folderID?: string | null;
  onUploaded: () => void;
}

interface CustomField {
  key: string;
  value: string;
  type: "text" | "number" | "date" | "boolean";
}

const providerColors: Record<Storage["provider"], string> = {
  r2: "#f59e0b",
  s3: "#6366f1",
  minio: "#10b981",
};

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

function formatBytes(n: number): string {
  if (!n) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(n) / Math.log(1024));
  return `${(n / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

export function UploadDialog({
  open,
  onClose,
  storages,
  folderID,
  onUploaded,
}: Props) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [storageID, setStorageID] = useState(storages[0]?.id ?? "");
  const [tagIDs, setTagIDs] = useState<string[]>([]);
  const [fields, setFields] = useState<CustomField[]>([]);
  const [progress, setProgress] = useState<UploadProgress | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const [processTarget, setProcessTarget] = useState<{
    fileID: string;
    name: string;
  } | null>(null);

  const { data: tags = [] } = useQuery<Tag[]>({
    queryKey: ["tags"],
    queryFn: () => tagsApi.list(),
    enabled: open,
  });

  useEffect(() => {
    if (!open) return;
    if (!storageID && storages[0]) setStorageID(storages[0].id);
  }, [open, storages, storageID]);

  const activeStorage = useMemo(
    () => storages.find((s) => s.id === storageID) ?? null,
    [storages, storageID],
  );

  function reset() {
    if (fileRef.current) fileRef.current.value = "";
    setFile(null);
    setFields([]);
    setTagIDs([]);
    setProgress(null);
    setErr(null);
  }

  function handleClose() {
    if (loading) return;
    reset();
    onClose();
  }

  function onDrop(e: DragEvent<HTMLLabelElement>) {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    const dropped = e.dataTransfer.files?.[0];
    if (dropped) setFile(dropped);
  }

  async function handle(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    if (!file) {
      setErr("Choose a file");
      return;
    }
    if (!storageID) {
      setErr("Choose a storage");
      return;
    }
    setLoading(true);
    try {
      const result = await uploadFile(
        file,
        {
          storage_id: storageID,
          folder_id: folderID ?? null,
          custom_fields: fields.filter((f) => f.key),
          tag_ids: tagIDs,
        },
        (p) => setProgress(p),
      );
      onUploaded();
      // Offer vision-AI extraction for supported formats (PDFs and images).
      // Skip the modal for unsupported types so we don't dangle a modal that
      // can't do anything.
      const mime = (file.type || "").toLowerCase();
      const supportsVision =
        mime === "application/pdf" || mime.startsWith("image/");
      if (supportsVision) {
        setProcessTarget({ fileID: result.fileID, name: file.name });
      }
      reset();
      onClose();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Upload failed");
    } finally {
      setLoading(false);
    }
  }

  function toggleTag(id: string) {
    setTagIDs((cur) =>
      cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id],
    );
  }

  return (
    <>
    <AiProcessModal
      open={processTarget !== null}
      fileID={processTarget?.fileID ?? null}
      fileName={processTarget?.name}
      onClose={() => setProcessTarget(null)}
      onDone={() => onUploaded()}
    />
    <Dialog
      open={open}
      onClose={handleClose}
      title="Upload file"
      description={
        folderID ? "File will be placed in the current folder." : undefined
      }
      size="md"
      footer={
        <>
          <Button
            type="button"
            variant="secondary"
            onClick={handleClose}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            form="upload-form"
            variant="accent"
            disabled={loading || !storageID || !file}
          >
            {loading ? "Uploading…" : "Upload"}
          </Button>
        </>
      }
    >
      <form id="upload-form" onSubmit={handle} className="space-y-4">
        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">
            Storage
          </label>
          {storages.length === 0 ? (
            <div className="rounded border border-border bg-surface-2 p-3 text-[12px] text-text-3">
              No storage connected.
            </div>
          ) : (
            <div className="flex flex-wrap gap-1.5">
              {storages.map((s) => {
                const active = s.id === storageID;
                return (
                  <button
                    key={s.id}
                    type="button"
                    onClick={() => setStorageID(s.id)}
                    className={cn(
                      "inline-flex items-center gap-2 rounded border px-2.5 h-8 text-[12px] transition-colors",
                      active
                        ? "bg-accent-soft border-accent-border text-accent-2"
                        : "bg-surface border-border text-text-2 hover:border-border-strong",
                    )}
                  >
                    <span
                      className="inline-flex h-5 w-5 items-center justify-center rounded text-white"
                      style={{ background: providerColors[s.provider] }}
                    >
                      <Icon name="Cloud" size={11} />
                    </span>
                    <span className="truncate max-w-[160px]">
                      {s.display_name}
                    </span>
                    <span className="text-text-3 text-[11px] hidden sm:inline">
                      {s.bucket}
                    </span>
                  </button>
                );
              })}
            </div>
          )}
          {activeStorage && (
            <div className="text-[11px] text-text-3">
              {activeStorage.provider.toUpperCase()} · {activeStorage.bucket}
            </div>
          )}
        </div>

        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">
            File
          </label>
          <label
            onDragEnter={(e) => {
              e.preventDefault();
              setDragActive(true);
            }}
            onDragOver={(e) => {
              e.preventDefault();
              setDragActive(true);
            }}
            onDragLeave={() => setDragActive(false)}
            onDrop={onDrop}
            className={cn(
              "relative block rounded-lg border-2 border-dashed px-4 py-6 cursor-pointer transition-colors",
              dragActive
                ? "border-accent bg-accent-soft"
                : file
                  ? "border-accent-border bg-accent-soft/60"
                  : "border-border bg-surface-2 hover:border-accent/50 hover:bg-accent-soft/30",
            )}
          >
            <input
              ref={fileRef}
              type="file"
              className="sr-only"
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            />
            {file ? (
              <div className="flex items-center gap-3">
                <FileIcon mime={file.type} name={file.name} size="lg" />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-[13px] font-medium text-text">
                    {file.name}
                  </div>
                  <div className="text-[11px] text-text-3 mt-0.5">
                    {formatBytes(file.size)} · {file.type || "unknown"}
                  </div>
                </div>
                <IconButton
                  size="sm"
                  aria-label="Remove file"
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setFile(null);
                    if (fileRef.current) fileRef.current.value = "";
                  }}
                >
                  <Icon name="X" size={12} />
                </IconButton>
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center gap-2 text-center">
                <Icon name="Upload" size={22} className="text-text-3" />
                <div className="text-[13px] text-text">
                  Drop a file here or{" "}
                  <span className="text-accent-2 font-medium">browse</span>
                </div>
                <div className="text-[11px] text-text-3">
                  Any file type — size limits depend on your storage.
                </div>
              </div>
            )}
          </label>
        </div>

        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">
            Tags
          </label>
          {tags.length === 0 ? (
            <div className="text-[11px] text-text-3">
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
                  onClick={() => toggleTag(t.id)}
                />
              ))}
            </div>
          )}
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <label className="text-[12px] font-medium text-text-2">
              Custom fields
            </label>
            <button
              type="button"
              className="text-[11px] text-accent-2 hover:underline"
              onClick={() =>
                setFields((cur) => [...cur, { key: "", value: "", type: "text" }])
              }
            >
              + Add field
            </button>
          </div>
          {fields.length === 0 ? (
            <div className="text-[11px] text-text-3">No custom fields.</div>
          ) : (
            <div className="space-y-1.5">
              {fields.map((f, i) => (
                <div key={i} className="flex items-center gap-1.5">
                  <Input
                    placeholder="key"
                    value={f.key}
                    onChange={(e) =>
                      setFields((cur) =>
                        cur.map((x, j) =>
                          j === i ? { ...x, key: e.target.value } : x,
                        ),
                      )
                    }
                    className="h-7 max-w-[140px] text-[12px]"
                  />
                  <Input
                    placeholder="value"
                    value={f.value}
                    onChange={(e) =>
                      setFields((cur) =>
                        cur.map((x, j) =>
                          j === i ? { ...x, value: e.target.value } : x,
                        ),
                      )
                    }
                    className="h-7 flex-1 text-[12px]"
                  />
                  <Select
                    value={f.type}
                    onChange={(e) =>
                      setFields((cur) =>
                        cur.map((x, j) =>
                          j === i
                            ? {
                                ...x,
                                type: e.target.value as CustomField["type"],
                              }
                            : x,
                        ),
                      )
                    }
                    className="h-7 max-w-[100px] text-[12px]"
                  >
                    <option value="text">text</option>
                    <option value="number">number</option>
                    <option value="date">date</option>
                    <option value="boolean">boolean</option>
                  </Select>
                  <IconButton
                    size="sm"
                    aria-label="Remove field"
                    onClick={() =>
                      setFields((cur) => cur.filter((_, j) => j !== i))
                    }
                  >
                    <Icon name="X" size={12} />
                  </IconButton>
                </div>
              ))}
            </div>
          )}
        </div>

        {progress && (
          <div className="space-y-1">
            <Progress value={progress.percent} />
            <div className="flex items-center justify-between text-[11px] text-text-3 tabular-nums">
              <span>{progress.percent.toFixed(1)}%</span>
              <span>
                {formatBytes(progress.loaded)} / {formatBytes(progress.total)}
              </span>
            </div>
          </div>
        )}

        {err && <div className="text-[12px] text-danger">{err}</div>}
      </form>
    </Dialog>
    </>
  );
}
