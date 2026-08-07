"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import {
  aiApi,
  AIConfig,
  AIExtraction,
  AskResponse,
  SuggestedField,
  SuggestedTag,
  SuggestFolderResponse,
} from "@/lib/api/ai";
import { ApiError } from "@/lib/api/client";
import { filesApi, FileItem, CustomField } from "@/lib/api/files";
import { tagsApi, Tag } from "@/lib/api/tags";
import { toast } from "@/lib/stores/toast-store";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/v2/badge";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { Icon } from "@/components/ui/v2/icon";
import { Pill, PillColor } from "@/components/ui/v2/pill";
import { Textarea } from "@/components/ui/v2/textarea";
import Link from "next/link";
import { AiProcessModal } from "./ai-process-modal";

export interface AiChatPanelProps {
  file: FileItem;
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

export function AiChatPanel({ file }: AiChatPanelProps) {
  const qc = useQueryClient();
  const { data: configs = [] } = useQuery<AIConfig[]>({
    queryKey: ["ai-configs"],
    queryFn: () => aiApi.list(),
    staleTime: 60_000,
  });
  const chatProvider = configs.find((c) => c.is_default_chat) ??
    configs.find((c) => c.capabilities.includes("chat"));

  const { data: allTags = [] } = useQuery<Tag[]>({
    queryKey: ["tags"],
    queryFn: () => tagsApi.list(),
  });

  const [fields, setFields] = useState<SuggestedField[]>([]);
  const [tags, setTags] = useState<SuggestedTag[]>([]);
  const [name, setName] = useState<string>("");
  const [folder, setFolder] = useState<SuggestFolderResponse | null>(null);
  const [meta, setMeta] = useState<string>("");
  const [err, setErr] = useState<string>("");
  const [processOpen, setProcessOpen] = useState<boolean>(false);
  const [question, setQuestion] = useState<string>("");
  const [chatLog, setChatLog] = useState<
    { question: string; answer: AskResponse }[]
  >([]);

  const mime = (file.mime_type || "").toLowerCase();
  const visionSupported =
    mime === "application/pdf" || mime.startsWith("image/");

  const { data: extraction, refetch: refetchExtraction } =
    useQuery<AIExtraction>({
      queryKey: ["file-ai-extraction", file.id],
      queryFn: () => aiApi.getExtraction(file.id),
      staleTime: 10_000,
    });
  const extStatus = extraction?.status ?? file.ai_status ?? "none";

  const fieldsMut = useMutation({
    mutationFn: () => aiApi.suggestFields(file.id),
    onSuccess: (res) => {
      setFields(res.fields);
      const extStatus = res.meta.extraction_status;
      if (extStatus === "unsupported") {
        setMeta("Document type can't be read for extraction yet.");
      } else if (extStatus === "error") {
        setMeta("Extraction failed — see server logs.");
      } else {
        setMeta(formatMeta(res.meta));
      }
      setErr("");
    },
    onError: (e: unknown) => {
      setErr(e instanceof ApiError ? e.message : "Suggestion failed");
    },
  });

  const tagsMut = useMutation({
    mutationFn: () => aiApi.suggestTags(file.id),
    onSuccess: (res) => {
      setTags(res.tags);
      setMeta(formatMeta(res.meta));
      setErr("");
    },
    onError: (e: unknown) => {
      setErr(e instanceof ApiError ? e.message : "Suggestion failed");
    },
  });

  const nameMut = useMutation({
    mutationFn: () => aiApi.suggestName(file.id),
    onSuccess: (res) => {
      setName(res.name);
      setMeta(formatMeta(res.meta));
      setErr("");
    },
    onError: (e: unknown) => {
      setErr(e instanceof ApiError ? e.message : "Suggestion failed");
    },
  });

  const folderMut = useMutation({
    mutationFn: () => aiApi.suggestFolder(file.id),
    onSuccess: (res) => {
      setFolder(res);
      setErr("");
    },
    onError: (e: unknown) => {
      setErr(e instanceof ApiError ? e.message : "Folder suggestion failed");
    },
  });

  const chatMut = useMutation({
    mutationFn: (q: string) => aiApi.fileChat(file.id, q),
    onSuccess: (res, q) => {
      setChatLog((prev) => [...prev, { question: q, answer: res }]);
      setQuestion("");
      setErr("");
    },
    onError: (e: unknown) => {
      setErr(e instanceof ApiError ? e.message : "Chat failed");
    },
  });

  const applyFolder = useMutation({
    mutationFn: async (folderID: string) => {
      await filesApi.update(file.id, { folder_id: folderID });
    },
    onSuccess: () => {
      toast("File moved", "success");
      qc.invalidateQueries({ queryKey: ["file", file.id] });
      qc.invalidateQueries({ queryKey: ["files"] });
      setFolder(null);
    },
    onError: (e: unknown) => {
      toast(e instanceof ApiError ? e.message : "Move failed", "error");
    },
  });

  const applyField = useMutation({
    mutationFn: async (f: SuggestedField) => {
      const next: CustomField[] = [...file.custom_fields];
      const existing = next.findIndex((x) => x.key === f.key);
      const entry: CustomField = {
        key: f.key,
        value: f.value,
        type: f.type,
      };
      if (existing >= 0) next[existing] = entry;
      else next.push(entry);
      await filesApi.update(file.id, { custom_fields: next });
    },
    onSuccess: (_r, f) => {
      toast(`Added field ${f.key}`, "success");
      qc.invalidateQueries({ queryKey: ["file", file.id] });
    },
    onError: (e: unknown) => {
      toast(e instanceof ApiError ? e.message : "Apply failed", "error");
    },
  });

  const applyAllFields = useMutation({
    mutationFn: async () => {
      const existing = new Map(file.custom_fields.map((cf) => [cf.key, cf]));
      for (const f of fields) {
        existing.set(f.key, { key: f.key, value: f.value, type: f.type });
      }
      await filesApi.update(file.id, { custom_fields: Array.from(existing.values()) });
    },
    onSuccess: () => {
      toast(`Added ${fields.length} fields`, "success");
      qc.invalidateQueries({ queryKey: ["file", file.id] });
    },
  });

  const applyTag = useMutation({
    mutationFn: async (t: SuggestedTag) => {
      const existing = allTags.find(
        (x) => x.name.toLowerCase() === t.name.toLowerCase(),
      );
      let tagID: string;
      if (existing) {
        tagID = existing.id;
      } else {
        const created = await tagsApi.create(t.name);
        tagID = created.id;
      }
      const nextIDs = file.tag_ids.includes(tagID)
        ? file.tag_ids
        : [...file.tag_ids, tagID];
      await filesApi.updateTags(file.id, nextIDs);
    },
    onSuccess: (_r, t) => {
      toast(`Added tag ${t.name}`, "success");
      qc.invalidateQueries({ queryKey: ["file", file.id] });
      qc.invalidateQueries({ queryKey: ["tags"] });
    },
    onError: (e: unknown) => {
      toast(e instanceof ApiError ? e.message : "Apply failed", "error");
    },
  });

  const applyName = useMutation({
    mutationFn: async (newName: string) => {
      await filesApi.update(file.id, { name: newName });
    },
    onSuccess: () => {
      toast("Renamed", "success");
      qc.invalidateQueries({ queryKey: ["file", file.id] });
      qc.invalidateQueries({ queryKey: ["files"] });
    },
    onError: (e: unknown) => {
      toast(e instanceof ApiError ? e.message : "Rename failed", "error");
    },
  });

  if (configs.length === 0) {
    return (
      <Card
        title={
          <span className="flex items-center gap-2">
            <Icon name="Sparkles" size={14} className="text-accent-2" />
            AI assistant
          </span>
        }
        className="flex flex-col"
      >
        <div className="py-4 text-center space-y-3">
          <div className="text-[13px] text-text font-medium">
            No AI provider connected
          </div>
          <div className="text-[11px] text-text-3">
            Bring your own Anthropic, OpenAI, or compatible key to enable
            auto-fill, tag suggestions, and rename assist.
          </div>
          <Link
            href="/dashboard/ai"
            className="inline-flex items-center gap-1.5 h-8 px-3 rounded bg-accent text-white text-[13px] font-medium hover:bg-accent-2"
          >
            <Icon name="Plus" size={12} /> Add provider
          </Link>
        </div>
      </Card>
    );
  }

  if (!chatProvider) {
    return (
      <Card
        title={
          <span className="flex items-center gap-2">
            <Icon name="Sparkles" size={14} className="text-accent-2" />
            AI assistant
          </span>
        }
      >
        <div className="py-4 text-center text-[12px] text-text-3">
          No provider with chat capability. Enable one in{" "}
          <Link href="/dashboard/ai" className="text-accent-2 hover:underline">
            AI providers
          </Link>
          .
        </div>
      </Card>
    );
  }

  const working =
    fieldsMut.isPending ||
    tagsMut.isPending ||
    nameMut.isPending ||
    folderMut.isPending ||
    chatMut.isPending ||
    applyField.isPending ||
    applyAllFields.isPending ||
    applyTag.isPending ||
    applyName.isPending ||
    applyFolder.isPending;

  return (
    <Card
      padding="none"
      title={
        <span className="flex items-center gap-2">
          <Icon name="Sparkles" size={14} className="text-accent-2" />
          AI assistant
        </span>
      }
      action={
        <Badge color="slate">
          {chatProvider.provider} · {chatProvider.chat_model || "default"}
        </Badge>
      }
      className="flex flex-col"
    >
      <AiProcessModal
        open={processOpen}
        fileID={file.id}
        fileName={file.name}
        onClose={() => setProcessOpen(false)}
        onDone={() => {
          refetchExtraction();
          qc.invalidateQueries({ queryKey: ["file", file.id] });
        }}
      />
      <div className="p-3 space-y-3">
        {visionSupported && extStatus !== "ready" ? (
          <ExtractionGate
            status={extStatus}
            error={extraction?.error}
            onProcess={() => setProcessOpen(true)}
          />
        ) : (
          <div className="grid grid-cols-2 gap-1.5">
            <Button
              size="sm"
              variant="secondary"
              onClick={() => fieldsMut.mutate()}
              disabled={working}
            >
              <Icon name="Info" size={12} />
              Fields
            </Button>
            <Button
              size="sm"
              variant="secondary"
              onClick={() => tagsMut.mutate()}
              disabled={working}
            >
              <Icon name="Tag" size={12} />
              Tags
            </Button>
            <Button
              size="sm"
              variant="secondary"
              onClick={() => nameMut.mutate()}
              disabled={working}
            >
              <Icon name="Pencil" size={12} />
              Rename
            </Button>
            <Button
              size="sm"
              variant="secondary"
              onClick={() => folderMut.mutate()}
              disabled={working}
            >
              <Icon name="Folder" size={12} />
              Folder
            </Button>
          </div>
        )}

        {visionSupported && extStatus === "ready" && (
          <div className="flex items-center justify-between text-[10px] text-text-3">
            <span className="flex items-center gap-1">
              <Icon name="Check" size={10} className="text-emerald-500" />
              AI-processed · {extraction?.model ?? "cached"}
            </span>
            <button
              type="button"
              className="text-accent-2 hover:underline"
              onClick={() => setProcessOpen(true)}
            >
              Qayta tahlil
            </button>
          </div>
        )}

        {err && (
          <div className="flex gap-2 rounded border border-red-200 bg-red-50 px-2.5 py-1.5 text-[11px] text-red-700">
            <Icon name="AlertCircle" size={12} className="shrink-0 mt-0.5" />
            <span className="break-words">{err}</span>
          </div>
        )}

        {fields.length > 0 && (
          <section className="space-y-1.5">
            <div className="flex items-center justify-between">
              <h4 className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
                Suggested fields
              </h4>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => applyAllFields.mutate()}
                disabled={working}
              >
                Accept all
              </Button>
            </div>
            <ul className="space-y-1.5">
              {fields.map((f) => (
                <li
                  key={f.key}
                  className="rounded border border-dashed border-accent-border bg-accent-soft/30 px-2 py-1.5"
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="min-w-0">
                      <div className="text-[11px] text-text-3 font-mono">
                        {f.key}{" "}
                        <span className="text-[10px]">({f.type})</span>
                      </div>
                      <div className="text-[12px] text-text truncate">
                        {f.value || <em className="text-text-3">empty</em>}
                      </div>
                    </div>
                    <Button
                      size="sm"
                      variant="accent"
                      onClick={() => applyField.mutate(f)}
                      disabled={working}
                    >
                      <Icon name="Check" size={12} />
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        )}

        {tags.length > 0 && (
          <section className="space-y-1.5">
            <h4 className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              Suggested tags
            </h4>
            <div className="flex flex-wrap gap-1.5">
              {tags.map((t) => {
                const existing = allTags.find(
                  (x) => x.name.toLowerCase() === t.name.toLowerCase(),
                );
                const attached =
                  existing && file.tag_ids.includes(existing.id);
                return (
                  <button
                    key={t.name}
                    type="button"
                    onClick={() => applyTag.mutate(t)}
                    disabled={working || attached}
                    className={cn(
                      "group relative inline-flex",
                      attached && "opacity-60",
                    )}
                  >
                    <Pill
                      name={t.name}
                      color={
                        existing ? pillColorFor(existing.color) : "slate"
                      }
                      active={attached}
                    />
                  </button>
                );
              })}
            </div>
          </section>
        )}

        {name && (
          <section className="space-y-1.5">
            <h4 className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              Suggested name
            </h4>
            <div className="flex items-center gap-2 rounded border border-dashed border-accent-border bg-accent-soft/30 px-2 py-1.5">
              <span className="flex-1 text-[12px] text-text truncate">
                {name}
              </span>
              <Button
                size="sm"
                variant="accent"
                onClick={() => applyName.mutate(name)}
                disabled={working}
              >
                Apply
              </Button>
            </div>
          </section>
        )}

        {folder && (
          <section className="space-y-1.5">
            <h4 className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              Suggested folder
            </h4>
            {folder.folder_id ? (
              <div className="flex items-center gap-2 rounded border border-dashed border-accent-border bg-accent-soft/30 px-2 py-1.5">
                <Icon name="Folder" size={12} className="text-accent-2" />
                <span className="flex-1 text-[12px] text-text truncate">
                  {folder.folder_name}
                </span>
                <Button
                  size="sm"
                  variant="accent"
                  onClick={() => applyFolder.mutate(folder.folder_id!)}
                  disabled={working}
                >
                  Move here
                </Button>
              </div>
            ) : (
              <div className="text-[11px] text-text-3">
                No confident match. The file probably doesn&rsquo;t fit an
                existing folder.
              </div>
            )}
          </section>
        )}

        {meta && (
          <div className="text-[10px] text-text-3">{meta}</div>
        )}

        {extStatus === "ready" && (
          <section className="space-y-1.5 border-t border-border pt-3">
            <h4 className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
              Ask this document
            </h4>
            {chatLog.length > 0 && (
              <div className="space-y-2 max-h-60 overflow-y-auto pr-1">
                {chatLog.map((entry, i) => (
                  <div key={i} className="space-y-1">
                    <div className="text-[11px] text-text-2 font-medium">
                      {entry.question}
                    </div>
                    <div className="text-[12px] text-text whitespace-pre-wrap rounded bg-accent-soft/30 border border-accent-border px-2 py-1.5">
                      {entry.answer.answer}
                    </div>
                  </div>
                ))}
              </div>
            )}
            <form
              onSubmit={(e) => {
                e.preventDefault();
                const q = question.trim();
                if (!q || chatMut.isPending) return;
                chatMut.mutate(q);
              }}
              className="space-y-1.5"
            >
              <Textarea
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                placeholder="Savol bering (masalan: expiration date qachon?)"
                rows={2}
                className="text-[12px]"
                onKeyDown={(e) => {
                  if (
                    e.key === "Enter" &&
                    (e.metaKey || e.ctrlKey) &&
                    !chatMut.isPending
                  ) {
                    const q = question.trim();
                    if (q) chatMut.mutate(q);
                  }
                }}
              />
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-text-3">
                  Faqat shu fayl mazmuni asosida javob beradi
                </span>
                <Button
                  type="submit"
                  size="sm"
                  variant="accent"
                  disabled={!question.trim() || chatMut.isPending}
                >
                  {chatMut.isPending ? "…" : "Send"}
                  <Icon name="ArrowRight" size={12} />
                </Button>
              </div>
            </form>
          </section>
        )}

        {!fields.length && !tags.length && !name && !folder && !err &&
          chatLog.length === 0 && extStatus === "ready" && (
            <div className="rounded-lg bg-accent-soft/40 border border-accent-border px-3 py-2.5 text-[11px] text-text-2">
              <div className="font-medium text-text mb-1">
                Ready when you are.
              </div>
              Ask me to extract fields, suggest tags, or pick a better name.
              Your file never leaves your account.
            </div>
          )}
      </div>
    </Card>
  );
}

function ExtractionGate({
  status,
  error,
  onProcess,
}: {
  status: string;
  error?: string;
  onProcess: () => void;
}) {
  if (status === "pending") {
    return (
      <div className="rounded-lg border border-accent-border bg-accent-soft/40 px-3 py-4 text-center space-y-2">
        <div className="text-[12px] text-text-2 flex items-center justify-center gap-2">
          <Icon name="Loader" size={12} className="animate-spin" />
          Tahlil qilinmoqda…
        </div>
        <div className="text-[11px] text-text-3">
          AI hujjatni o&rsquo;qiyapti. Odatda 10-30 soniya.
        </div>
      </div>
    );
  }
  if (status === "failed") {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-3 space-y-2">
        <div className="text-[12px] text-red-700 flex items-center gap-1.5">
          <Icon name="AlertCircle" size={12} />
          Tahlil muvaffaqiyatsiz
        </div>
        {error && (
          <div className="text-[11px] text-red-700 break-words">{error}</div>
        )}
        <Button
          size="sm"
          variant="accent"
          className="w-full"
          onClick={onProcess}
        >
          Qayta urinib ko&rsquo;rish
        </Button>
      </div>
    );
  }
  return (
    <div className="rounded-lg border border-accent-border bg-accent-soft/40 px-3 py-3 space-y-2">
      <div className="text-[12px] text-text">
        Hujjat hali AI bilan tahlil qilinmagan
      </div>
      <div className="text-[11px] text-text-3">
        Fields / Tags / Rename / Folder tugmalarini ishlatish uchun avval
        hujjatni AI ga yuborib tahlil qildiring.
      </div>
      <Button
        size="sm"
        variant="accent"
        className="w-full"
        onClick={onProcess}
      >
        <Icon name="Sparkles" size={12} />
        Hujjatni tahlil qilish
      </Button>
    </div>
  );
}

function formatMeta(m: {
  provider?: string;
  model?: string;
  tokens_in?: number;
  tokens_out?: number;
  latency_ms?: number;
  extraction_status: string;
}): string {
  const parts: string[] = [];
  if (m.provider) parts.push(m.provider);
  if (m.model) parts.push(m.model);
  if (m.tokens_in || m.tokens_out) {
    parts.push(`${m.tokens_in ?? 0}→${m.tokens_out ?? 0} tok`);
  }
  if (m.latency_ms) parts.push(`${(m.latency_ms / 1000).toFixed(1)}s`);
  if (m.extraction_status && m.extraction_status !== "ready") {
    parts.push(`extraction: ${m.extraction_status}`);
  }
  return parts.join(" · ");
}
