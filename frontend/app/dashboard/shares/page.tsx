"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Share, sharesApi } from "@/lib/api/shares";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { Badge } from "@/components/ui/v2/badge";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { Icon } from "@/components/ui/v2/icon";
import { IconButton } from "@/components/ui/v2/icon-button";
import { SectionHead } from "@/components/ui/v2/section-head";
import { Tabs } from "@/components/ui/v2/tabs";
import { Tooltip } from "@/components/ui/v2/tooltip";
import { Table, THead, THCell, TRow, TCell } from "@/components/ui/v2/table";

type StatusKind = "active" | "expired" | "revoked" | "consumed";

function computeStatus(s: Share): StatusKind {
  if (s.revoked) return "revoked";
  if (s.consumed) return "consumed";
  if (s.expires_at && new Date(s.expires_at) < new Date()) return "expired";
  return "active";
}

type FilterTab = "all" | "active" | "expired" | "revoked";

const statusMeta: Record<StatusKind, { label: string; color: "accent" | "danger" | "warn" | "slate" }> = {
  active: { label: "Active", color: "accent" },
  consumed: { label: "Consumed", color: "slate" },
  expired: { label: "Expired", color: "warn" },
  revoked: { label: "Revoked", color: "danger" },
};

function csvEscape(v: string): string {
  if (/[",\n]/.test(v)) return `"${v.replace(/"/g, '""')}"`;
  return v;
}

function buildCSV(shares: Share[]): string {
  const header = [
    "id",
    "target_type",
    "target_id",
    "status",
    "has_password",
    "one_time_use",
    "expires_at",
    "created_at",
    "url",
  ];
  const origin = typeof window !== "undefined" ? window.location.origin : "";
  const lines = [header.join(",")];
  for (const s of shares) {
    lines.push(
      [
        s.id,
        s.target_type,
        s.target_id,
        computeStatus(s),
        String(s.has_password),
        String(s.one_time_use),
        s.expires_at ?? "",
        s.created_at,
        `${origin}/share/${s.token}`,
      ]
        .map(csvEscape)
        .join(","),
    );
  }
  return lines.join("\n");
}

function downloadCSV(content: string, filename: string) {
  const blob = new Blob([content], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

export default function SharesPage() {
  usePageTitle("Shares", "View-only links");
  const qc = useQueryClient();

  const { data: shares = [], isLoading } = useQuery<Share[]>({
    queryKey: ["shares"],
    queryFn: () => sharesApi.list(),
  });

  const [filter, setFilter] = useState<FilterTab>("all");
  const [copiedID, setCopiedID] = useState<string | null>(null);

  const revoke = useMutation({
    mutationFn: (id: string) => sharesApi.revoke(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["shares"] }),
  });

  const counts = useMemo(() => {
    const acc = { all: shares.length, active: 0, expired: 0, revoked: 0, consumed: 0 };
    for (const s of shares) {
      const k = computeStatus(s);
      acc[k] += 1;
    }
    return acc;
  }, [shares]);

  const filtered = useMemo(() => {
    if (filter === "all") return shares;
    return shares.filter((s) => {
      const k = computeStatus(s);
      if (filter === "active") return k === "active";
      if (filter === "expired") return k === "expired" || k === "consumed";
      if (filter === "revoked") return k === "revoked";
      return true;
    });
  }, [shares, filter]);

  async function copyLink(s: Share) {
    const url = `${window.location.origin}/share/${s.token}`;
    try {
      await navigator.clipboard.writeText(url);
      setCopiedID(s.id);
      setTimeout(() => setCopiedID((cur) => (cur === s.id ? null : cur)), 2000);
    } catch {
      alert(url);
    }
  }

  function exportCSV() {
    if (filtered.length === 0) return;
    const date = new Date().toISOString().slice(0, 10);
    downloadCSV(buildCSV(filtered), `shares-${date}.csv`);
  }

  return (
    <div className="space-y-6">
      <SectionHead
        level="h1"
        title="Shares"
        subtitle="View-only share links you've created."
        action={
          <Button
            variant="secondary"
            onClick={exportCSV}
            disabled={filtered.length === 0}
          >
            <Icon name="Download" size={14} /> Export CSV
          </Button>
        }
      />

      <Tabs
        value={filter}
        onChange={(v) => setFilter(v as FilterTab)}
        items={[
          {
            value: "all",
            label: "All",
            badge: <Badge color="slate">{counts.all}</Badge>,
          },
          {
            value: "active",
            label: "Active",
            badge: <Badge color="accent">{counts.active}</Badge>,
          },
          {
            value: "expired",
            label: "Expired",
            badge: <Badge color="warn">{counts.expired + counts.consumed}</Badge>,
          },
          {
            value: "revoked",
            label: "Revoked",
            badge: <Badge color="danger">{counts.revoked}</Badge>,
          },
        ]}
      />

      {isLoading ? (
        <Card><div className="text-[12px] text-text-3">Loading…</div></Card>
      ) : shares.length === 0 ? (
        <Card>
          <EmptyState
            icon={<Icon name="Share2" size={18} />}
            title="No share links yet"
            description="Create share links from the Files page to see them here."
          />
        </Card>
      ) : filtered.length === 0 ? (
        <Card>
          <div className="py-6 text-center text-[12px] text-text-3">
            No shares in this view.
          </div>
        </Card>
      ) : (
        <Card padding="none">
          <div className="overflow-x-auto">
            <Table>
              <THead>
                <tr>
                  <THCell>Target</THCell>
                  <THCell className="w-28">Status</THCell>
                  <THCell className="w-40">Options</THCell>
                  <THCell className="w-40">Expires</THCell>
                  <THCell className="w-32">Created</THCell>
                  <THCell className="w-24 text-right">Actions</THCell>
                </tr>
              </THead>
              <tbody>
                {filtered.map((s) => {
                  const kind = computeStatus(s);
                  const meta = statusMeta[kind];
                  return (
                    <TRow key={s.id}>
                      <TCell>
                        <div className="flex items-center gap-2 min-w-0">
                          <span className="inline-flex h-6 w-6 items-center justify-center rounded bg-surface-2 text-text-3">
                            <Icon
                              name={s.target_type === "folder" ? "Folder" : "File"}
                              size={12}
                            />
                          </span>
                          <div className="min-w-0">
                            <div className="text-[12px] text-text-2 capitalize">
                              {s.target_type}
                            </div>
                            <div className="font-mono text-[11px] text-text-3 truncate">
                              {s.target_id}
                            </div>
                          </div>
                        </div>
                      </TCell>
                      <TCell>
                        <Badge color={meta.color} dot>
                          {meta.label}
                        </Badge>
                      </TCell>
                      <TCell>
                        <div className="flex flex-wrap items-center gap-1">
                          {s.has_password && (
                            <Badge color="warn">
                              <Icon name="Lock" size={10} /> password
                            </Badge>
                          )}
                          {s.one_time_use && (
                            <Badge color="info">one-time</Badge>
                          )}
                          {!s.has_password && !s.one_time_use && (
                            <span className="text-[11px] text-text-3">—</span>
                          )}
                        </div>
                      </TCell>
                      <TCell className="text-[12px] text-text-2 whitespace-nowrap">
                        {s.expires_at
                          ? new Date(s.expires_at).toLocaleString()
                          : "Never"}
                      </TCell>
                      <TCell className="text-[12px] text-text-3 whitespace-nowrap">
                        {new Date(s.created_at).toLocaleDateString()}
                      </TCell>
                      <TCell>
                        <div className="flex items-center justify-end gap-1">
                          <Tooltip
                            content={copiedID === s.id ? "Copied!" : "Copy link"}
                            side="top"
                          >
                            <IconButton
                              size="sm"
                              aria-label="Copy link"
                              onClick={() => copyLink(s)}
                            >
                              <Icon
                                name={copiedID === s.id ? "Check" : "Copy"}
                                size={12}
                                className={
                                  copiedID === s.id ? "text-accent-2" : undefined
                                }
                              />
                            </IconButton>
                          </Tooltip>
                          {!s.revoked && (
                            <Tooltip content="Revoke" side="top">
                              <IconButton
                                size="sm"
                                aria-label="Revoke"
                                onClick={() => {
                                  if (confirm("Revoke this share link?"))
                                    revoke.mutate(s.id);
                                }}
                              >
                                <Icon name="Trash2" size={12} />
                              </IconButton>
                            </Tooltip>
                          )}
                        </div>
                      </TCell>
                    </TRow>
                  );
                })}
              </tbody>
            </Table>
          </div>
        </Card>
      )}
    </div>
  );
}
