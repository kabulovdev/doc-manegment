"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { CreateStorageInput, Storage, storagesApi } from "@/lib/api/storages";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { StorageForm } from "@/components/storages/storage-form";
import { Badge } from "@/components/ui/v2/badge";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { Dialog } from "@/components/ui/v2/dialog";
import { DropdownMenu } from "@/components/ui/v2/dropdown-menu";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { Icon } from "@/components/ui/v2/icon";
import { IconButton } from "@/components/ui/v2/icon-button";
import { Progress } from "@/components/ui/v2/progress";
import { SectionHead } from "@/components/ui/v2/section-head";
import { Toggle } from "@/components/ui/v2/toggle";
import { Table, THead, THCell, TRow, TCell } from "@/components/ui/v2/table";

function formatBytes(n: number): string {
  if (!n) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const i = Math.floor(Math.log(n) / Math.log(1024));
  return `${(n / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

function relTime(iso?: string | null) {
  if (!iso) return "never";
  const d = new Date(iso);
  const diff = (Date.now() - d.getTime()) / 1000;
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 86400 * 7) return `${Math.floor(diff / 86400)}d ago`;
  return d.toLocaleString();
}

type Status = { kind: "active" | "stale" | "error"; label: string; tone: "accent" | "warn" | "danger" };

function computeStatus(s: Storage): Status {
  if (s.last_error) return { kind: "error", label: "Error", tone: "danger" };
  if (!s.last_sync_at) return { kind: "stale", label: "Stale", tone: "warn" };
  const ageMs = Date.now() - new Date(s.last_sync_at).getTime();
  if (ageMs > 24 * 3600 * 1000) return { kind: "stale", label: "Stale", tone: "warn" };
  return { kind: "active", label: "Active", tone: "accent" };
}

const providerColors: Record<Storage["provider"], string> = {
  r2: "#f59e0b",
  s3: "#6366f1",
  minio: "#10b981",
};

export default function StoragesPage() {
  usePageTitle("Storages", "Connect your own buckets");
  const qc = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data: storages = [], isLoading } = useQuery<Storage[]>({
    queryKey: ["storages"],
    queryFn: () => storagesApi.list(),
  });

  const totalUsed = storages.reduce((a, s) => a + s.used_bytes, 0);
  const totalObjects = storages.reduce((a, s) => a + s.object_count, 0);

  const resyncMut = useMutation({
    mutationFn: (id: string) => storagesApi.resync(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["storages"] }),
  });
  const deleteMut = useMutation({
    mutationFn: (id: string) => storagesApi.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["storages"] }),
  });
  const forceDeleteMut = useMutation({
    mutationFn: (id: string) => storagesApi.remove(id, true),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["storages"] }),
  });
  const createMut = useMutation({
    mutationFn: (input: CreateStorageInput) => storagesApi.create(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["storages"] });
      setDialogOpen(false);
    },
  });

  async function onTest(id: string) {
    try {
      const res = await storagesApi.test(id);
      if (res.ok) {
        alert("Connection OK");
      } else {
        alert(`Test failed: ${res.error ?? "unknown"}`);
      }
    } catch (e) {
      alert((e as Error).message ?? "Test failed");
    }
  }

  async function onDelete(id: string) {
    try {
      await deleteMut.mutateAsync(id);
    } catch (e) {
      const err = e as { status?: number; message?: string };
      if (err.status === 409) {
        if (
          confirm(
            "This storage still has files. Force delete will mark those files as deleted (bucket contents remain untouched). Proceed?",
          )
        ) {
          await forceDeleteMut.mutateAsync(id);
        }
      } else {
        alert(err.message ?? "Delete failed");
      }
    }
  }

  const syncEvents = useMemo(() => {
    return [...storages]
      .map((s) => ({
        id: s.id,
        name: s.display_name,
        when: s.last_sync_at ?? s.created_at,
        error: s.last_error,
      }))
      .sort((a, b) => new Date(b.when).getTime() - new Date(a.when).getTime())
      .slice(0, 6);
  }, [storages]);

  return (
    <div className="space-y-6">
      <SectionHead
        level="h1"
        title="Storages"
        subtitle="Connect your own object storage buckets."
        action={
          <Button variant="accent" onClick={() => setDialogOpen(true)}>
            <Icon name="Plus" size={14} /> Add storage
          </Button>
        }
      />

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Total capacity" value="—" sub="No quota set" />
        <StatCard label="Used" value={formatBytes(totalUsed)} sub={`${storages.length} storages`} />
        <StatCard label="Objects" value={totalObjects.toLocaleString()} />
        <StatCard label="Monthly cost" value="—" sub="Coming soon" />
      </div>

      {isLoading ? (
        <Card padding="md">
          <div className="text-[12px] text-text-3">Loading…</div>
        </Card>
      ) : storages.length === 0 ? (
        <Card>
          <EmptyState
            icon={<Icon name="Cloud" size={18} />}
            title="No storage yet"
            description="Connect your first bucket to start uploading files."
            action={
              <Button variant="accent" onClick={() => setDialogOpen(true)}>
                <Icon name="Plus" size={14} /> Add storage
              </Button>
            }
          />
        </Card>
      ) : (
        <Card padding="none">
          <div className="overflow-x-auto">
            <Table>
              <THead>
                <tr>
                  <THCell>Name</THCell>
                  <THCell className="w-28">Kind</THCell>
                  <THCell>Endpoint</THCell>
                  <THCell className="w-56">Usage</THCell>
                  <THCell className="w-24">Objects</THCell>
                  <THCell className="w-28">Status</THCell>
                  <THCell className="w-10" />
                </tr>
              </THead>
              <tbody>
                {storages.map((s) => {
                  const status = computeStatus(s);
                  const pct = Math.min(100, (s.used_bytes / Math.max(1, 100 * 1024 ** 3)) * 100);
                  const busy = resyncMut.isPending || deleteMut.isPending;
                  return (
                    <TRow key={s.id}>
                      <TCell>
                        <div className="flex items-center gap-2">
                          <span
                            className="inline-flex h-6 w-6 items-center justify-center rounded text-white"
                            style={{ background: providerColors[s.provider] }}
                          >
                            <Icon name="Cloud" size={12} />
                          </span>
                          <div className="min-w-0">
                            <div className="truncate text-text font-medium">
                              {s.display_name}
                            </div>
                            <div className="text-[11px] text-text-3 truncate">
                              {s.bucket}
                            </div>
                          </div>
                        </div>
                      </TCell>
                      <TCell>
                        <Badge color="slate">{s.provider.toUpperCase()}</Badge>
                      </TCell>
                      <TCell className="font-mono text-[11px] text-text-2 truncate max-w-xs">
                        {s.endpoint}
                      </TCell>
                      <TCell>
                        <Progress value={pct} />
                        <div className="text-[11px] text-text-3 mt-1 tabular-nums">
                          {formatBytes(s.used_bytes)}
                        </div>
                      </TCell>
                      <TCell className="tabular-nums text-text-2">
                        {s.object_count.toLocaleString()}
                      </TCell>
                      <TCell>
                        <Badge color={status.tone} dot>
                          {status.label}
                        </Badge>
                      </TCell>
                      <TCell className="text-right">
                        <DropdownMenu
                          align="end"
                          trigger={({ toggle }) => (
                            <IconButton
                              size="sm"
                              onClick={toggle}
                              aria-label="Actions"
                              disabled={busy}
                            >
                              <Icon name="MoreHorizontal" size={14} />
                            </IconButton>
                          )}
                          items={[
                            {
                              label: "Test connection",
                              icon: <Icon name="Zap" size={12} />,
                              onSelect: () => onTest(s.id),
                            },
                            {
                              label: "Resync",
                              icon: <Icon name="RefreshCw" size={12} />,
                              onSelect: () => resyncMut.mutate(s.id),
                            },
                            { separator: true, label: "" },
                            {
                              label: "Delete",
                              icon: <Icon name="Trash2" size={12} />,
                              onSelect: () => onDelete(s.id),
                              danger: true,
                            },
                          ]}
                        />
                      </TCell>
                    </TRow>
                  );
                })}
              </tbody>
            </Table>
          </div>
        </Card>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card title="Sync queue" subtitle="Recent activity">
          {syncEvents.length === 0 ? (
            <div className="text-[12px] text-text-3">No sync events yet.</div>
          ) : (
            <ul className="space-y-2">
              {syncEvents.map((e) => (
                <li key={e.id + e.when} className="flex items-center gap-2.5 text-[12px]">
                  <span
                    className={`inline-flex h-6 w-6 items-center justify-center rounded ${e.error ? "bg-red-50 text-red-600" : "bg-accent-soft text-accent-2"}`}
                  >
                    <Icon name={e.error ? "AlertCircle" : "RefreshCw"} size={12} />
                  </span>
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-text">{e.name}</div>
                    <div className="text-[11px] text-text-3 truncate">
                      {e.error ?? "Resynced successfully"}
                    </div>
                  </div>
                  <span className="text-[11px] text-text-3 shrink-0">{relTime(e.when)}</span>
                </li>
              ))}
            </ul>
          )}
        </Card>

        <Card
          title="Storage policies"
          subtitle="Per-storage automation"
          action={<Badge color="slate">Coming soon</Badge>}
        >
          <div className="space-y-3 opacity-80">
            {[
              { label: "Archive after 90 days", hint: "Move cold files to archive tier." },
              { label: "Encrypt at rest", hint: "AES-256 server-side encryption." },
              { label: "Replicate to backup", hint: "Cross-region mirror (pending)." },
              { label: "Delete from trash after 30 days", hint: "Auto purge soft-deleted files." },
            ].map((p) => (
              <div key={p.label} className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="text-[13px] text-text">{p.label}</div>
                  <div className="text-[11px] text-text-3">{p.hint}</div>
                </div>
                <Toggle checked={false} disabled aria-label={p.label} />
              </div>
            ))}
          </div>
        </Card>
      </div>

      <Dialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        title="Add storage"
        description="Provide credentials for a new bucket. We&rsquo;ll test the connection before saving."
        size="md"
      >
        <StorageForm
          onSubmit={(input) => createMut.mutateAsync(input).then(() => undefined)}
          onCancel={() => setDialogOpen(false)}
        />
      </Dialog>
    </div>
  );
}

function StatCard({
  label,
  value,
  sub,
}: {
  label: string;
  value: string;
  sub?: string;
}) {
  return (
    <Card padding="md">
      <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
        {label}
      </div>
      <div className="text-[22px] font-semibold text-text mt-1 tabular-nums">
        {value}
      </div>
      {sub && <div className="text-[11px] text-text-3 mt-0.5">{sub}</div>}
    </Card>
  );
}
