"use client";

import { FormEvent, useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  AICapability,
  AIProviderKind,
  CreateAIConfigInput,
  aiApi,
} from "@/lib/api/ai";
import { ApiError } from "@/lib/api/client";
import { toast } from "@/lib/stores/toast-store";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/v2/badge";
import { Button } from "@/components/ui/v2/button";
import { Checkbox } from "@/components/ui/v2/checkbox";
import { Dialog } from "@/components/ui/v2/dialog";
import { Icon } from "@/components/ui/v2/icon";
import { Input } from "@/components/ui/v2/input";
import { ProviderMeta, providerList, providerMeta } from "./provider-meta";

export interface AddProviderDialogProps {
  open: boolean;
  onClose: () => void;
}

export function AddProviderDialog({ open, onClose }: AddProviderDialogProps) {
  const qc = useQueryClient();
  const [kind, setKind] = useState<AIProviderKind>("anthropic");
  const [displayName, setDisplayName] = useState("");
  const [baseURL, setBaseURL] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [chatModel, setChatModel] = useState("");
  const [embedModel, setEmbedModel] = useState("");
  const [transcribeModel, setTranscribeModel] = useState("");
  const [caps, setCaps] = useState<AICapability[]>([]);
  const [def, setDef] = useState<AICapability[]>([]);
  const [err, setErr] = useState<string | null>(null);

  const meta: ProviderMeta = providerMeta[kind];

  useEffect(() => {
    if (!open) return;
    applyPreset("anthropic");
    setDisplayName("");
    setApiKey("");
    setBaseURL("");
    setErr(null);
  }, [open]);

  function applyPreset(k: AIProviderKind) {
    const m = providerMeta[k];
    setKind(k);
    setChatModel(m.chatDefault);
    setEmbedModel(m.embedDefault ?? "");
    setTranscribeModel(m.transcribeDefault ?? "");
    setCaps([...m.defaultCapabilities]);
    setDef([...m.defaultCapabilities]);
    if (k === "openai-compatible" || k === "ollama") {
      setBaseURL(m.baseUrlHint ?? "");
    } else {
      setBaseURL("");
    }
  }

  const create = useMutation({
    mutationFn: (input: CreateAIConfigInput) => aiApi.create(input),
    onSuccess: (cfg) => {
      qc.invalidateQueries({ queryKey: ["ai-configs"] });
      toast(`Added ${cfg.display_name}`, "success");
      onClose();
    },
    onError: (e: unknown) => {
      setErr(e instanceof ApiError ? e.message : "Create failed");
    },
  });

  function toggleCap(c: AICapability) {
    setCaps((cur) =>
      cur.includes(c) ? cur.filter((x) => x !== c) : [...cur, c],
    );
    setDef((cur) => (cur.includes(c) ? cur : cur));
  }

  function toggleDefault(c: AICapability) {
    setDef((cur) =>
      cur.includes(c) ? cur.filter((x) => x !== c) : [...cur, c],
    );
    setCaps((cur) => (cur.includes(c) ? cur : [...cur, c]));
  }

  async function handle(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    const input: CreateAIConfigInput = {
      display_name: displayName.trim(),
      provider: kind,
      base_url: baseURL.trim() || undefined,
      api_key: apiKey.trim() || undefined,
      chat_model: chatModel.trim() || undefined,
      embed_model: embedModel.trim() || undefined,
      transcribe_model: transcribeModel.trim() || undefined,
      capabilities: caps.length ? caps : undefined,
      default_chat: def.includes("chat"),
      default_embed: def.includes("embed"),
      default_transcribe: def.includes("transcribe"),
    };
    create.mutate(input);
  }

  const keyRequired = kind !== "ollama";
  const baseUrlVisible = kind === "openai-compatible" || kind === "ollama";

  return (
    <Dialog
      open={open}
      onClose={() => (create.isPending ? undefined : onClose())}
      title="Add AI provider"
      description="Bring your own key — we never see or bill it."
      size="lg"
      footer={
        <>
          <Button
            type="button"
            variant="secondary"
            onClick={onClose}
            disabled={create.isPending}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            form="add-ai-provider-form"
            variant="accent"
            disabled={
              create.isPending ||
              !displayName.trim() ||
              (keyRequired && !apiKey.trim()) ||
              caps.length === 0
            }
          >
            {create.isPending ? "Testing…" : "Test & save"}
          </Button>
        </>
      }
    >
      <form id="add-ai-provider-form" onSubmit={handle} className="space-y-4">
        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">
            Provider
          </label>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-1.5">
            {providerList.map((m) => (
              <button
                key={m.kind}
                type="button"
                onClick={() => applyPreset(m.kind)}
                className={cn(
                  "flex items-start gap-2 rounded-lg border px-3 py-2 text-left transition-colors",
                  kind === m.kind
                    ? "bg-accent-soft border-accent-border"
                    : "bg-surface border-border hover:border-border-strong",
                )}
              >
                <span
                  className="inline-flex h-7 w-7 items-center justify-center rounded text-white shrink-0"
                  style={{ background: m.color }}
                >
                  <Icon name="Sparkles" size={12} />
                </span>
                <div className="min-w-0 flex-1">
                  <div className="text-[13px] font-medium text-text truncate">
                    {m.label}
                  </div>
                  <div className="text-[11px] text-text-3 truncate">
                    {m.tagline}
                  </div>
                </div>
                {kind === m.kind && (
                  <Icon name="Check" size={14} className="text-accent-2 mt-0.5" />
                )}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">
            Display name
          </label>
          <Input
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder={`My ${meta.label} key`}
            maxLength={100}
          />
        </div>

        {baseUrlVisible && (
          <div className="space-y-1.5">
            <label className="block text-[12px] font-medium text-text-2">
              Base URL
            </label>
            <Input
              value={baseURL}
              onChange={(e) => setBaseURL(e.target.value)}
              placeholder={meta.baseUrlHint ?? ""}
              className="font-mono text-[12px]"
            />
            <p className="text-[11px] text-text-3">
              Must be HTTPS in production. In dev, localhost/127.0.0.1 are
              allowed.
            </p>
          </div>
        )}

        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">
            API key{!keyRequired && " (optional)"}
          </label>
          <Input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder={meta.keyHint}
            autoComplete="off"
            className="font-mono text-[12px]"
          />
          {meta.docsUrl && (
            <p className="text-[11px] text-text-3">
              Get one from{" "}
              <a
                href={meta.docsUrl}
                target="_blank"
                rel="noreferrer"
                className="text-accent-2 hover:underline"
              >
                {meta.docsUrl}
              </a>
            </p>
          )}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {meta.hasChat && (
            <div className="space-y-1.5">
              <label className="block text-[11px] font-medium text-text-2 uppercase tracking-wide">
                Chat model
              </label>
              <Input
                value={chatModel}
                onChange={(e) => setChatModel(e.target.value)}
                placeholder={meta.chatDefault}
                className="font-mono text-[12px]"
              />
            </div>
          )}
          {meta.hasEmbed && (
            <div className="space-y-1.5">
              <label className="block text-[11px] font-medium text-text-2 uppercase tracking-wide">
                Embed model
              </label>
              <Input
                value={embedModel}
                onChange={(e) => setEmbedModel(e.target.value)}
                placeholder={meta.embedDefault ?? ""}
                className="font-mono text-[12px]"
              />
            </div>
          )}
          {meta.hasTranscribe && (
            <div className="space-y-1.5">
              <label className="block text-[11px] font-medium text-text-2 uppercase tracking-wide">
                Transcribe model
              </label>
              <Input
                value={transcribeModel}
                onChange={(e) => setTranscribeModel(e.target.value)}
                placeholder={meta.transcribeDefault ?? ""}
                className="font-mono text-[12px]"
              />
            </div>
          )}
        </div>

        <div className="space-y-2">
          <label className="block text-[12px] font-medium text-text-2">
            Capabilities
          </label>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-2">
            {(["chat", "embed", "transcribe"] as AICapability[]).map((c) => {
              const supported =
                (c === "chat" && meta.hasChat) ||
                (c === "embed" && meta.hasEmbed) ||
                (c === "transcribe" && meta.hasTranscribe);
              return (
                <label
                  key={c}
                  className={cn(
                    "flex items-start gap-2 rounded border px-2.5 py-2 text-[12px]",
                    !supported && "opacity-50 cursor-not-allowed",
                    caps.includes(c)
                      ? "bg-accent-soft border-accent-border"
                      : "bg-surface border-border",
                  )}
                >
                  <Checkbox
                    checked={caps.includes(c)}
                    disabled={!supported}
                    onChange={() => toggleCap(c)}
                  />
                  <div className="flex-1">
                    <div className="font-medium text-text capitalize">{c}</div>
                    <label className="flex items-center gap-1.5 mt-1 text-[11px] text-text-3">
                      <input
                        type="checkbox"
                        disabled={!supported || !caps.includes(c)}
                        checked={def.includes(c)}
                        onChange={() => toggleDefault(c)}
                      />
                      Set as default
                    </label>
                  </div>
                </label>
              );
            })}
          </div>
          {!meta.hasEmbed && (
            <p className="text-[11px] text-text-3">
              {meta.label} doesn&rsquo;t support embeddings via this adapter —
              use a separate config for embed.
            </p>
          )}
        </div>

        {err && (
          <div className="flex gap-2 rounded border border-red-200 bg-red-50 px-3 py-2 text-[12px] text-red-700">
            <Icon name="AlertCircle" size={14} className="shrink-0 mt-0.5" />
            <span className="break-words">{err}</span>
          </div>
        )}

        <div className="flex items-start gap-2 rounded border border-amber-200 bg-amber-50 px-3 py-2 text-[11px] text-amber-900">
          <Icon name="Lock" size={12} className="shrink-0 mt-0.5" />
          <span>
            API keys are encrypted at rest (AES-GCM). We only ever read them to
            call the provider — never log them, never share them.
          </span>
        </div>

        <div className="flex items-center gap-1.5 text-[11px] text-text-3">
          <Badge color="slate">Wave 0</Badge>
          <span>
            On save we call the provider&rsquo;s test endpoint; it must succeed.
          </span>
        </div>
      </form>
    </Dialog>
  );
}
