"use client";

import Link from "next/link";
import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/lib/stores/auth-store";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { storagesApi, Storage } from "@/lib/api/storages";
import { filesApi, FileItem } from "@/lib/api/files";
import { sharesApi, Share } from "@/lib/api/shares";
import { activityApi, ActivityEntry } from "@/lib/api/activity";
import { aiApi, AIConfig } from "@/lib/api/ai";
import { ActivityFeed } from "@/components/activity/activity-feed";
import { Card } from "@/components/ui/v2/card";
import { SectionHead } from "@/components/ui/v2/section-head";
import { Badge } from "@/components/ui/v2/badge";
import { Icon } from "@/components/ui/v2/icon";
import { Progress } from "@/components/ui/v2/progress";
import { SegBar } from "@/components/ui/v2/seg-bar";
import { Sparkbars } from "@/components/ui/v2/sparkbars";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { FileIcon, kindFromMime } from "@/components/ui/v2/file-icon";

function formatBytes(n: number): string {
  if (!n) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const i = Math.floor(Math.log(n) / Math.log(1024));
  return `${(n / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

function daysAgoKey(d: Date) {
  return `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()}`;
}

function build30DayBuckets(files: FileItem[]): number[] {
  const now = new Date();
  const buckets = new Array<number>(30).fill(0);
  const keyToIdx = new Map<string, number>();
  for (let i = 0; i < 30; i++) {
    const d = new Date(now);
    d.setDate(d.getDate() - (29 - i));
    keyToIdx.set(daysAgoKey(d), i);
  }
  for (const f of files) {
    const when = f.uploaded_at ?? f.created_at;
    if (!when) continue;
    const d = new Date(when);
    const k = daysAgoKey(d);
    const idx = keyToIdx.get(k);
    if (idx !== undefined) buckets[idx] += 1;
  }
  return buckets;
}

interface MimeCategory {
  label: string;
  color: string;
  match: (f: FileItem) => boolean;
}

const mimeCategories: MimeCategory[] = [
  {
    label: "PDF",
    color: "var(--danger)",
    match: (f) => kindFromMime(f.mime_type, f.name) === "pdf",
  },
  {
    label: "Images",
    color: "var(--info)",
    match: (f) => kindFromMime(f.mime_type, f.name) === "image",
  },
  {
    label: "Docs",
    color: "var(--violet)",
    match: (f) => ["word", "excel", "powerpoint", "text"].includes(kindFromMime(f.mime_type, f.name)),
  },
  {
    label: "Video",
    color: "var(--warn)",
    match: (f) => ["video", "audio"].includes(kindFromMime(f.mime_type, f.name)),
  },
  {
    label: "Other",
    color: "var(--text-3)",
    match: () => true,
  },
];

function categorize(files: FileItem[]) {
  const counts = new Array<number>(mimeCategories.length).fill(0);
  outer: for (const f of files) {
    for (let i = 0; i < mimeCategories.length - 1; i++) {
      if (mimeCategories[i].match(f)) {
        counts[i] += 1;
        continue outer;
      }
    }
    counts[counts.length - 1] += 1;
  }
  return counts;
}

export default function DashboardPage() {
  usePageTitle("Overview");
  const user = useAuthStore((s) => s.user);

  const { data: storages = [] } = useQuery<Storage[]>({
    queryKey: ["storages"],
    queryFn: () => storagesApi.list(),
  });
  const { data: files = [] } = useQuery<FileItem[]>({
    queryKey: ["files", "dashboard"],
    queryFn: () => filesApi.list({ limit: "500" }),
  });
  const { data: shares = [] } = useQuery<Share[]>({
    queryKey: ["shares"],
    queryFn: () => sharesApi.list(),
  });

  const totalBytes = storages.reduce((a, s) => a + s.used_bytes, 0);
  const totalObjects = storages.reduce((a, s) => a + s.object_count, 0);
  const activeShares = shares.filter(
    (s) =>
      !s.revoked &&
      !s.consumed &&
      (!s.expires_at || new Date(s.expires_at) > new Date()),
  ).length;

  const uploadBuckets = useMemo(() => build30DayBuckets(files), [files]);
  const uploadsTotal = uploadBuckets.reduce((a, b) => a + b, 0);
  const mimeCounts = useMemo(() => categorize(files), [files]);
  const mimeSegments = mimeCategories
    .map((c, i) => ({ value: mimeCounts[i], color: c.color, label: c.label }))
    .filter((s) => s.value > 0);


  const stats = [
    {
      label: "Connected storages",
      value: storages.length.toString(),
      icon: "Cloud",
      href: "/dashboard/storages",
      sub: storages.length === 0 ? "None connected" : undefined,
    },
    {
      label: "Total used",
      value: formatBytes(totalBytes),
      icon: "HardDrive",
      sub: `${totalObjects.toLocaleString()} objects`,
    },
    {
      label: "Files",
      value: files.length.toString(),
      icon: "Files",
      href: "/dashboard/files",
    },
    {
      label: "Active shares",
      value: activeShares.toString(),
      icon: "Share2",
      href: "/dashboard/shares",
      sub: `${shares.length} total`,
    },
  ];

  return (
    <div className="space-y-6">
      <SectionHead
        level="h1"
        title={`Welcome, ${user?.display_name ?? "there"}`}
        subtitle="Overview of your document storage."
        action={
          <Link
            href="/dashboard/files"
            className="inline-flex items-center gap-1.5 h-8 px-3 rounded bg-accent text-white text-[13px] font-medium hover:bg-accent-2"
          >
            <Icon name="Upload" size={14} />
            Upload
          </Link>
        }
      />

      <AIOnboardingBanner />

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((c) => {
          const inner = (
            <Card padding="md" className="hover:border-border-strong transition-colors">
              <div className="flex items-center justify-between">
                <span className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
                  {c.label}
                </span>
                <Icon name={c.icon} size={14} className="text-text-3" />
              </div>
              <div className="text-[22px] font-semibold text-text mt-2 tabular-nums">
                {c.value}
              </div>
              {c.sub && (
                <div className="text-xs text-text-2 mt-1">{c.sub}</div>
              )}
            </Card>
          );
          return c.href ? (
            <Link key={c.label} href={c.href} className="block">
              {inner}
            </Link>
          ) : (
            <div key={c.label}>{inner}</div>
          );
        })}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
        <Card
          className="lg:col-span-3"
          title="Uploads"
          subtitle={`${uploadsTotal} files in the last 30 days`}
        >
          {uploadsTotal > 0 ? (
            <div className="flex items-end justify-between gap-4">
              <Sparkbars data={uploadBuckets} width={420} height={56} />
              <div className="text-right">
                <div className="text-[22px] font-semibold text-text tabular-nums">
                  {uploadsTotal}
                </div>
                <div className="text-[11px] text-text-3">last 30 days</div>
              </div>
            </div>
          ) : (
            <div className="py-6 text-center text-[12px] text-text-3">
              No uploads yet.
            </div>
          )}
        </Card>

        <Card
          className="lg:col-span-2"
          title="By type"
          subtitle={`${files.length} files total`}
        >
          {mimeSegments.length > 0 ? (
            <>
              <SegBar segments={mimeSegments} />
              <div className="mt-3 space-y-1">
                {mimeSegments.map((s) => (
                  <div
                    key={s.label}
                    className="flex items-center justify-between text-[11px] text-text-2"
                  >
                    <span className="flex items-center gap-1.5">
                      <span
                        className="h-2 w-2 rounded-full"
                        style={{ background: s.color }}
                      />
                      {s.label}
                    </span>
                    <span className="tabular-nums text-text-3">{s.value}</span>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <div className="py-6 text-center text-[12px] text-text-3">
              No files yet.
            </div>
          )}
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <RecentActivityCard />


        <Card title="Storage health">
          {storages.length > 0 ? (
            <ul className="space-y-3">
              {storages.slice(0, 4).map((s) => {
                const pct = Math.min(100, (s.used_bytes / Math.max(1, 100 * 1024 ** 3)) * 100);
                const err = !!s.last_error;
                return (
                  <li key={s.id}>
                    <div className="flex items-center justify-between text-[12px]">
                      <span className="truncate text-text">{s.display_name}</span>
                      <Badge color={err ? "danger" : "accent"} dot>
                        {err ? "Error" : "Active"}
                      </Badge>
                    </div>
                    <Progress value={pct} className="mt-1.5" />
                    <div className="text-[11px] text-text-3 mt-1">
                      {formatBytes(s.used_bytes)} · {s.object_count.toLocaleString()} objects
                    </div>
                  </li>
                );
              })}
            </ul>
          ) : (
            <div className="py-4 text-center text-[12px] text-text-3">
              No storages connected.
            </div>
          )}
        </Card>

        <Card title="Team">
          <EmptyState
            icon={<Icon name="Users" size={18} />}
            title="Invite your team"
            description="Team collaboration arrives in the next release."
          />
        </Card>
      </div>

      <StarredFilesCard />

      {storages.length === 0 && (
        <Card>
          <EmptyState
            icon={<Icon name="Cloud" size={18} />}
            title="No storage connected yet"
            description="Connect Cloudflare R2, AWS S3, or MinIO to start uploading."
            action={
              <Link
                href="/dashboard/storages"
                className="inline-flex items-center gap-1.5 h-8 px-3 rounded bg-accent text-white text-[13px] font-medium hover:bg-accent-2"
              >
                <Icon name="Plus" size={14} /> Add storage
              </Link>
            }
          />
        </Card>
      )}
    </div>
  );
}

function RecentActivityCard() {
  const { data = [], isLoading } = useQuery<ActivityEntry[]>({
    queryKey: ["activity", "recent"],
    queryFn: () => activityApi.recent(8),
    staleTime: 30_000,
  });
  return (
    <Card title="Recent activity">
      {isLoading ? (
        <div className="text-[12px] text-text-3">Loading…</div>
      ) : (
        <ActivityFeed
          entries={data}
          emptyTitle="No activity yet"
          emptyDescription="Upload a file or create a folder to start a trail."
        />
      )}
    </Card>
  );
}

function StarredFilesCard() {
  const { data = [] } = useQuery<FileItem[]>({
    queryKey: ["files", "starred"],
    queryFn: () => filesApi.list({ starred: "true", limit: "8" }),
  });
  return (
    <Card
      title="Starred files"
      action={
        data.length > 0 ? (
          <Link
            href="/dashboard/files?mode=starred"
            className="text-[12px] text-accent-2 hover:underline"
          >
            See all
          </Link>
        ) : undefined
      }
    >
      {data.length === 0 ? (
        <EmptyState
          icon={<Icon name="Star" size={18} />}
          title="No starred files yet"
          description="Click the star on any file to pin it here for quick access."
        />
      ) : (
        <div
          className="grid gap-2"
          style={{ gridTemplateColumns: "repeat(auto-fill, minmax(180px, 1fr))" }}
        >
          {data.map((f) => (
            <Link
              key={f.id}
              href={`/dashboard/files/view/${f.id}`}
              className="flex items-center gap-2 rounded-lg border border-border bg-surface px-3 py-2 hover:border-border-strong"
            >
              <FileIcon mime={f.mime_type} name={f.name} size="sm" />
              <div className="min-w-0 flex-1">
                <div className="truncate text-[12px] font-medium text-text">
                  {f.name}
                </div>
                <div className="text-[10px] text-text-3">
                  {formatBytes(f.size_bytes)}
                </div>
              </div>
              <Icon name="Star" size={12} className="text-warn fill-current" />
            </Link>
          ))}
        </div>
      )}
    </Card>
  );
}

function AIOnboardingBanner() {
  const { data = [], isLoading } = useQuery<AIConfig[]>({
    queryKey: ["ai-configs"],
    queryFn: () => aiApi.list(),
    staleTime: 60_000,
  });
  if (isLoading || data.length > 0) return null;
  return (
    <div className="flex flex-col md:flex-row md:items-center gap-3 rounded-lg border border-accent-border bg-accent-soft px-4 py-3">
      <span className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-accent text-white shrink-0">
        <Icon name="Sparkles" size={18} />
      </span>
      <div className="min-w-0 flex-1">
        <div className="text-[13px] font-semibold text-accent-2">
          Unlock AI features — add your own provider key.
        </div>
        <div className="text-[12px] text-text-2 mt-0.5">
          Auto-fill fields, smart tagging, semantic search, and more. We
          orchestrate; you pay the provider directly. Your prompts never leave
          your account.
        </div>
      </div>
      <Link
        href="/dashboard/ai"
        className="inline-flex items-center gap-1.5 h-8 px-3 rounded bg-accent text-white text-[13px] font-medium hover:bg-accent-2 shrink-0"
      >
        Add provider
        <Icon name="ArrowRight" size={12} />
      </Link>
    </div>
  );
}
