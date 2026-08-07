"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { AIConfig, aiApi } from "@/lib/api/ai";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { Icon } from "@/components/ui/v2/icon";
import { SectionHead } from "@/components/ui/v2/section-head";
import { ProviderCard } from "@/components/ai/provider-card";
import { AddProviderDialog } from "@/components/ai/add-provider-dialog";
import { providerList } from "@/components/ai/provider-meta";

export default function AIProvidersPage() {
  usePageTitle("AI providers", "Bring your own keys — we don't bill");
  const [addOpen, setAddOpen] = useState(false);

  const { data: configs = [], isLoading } = useQuery<AIConfig[]>({
    queryKey: ["ai-configs"],
    queryFn: () => aiApi.list(),
  });

  const hasChatDefault = configs.some((c) => c.is_default_chat);
  const hasEmbedDefault = configs.some((c) => c.is_default_embed);
  const hasTranscribeDefault = configs.some((c) => c.is_default_transcribe);

  return (
    <div className="space-y-6">
      <SectionHead
        level="h1"
        title="AI providers"
        subtitle="Connect your own keys. Features run against your account, and we never see the prompts."
        action={
          <Button variant="accent" onClick={() => setAddOpen(true)}>
            <Icon name="Plus" size={14} /> Add provider
          </Button>
        }
      />

      {!isLoading && configs.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          <CapabilityCheck label="Chat" ok={hasChatDefault} icon="MessageSquare" />
          <CapabilityCheck label="Embed" ok={hasEmbedDefault} icon="Grid" />
          <CapabilityCheck
            label="Transcribe"
            ok={hasTranscribeDefault}
            icon="Mic"
          />
        </div>
      )}

      {isLoading ? (
        <Card>
          <div className="text-[12px] text-text-3">Loading…</div>
        </Card>
      ) : configs.length === 0 ? (
        <Card>
          <EmptyState
            icon={<Icon name="Sparkles" size={18} />}
            title="No AI providers yet"
            description="Add a provider to unlock auto-fill, smart tagging, semantic search, and more. We never bill — your key, your account."
            action={
              <Button variant="accent" onClick={() => setAddOpen(true)}>
                <Icon name="Plus" size={14} /> Add provider
              </Button>
            }
          />
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          {configs.map((c) => (
            <ProviderCard key={c.id} config={c} />
          ))}
        </div>
      )}

      <Card title="Supported providers" subtitle="Anything that speaks one of these APIs works.">
        <ul className="grid grid-cols-1 md:grid-cols-2 gap-2">
          {providerList.map((m) => (
            <li
              key={m.kind}
              className="flex items-start gap-2 rounded border border-border bg-surface-2 px-2.5 py-2"
            >
              <span
                className="inline-flex h-7 w-7 items-center justify-center rounded text-white shrink-0"
                style={{ background: m.color }}
              >
                <Icon name="Sparkles" size={12} />
              </span>
              <div className="min-w-0">
                <div className="text-[13px] font-medium text-text">
                  {m.label}
                </div>
                <div className="text-[11px] text-text-3">{m.tagline}</div>
              </div>
            </li>
          ))}
        </ul>
      </Card>

      <AddProviderDialog open={addOpen} onClose={() => setAddOpen(false)} />
    </div>
  );
}

function CapabilityCheck({
  label,
  ok,
  icon,
}: {
  label: string;
  ok: boolean;
  icon: string;
}) {
  return (
    <div
      className={
        ok
          ? "flex items-center gap-2 rounded-lg border border-accent-border bg-accent-soft px-3 py-2 text-[12px] text-accent-2"
          : "flex items-center gap-2 rounded-lg border border-border bg-surface-2 px-3 py-2 text-[12px] text-text-3"
      }
    >
      <Icon name={icon} size={14} />
      <span className="flex-1">{label}</span>
      <Icon name={ok ? "Check" : "X"} size={12} />
    </div>
  );
}
