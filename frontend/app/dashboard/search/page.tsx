"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { filesApi, FileItem } from "@/lib/api/files";
import { foldersApi, Folder } from "@/lib/api/folders";
import { tagsApi, Tag } from "@/lib/api/tags";
import { aiApi, AIConfig, SemanticSearchHit } from "@/lib/api/ai";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/v2/badge";
import { Card } from "@/components/ui/v2/card";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { FileIcon } from "@/components/ui/v2/file-icon";
import { Icon } from "@/components/ui/v2/icon";
import { Kbd } from "@/components/ui/v2/kbd";
import { SectionHead } from "@/components/ui/v2/section-head";
import { Tabs } from "@/components/ui/v2/tabs";
import { ResultFile } from "@/components/search/result-file";
import { ResultFolder } from "@/components/search/result-folder";
import { ResultTag } from "@/components/search/result-tag";

type Scope = "all" | "files" | "tags" | "folders" | "people";

function useDebounced<T>(value: T, delay = 300) {
  const [d, setD] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setD(value), delay);
    return () => clearTimeout(t);
  }, [value, delay]);
  return d;
}

type Mode = "keyword" | "semantic";

export default function SearchPage() {
  usePageTitle("Search");
  const router = useRouter();
  const [scope, setScope] = useState<Scope>("all");
  const [mode, setMode] = useState<Mode>("keyword");
  const [input, setInput] = useState("");
  const query = useDebounced(input, 300);
  const inputRef = useRef<HTMLInputElement>(null);

  const { data: configs = [] } = useQuery<AIConfig[]>({
    queryKey: ["ai-configs"],
    queryFn: () => aiApi.list(),
    staleTime: 60_000,
  });
  const hasEmbedConfig = configs.some((c) => c.capabilities.includes("embed"));

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (input) {
          setInput("");
        } else {
          router.back();
        }
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [input, router]);

  const filesQuery = useQuery({
    queryKey: ["search", "files", query],
    queryFn: () => filesApi.list({ q: query, limit: "50" }),
    enabled:
      !!query && mode === "keyword" && (scope === "all" || scope === "files"),
    staleTime: 15_000,
  });

  const semanticQuery = useQuery({
    queryKey: ["search", "semantic", query],
    queryFn: () => aiApi.semanticSearch(query, 20),
    enabled: !!query && mode === "semantic",
    staleTime: 15_000,
  });

  const foldersQuery = useQuery({
    queryKey: ["search", "folders"],
    queryFn: () => foldersApi.list(),
    enabled: !!query && (scope === "all" || scope === "folders"),
    staleTime: 60_000,
  });

  const tagsQuery = useQuery({
    queryKey: ["search", "tags"],
    queryFn: () => tagsApi.list(),
    enabled: !!query && (scope === "all" || scope === "tags"),
    staleTime: 60_000,
  });

  const q = query.trim().toLowerCase();
  const matches = (s: string) => !q || s.toLowerCase().includes(q);

  const fileResults: FileItem[] = useMemo(
    () => (filesQuery.data ?? []).filter((f) => matches(f.name)),
    [filesQuery.data, q],
  );
  const folderResults: Folder[] = useMemo(
    () => (foldersQuery.data ?? []).filter((f) => matches(f.name)),
    [foldersQuery.data, q],
  );
  const tagResults: Tag[] = useMemo(
    () => (tagsQuery.data ?? []).filter((t) => matches(t.name)),
    [tagsQuery.data, q],
  );

  const tagByID = useMemo(
    () => Object.fromEntries((tagsQuery.data ?? []).map((t) => [t.id, t])),
    [tagsQuery.data],
  );

  const showFiles = scope === "all" || scope === "files";
  const showFolders = scope === "all" || scope === "folders";
  const showTags = scope === "all" || scope === "tags";

  const semanticHits: SemanticSearchHit[] = semanticQuery.data?.hits ?? [];

  const anyResults =
    mode === "semantic"
      ? semanticHits.length > 0
      : (showFiles && fileResults.length > 0) ||
        (showFolders && folderResults.length > 0) ||
        (showTags && tagResults.length > 0);

  return (
    <div className="max-w-4xl mx-auto space-y-5">
      <SectionHead
        level="h1"
        title="Search"
        subtitle="Search across files, folders, and tags."
        action={
          <span className="flex items-center gap-1 text-[11px] text-text-3">
            <Kbd>Esc</Kbd>
            <span>to exit</span>
          </span>
        }
      />

      <div className="relative">
        <Icon
          name="Search"
          size={16}
          className="absolute left-3 top-1/2 -translate-y-1/2 text-text-3"
        />
        <input
          ref={inputRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={
            mode === "semantic"
              ? "Describe what you're looking for…"
              : "Search files, folders, tags…"
          }
          className="h-11 w-full rounded-lg border border-border bg-surface pl-9 pr-3 text-[14px] text-text placeholder:text-text-3 focus:outline-none focus:border-accent focus:ring-2 focus:ring-accent/20"
        />
      </div>

      <div className="flex items-center gap-2">
        <div className="inline-flex rounded-lg border border-border bg-surface p-0.5">
          {(["keyword", "semantic"] as Mode[]).map((m) => {
            const active = mode === m;
            const disabled = m === "semantic" && !hasEmbedConfig;
            return (
              <button
                key={m}
                type="button"
                disabled={disabled}
                onClick={() => setMode(m)}
                className={cn(
                  "inline-flex items-center gap-1.5 h-7 px-2.5 rounded-md text-[12px] transition-colors",
                  active
                    ? "bg-accent-soft text-accent-2 font-medium"
                    : "text-text-2 hover:text-text",
                  disabled && "opacity-50 cursor-not-allowed",
                )}
                title={
                  disabled
                    ? "Add an embed-capable AI provider to enable semantic search"
                    : undefined
                }
              >
                <Icon name={m === "semantic" ? "Sparkles" : "Search"} size={11} />
                {m === "semantic" ? "Semantic" : "Keyword"}
              </button>
            );
          })}
        </div>
        {mode === "semantic" && !hasEmbedConfig && (
          <Link
            href="/dashboard/ai"
            className="text-[11px] text-accent-2 hover:underline"
          >
            Add embed provider →
          </Link>
        )}
        {mode === "semantic" && (
          <Link
            href="/dashboard/ask"
            className="ml-auto inline-flex items-center gap-1 text-[12px] text-accent-2 hover:underline"
          >
            <Icon name="MessageSquare" size={12} /> Ask your docs
          </Link>
        )}
      </div>

      <Tabs
        value={scope}
        onChange={(v) => setScope(v as Scope)}
        items={[
          { value: "all", label: "Everything" },
          {
            value: "files",
            label: "Files",
            badge: query ? (
              <Badge color="slate">{fileResults.length}</Badge>
            ) : undefined,
          },
          {
            value: "folders",
            label: "Folders",
            badge: query ? (
              <Badge color="slate">{folderResults.length}</Badge>
            ) : undefined,
          },
          {
            value: "tags",
            label: "Tags",
            badge: query ? (
              <Badge color="slate">{tagResults.length}</Badge>
            ) : undefined,
          },
          {
            value: "people",
            label: "People",
            badge: <Badge color="slate">Soon</Badge>,
          },
        ]}
      />

      {!query ? (
        <Card>
          <EmptyState
            icon={<Icon name="Search" size={18} />}
            title="Start typing to search"
            description="Results update as you type. Press Esc to close."
          />
        </Card>
      ) : scope === "people" ? (
        <Card>
          <EmptyState
            icon={<Icon name="Users" size={18} />}
            title="People search coming soon"
            description="Team management arrives in the next release."
          />
        </Card>
      ) : !anyResults ? (
        <Card>
          <EmptyState
            icon={<Icon name="SearchX" size={18} />}
            title="No results"
            description={`Nothing matches “${query}” in ${scope === "all" ? "your workspace" : scope}.`}
          />
        </Card>
      ) : mode === "semantic" ? (
        <div className="space-y-2">
          <div className="flex items-center justify-between px-1">
            <h2 className="text-[11px] font-semibold uppercase tracking-wide text-text-3">
              Semantic matches
            </h2>
            <span className="text-[11px] text-text-3">
              {semanticHits.length}
            </span>
          </div>
          <ul className="space-y-1.5">
            {semanticHits.map((h) => (
              <li key={h.file_id}>
                <Link
                  href={`/dashboard/files/view/${h.file_id}`}
                  className="flex items-start gap-3 rounded-lg border border-border bg-surface px-3 py-2.5 hover:border-border-strong"
                >
                  <FileIcon mime={h.mime} name={h.name} size="md" />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-1.5">
                      <span className="truncate text-[13px] font-medium text-text">
                        {h.name}
                      </span>
                      <Badge color="accent">
                        {(h.score * 100).toFixed(0)}%
                      </Badge>
                    </div>
                    {h.text_preview && (
                      <div className="mt-0.5 text-[11px] text-text-3 line-clamp-2">
                        {h.text_preview}
                      </div>
                    )}
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      ) : (
        <div className="space-y-5">
          {showFiles && fileResults.length > 0 && (
            <section className="space-y-2">
              <div className="flex items-center justify-between px-1">
                <h2 className="text-[11px] font-semibold uppercase tracking-wide text-text-3">
                  Files
                </h2>
                <span className="text-[11px] text-text-3">
                  {fileResults.length}
                </span>
              </div>
              <div className="space-y-1.5">
                {fileResults.map((f) => (
                  <ResultFile
                    key={f.id}
                    file={f}
                    query={query}
                    tagByID={tagByID}
                  />
                ))}
              </div>
            </section>
          )}

          {showFolders && folderResults.length > 0 && (
            <section className="space-y-2">
              <div className="flex items-center justify-between px-1">
                <h2 className="text-[11px] font-semibold uppercase tracking-wide text-text-3">
                  Folders
                </h2>
                <span className="text-[11px] text-text-3">
                  {folderResults.length}
                </span>
              </div>
              <div className="space-y-1.5">
                {folderResults.map((f) => (
                  <ResultFolder key={f.id} folder={f} query={query} />
                ))}
              </div>
            </section>
          )}

          {showTags && tagResults.length > 0 && (
            <section className="space-y-2">
              <div className="flex items-center justify-between px-1">
                <h2 className="text-[11px] font-semibold uppercase tracking-wide text-text-3">
                  Tags
                </h2>
                <span className="text-[11px] text-text-3">
                  {tagResults.length}
                </span>
              </div>
              <div className="flex flex-wrap gap-1.5">
                {tagResults.map((t) => (
                  <ResultTag key={t.id} tag={t} />
                ))}
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  );
}
