"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { useMutation, useQuery } from "@tanstack/react-query";
import { aiApi, AIConfig, AskResponse } from "@/lib/api/ai";
import { ApiError } from "@/lib/api/client";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { Badge } from "@/components/ui/v2/badge";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { FileIcon } from "@/components/ui/v2/file-icon";
import { Icon } from "@/components/ui/v2/icon";
import { SectionHead } from "@/components/ui/v2/section-head";
import { Textarea } from "@/components/ui/v2/textarea";

interface ChatEntry {
  question: string;
  response: AskResponse;
}

export default function AskPage() {
  usePageTitle("Ask your docs");
  const [question, setQuestion] = useState("");
  const [history, setHistory] = useState<ChatEntry[]>([]);

  const { data: configs = [] } = useQuery<AIConfig[]>({
    queryKey: ["ai-configs"],
    queryFn: () => aiApi.list(),
    staleTime: 60_000,
  });
  const hasChat = configs.some((c) => c.capabilities.includes("chat"));
  const hasEmbed = configs.some((c) => c.capabilities.includes("embed"));

  const ask = useMutation({
    mutationFn: (q: string) => aiApi.ask(q, 5),
    onSuccess: (res, q) => {
      setHistory((h) => [...h, { question: q, response: res }]);
      setQuestion("");
    },
  });

  function submit(e: FormEvent) {
    e.preventDefault();
    const q = question.trim();
    if (!q || ask.isPending) return;
    ask.mutate(q);
  }

  if (!hasChat || !hasEmbed) {
    const missing = !hasChat && !hasEmbed
      ? "chat and embed"
      : !hasChat
        ? "chat"
        : "embed";
    return (
      <div className="space-y-5 max-w-4xl">
        <SectionHead level="h1" title="Ask your docs" />
        <Card>
          <EmptyState
            icon={<Icon name="Sparkles" size={18} />}
            title={`Needs a ${missing}-capable provider`}
            description="Bring your own key — Anthropic for chat, OpenAI or Google for embeddings. Top-k retrieval runs over your own Qdrant index."
            action={
              <Link
                href="/dashboard/ai"
                className="inline-flex items-center gap-1.5 h-8 px-3 rounded bg-accent text-white text-[13px] font-medium hover:bg-accent-2"
              >
                <Icon name="Plus" size={14} /> Add provider
              </Link>
            }
          />
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-4 max-w-4xl">
      <SectionHead
        level="h1"
        title="Ask your docs"
        subtitle="Questions are answered from your files. Every claim cites the source."
      />

      <Card padding="md">
        <form onSubmit={submit} className="space-y-2">
          <Textarea
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            placeholder="What does the invoice from Acme say about net-30 terms?"
            rows={3}
            onKeyDown={(e) => {
              if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
                submit(e);
              }
            }}
          />
          <div className="flex items-center justify-between">
            <span className="text-[11px] text-text-3">
              Uses top-5 semantic hits as context.{" "}
              <kbd className="inline-flex items-center rounded border border-border bg-surface-2 px-1 py-0.5 font-mono text-[10px]">
                ⌘+Enter
              </kbd>{" "}
              to send.
            </span>
            <Button
              type="submit"
              variant="accent"
              disabled={!question.trim() || ask.isPending}
            >
              {ask.isPending ? "Thinking…" : "Ask"}
              <Icon name="ArrowRight" size={12} />
            </Button>
          </div>
        </form>
        {ask.isError && (
          <div className="mt-2 flex gap-2 rounded border border-red-200 bg-red-50 px-2.5 py-1.5 text-[11px] text-red-700">
            <Icon name="AlertCircle" size={12} className="shrink-0 mt-0.5" />
            <span className="break-words">
              {ask.error instanceof ApiError ? ask.error.message : String(ask.error)}
            </span>
          </div>
        )}
      </Card>

      {history.length === 0 ? (
        <Card>
          <EmptyState
            icon={<Icon name="MessageSquare" size={18} />}
            title="Ask anything about your documents"
            description="The answer includes citations back to the files it used. Try: “Summarize the terms in last month's contracts.”"
          />
        </Card>
      ) : (
        <div className="space-y-4">
          {[...history].reverse().map((entry, idx) => (
            <AnswerCard key={history.length - idx} entry={entry} />
          ))}
        </div>
      )}
    </div>
  );
}

function AnswerCard({ entry }: { entry: ChatEntry }) {
  const { question, response } = entry;
  return (
    <Card padding="md" className="space-y-3">
      <div className="flex items-start gap-2">
        <span className="inline-flex h-6 w-6 items-center justify-center rounded bg-surface-2 text-text-3 shrink-0 mt-0.5">
          <Icon name="User" size={12} />
        </span>
        <div className="text-[13px] text-text font-medium">{question}</div>
      </div>

      <div className="flex items-start gap-2">
        <span className="inline-flex h-6 w-6 items-center justify-center rounded bg-accent-soft text-accent-2 shrink-0 mt-0.5">
          <Icon name="Sparkles" size={12} />
        </span>
        <div className="flex-1 min-w-0 space-y-2">
          <div className="text-[13px] text-text whitespace-pre-wrap">
            {renderWithCitations(response.answer, response.citations.length)}
          </div>

          {response.citations.length > 0 && (
            <div className="space-y-1.5">
              <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
                Sources
              </div>
              <ol className="space-y-1.5">
                {response.citations.map((c) => (
                  <li key={c.index}>
                    <Link
                      href={`/dashboard/files/view/${c.file_id}`}
                      className="flex items-start gap-2 rounded border border-border bg-surface-2 px-2 py-1.5 hover:border-border-strong"
                    >
                      <Badge color="accent">[{c.index}]</Badge>
                      <FileIcon
                        mime={c.mime}
                        name={c.name}
                        size="sm"
                        className="shrink-0"
                      />
                      <div className="min-w-0 flex-1">
                        <div className="truncate text-[12px] font-medium text-text">
                          {c.name}
                        </div>
                        {c.snippet && (
                          <div className="text-[11px] text-text-3 line-clamp-2 mt-0.5">
                            {c.snippet}
                          </div>
                        )}
                      </div>
                      {typeof c.score === "number" && (
                        <span className="text-[10px] text-text-3 shrink-0 mt-0.5">
                          {(c.score * 100).toFixed(0)}%
                        </span>
                      )}
                    </Link>
                  </li>
                ))}
              </ol>
            </div>
          )}

          <div className="text-[10px] text-text-3">
            {response.provider} · {response.model} · {response.tokens_in ?? 0}→
            {response.tokens_out ?? 0} tok
            {response.latency_ms
              ? ` · ${(response.latency_ms / 1000).toFixed(1)}s`
              : ""}
          </div>
        </div>
      </div>
    </Card>
  );
}

// renderWithCitations highlights `[N]` tokens in the answer body so users can
// visually match them with the sources list.
function renderWithCitations(body: string, total: number): React.ReactNode {
  if (total <= 0) return body;
  const nodes: React.ReactNode[] = [];
  const re = /\[(\d+)\]/g;
  let last = 0;
  let m: RegExpExecArray | null;
  let key = 0;
  while ((m = re.exec(body))) {
    if (m.index > last) nodes.push(body.slice(last, m.index));
    nodes.push(
      <span
        key={key++}
        className="inline-flex items-center rounded bg-accent-soft text-accent-2 text-[11px] font-medium px-1 mx-0.5"
      >
        [{m[1]}]
      </span>,
    );
    last = m.index + m[0].length;
  }
  if (last < body.length) nodes.push(body.slice(last));
  return <>{nodes}</>;
}
