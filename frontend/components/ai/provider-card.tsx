"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AICapability, AIConfig, aiApi } from "@/lib/api/ai";
import { toast } from "@/lib/stores/toast-store";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/v2/badge";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { DropdownMenu } from "@/components/ui/v2/dropdown-menu";
import { Icon } from "@/components/ui/v2/icon";
import { IconButton } from "@/components/ui/v2/icon-button";
import { providerMeta } from "./provider-meta";

function relTime(iso?: string | null) {
  if (!iso) return "never";
  const d = new Date(iso);
  const diff = (Date.now() - d.getTime()) / 1000;
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 86400 * 7) return `${Math.floor(diff / 86400)}d ago`;
  return d.toLocaleDateString();
}

function capabilityLabel(c: AICapability): string {
  if (c === "chat") return "Chat";
  if (c === "embed") return "Embed";
  return "Transcribe";
}

export interface ProviderCardProps {
  config: AIConfig;
}

export function ProviderCard({ config }: ProviderCardProps) {
  const qc = useQueryClient();
  const meta = providerMeta[config.provider];

  const test = useMutation({
    mutationFn: () => aiApi.test(config.id),
    onSuccess: (res) => {
      if (res.ok) {
        toast("Connection OK", "success");
      } else {
        toast(`Test failed: ${res.error ?? "unknown"}`, "error");
      }
      qc.invalidateQueries({ queryKey: ["ai-configs"] });
    },
    onError: (e: unknown) => {
      toast(e instanceof Error ? e.message : "Test failed", "error");
    },
  });

  const remove = useMutation({
    mutationFn: () => aiApi.remove(config.id),
    onSuccess: () => {
      toast(`Removed ${config.display_name}`, "success");
      qc.invalidateQueries({ queryKey: ["ai-configs"] });
    },
    onError: (e: unknown) => {
      toast(e instanceof Error ? e.message : "Delete failed", "error");
    },
  });

  const setDefault = useMutation({
    mutationFn: (cap: AICapability) => aiApi.setDefault(config.id, cap),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["ai-configs"] });
      toast("Default updated", "success");
    },
    onError: (e: unknown) => {
      toast(e instanceof Error ? e.message : "Update failed", "error");
    },
  });

  const hasError = !!config.last_error;
  const defaults: AICapability[] = [];
  if (config.is_default_chat) defaults.push("chat");
  if (config.is_default_embed) defaults.push("embed");
  if (config.is_default_transcribe) defaults.push("transcribe");

  return (
    <Card padding="md" className="flex flex-col gap-3">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-2.5 min-w-0">
          <span
            className="inline-flex h-8 w-8 items-center justify-center rounded text-white shrink-0"
            style={{ background: meta.color }}
          >
            <Icon name="Sparkles" size={14} />
          </span>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <span className="font-semibold text-text truncate text-[13px]">
                {config.display_name}
              </span>
              <Badge color="slate">{meta.label}</Badge>
            </div>
            {config.key_prefix && (
              <div className="text-[11px] text-text-3 mt-0.5 font-mono">
                {config.key_prefix}…
              </div>
            )}
          </div>
        </div>
        <DropdownMenu
          align="end"
          trigger={({ toggle }) => (
            <IconButton size="sm" onClick={toggle} aria-label="Actions">
              <Icon name="MoreHorizontal" size={14} />
            </IconButton>
          )}
          items={[
            {
              label: "Test connection",
              icon: <Icon name="Zap" size={12} />,
              onSelect: () => test.mutate(),
            },
            ...config.capabilities.map((c) => ({
              label: `Set default for ${capabilityLabel(c)}`,
              icon: <Icon name="Star" size={12} />,
              onSelect: () => setDefault.mutate(c),
              disabled:
                (c === "chat" && config.is_default_chat) ||
                (c === "embed" && config.is_default_embed) ||
                (c === "transcribe" && config.is_default_transcribe),
            })),
            { separator: true, label: "" },
            {
              label: "Remove",
              icon: <Icon name="Trash2" size={12} />,
              danger: true,
              onSelect: () => {
                if (confirm(`Remove ${config.display_name}?`)) remove.mutate();
              },
            },
          ]}
        />
      </div>

      <div className="flex flex-wrap gap-1">
        {config.capabilities.map((c) => (
          <Badge
            key={c}
            color={defaults.includes(c) ? "accent" : "slate"}
            dot={defaults.includes(c)}
          >
            {capabilityLabel(c)}
            {defaults.includes(c) && " · default"}
          </Badge>
        ))}
      </div>

      <dl className="grid grid-cols-2 gap-x-3 gap-y-1.5 text-[11px]">
        {config.chat_model && (
          <>
            <dt className="text-text-3">Chat model</dt>
            <dd className="text-text font-mono truncate">{config.chat_model}</dd>
          </>
        )}
        {config.embed_model && (
          <>
            <dt className="text-text-3">Embed model</dt>
            <dd className="text-text font-mono truncate">{config.embed_model}</dd>
          </>
        )}
        {config.transcribe_model && (
          <>
            <dt className="text-text-3">Transcribe</dt>
            <dd className="text-text font-mono truncate">
              {config.transcribe_model}
            </dd>
          </>
        )}
        {config.base_url && (
          <>
            <dt className="text-text-3">Base URL</dt>
            <dd className="text-text font-mono truncate">{config.base_url}</dd>
          </>
        )}
        <dt className="text-text-3">Last tested</dt>
        <dd className="text-text">{relTime(config.last_tested_at)}</dd>
        <dt className="text-text-3">Last used</dt>
        <dd className="text-text">{relTime(config.last_used_at)}</dd>
        <dt className="text-text-3">Tokens</dt>
        <dd className="text-text tabular-nums">
          {config.used_tokens_in.toLocaleString()} in ·{" "}
          {config.used_tokens_out.toLocaleString()} out
        </dd>
      </dl>

      {hasError && (
        <div className="flex gap-2 rounded border border-red-200 bg-red-50 px-2.5 py-1.5 text-[11px] text-red-700">
          <Icon name="AlertCircle" size={12} className="shrink-0 mt-0.5" />
          <span className="break-words">{config.last_error}</span>
        </div>
      )}

      <div className="flex gap-2 border-t border-border pt-2 -mx-1 px-1">
        <Button
          variant="secondary"
          size="sm"
          className={cn("flex-1")}
          disabled={test.isPending}
          onClick={() => test.mutate()}
        >
          <Icon name="Zap" size={12} />
          {test.isPending ? "Testing…" : "Test"}
        </Button>
      </div>
    </Card>
  );
}
