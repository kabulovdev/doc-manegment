"use client";

import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ApiError } from "@/lib/api/client";
import { ApiToken, CreateApiTokenResponse, tokensApi } from "@/lib/api/tokens";
import { toast } from "@/lib/stores/toast-store";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/v2/badge";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { Checkbox } from "@/components/ui/v2/checkbox";
import { Dialog } from "@/components/ui/v2/dialog";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { Icon } from "@/components/ui/v2/icon";
import { IconButton } from "@/components/ui/v2/icon-button";
import { Input } from "@/components/ui/v2/input";
import { Tooltip } from "@/components/ui/v2/tooltip";
import { Table, THead, THCell, TRow, TCell } from "@/components/ui/v2/table";

function relTime(iso?: string | null) {
  if (!iso) return "—";
  const d = new Date(iso);
  const diff = (Date.now() - d.getTime()) / 1000;
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 86400 * 7) return `${Math.floor(diff / 86400)}d ago`;
  return d.toLocaleDateString();
}

export function ApiTokensTab() {
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [justCreated, setJustCreated] = useState<CreateApiTokenResponse | null>(null);

  const { data: tokens = [], isLoading } = useQuery<ApiToken[]>({
    queryKey: ["api-tokens"],
    queryFn: () => tokensApi.list(),
  });

  const revoke = useMutation({
    mutationFn: (id: string) => tokensApi.revoke(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["api-tokens"] });
      toast("Token revoked", "success");
    },
    onError: (e: unknown) => {
      const msg = e instanceof ApiError ? e.message : "Revoke failed";
      toast(msg, "error");
    },
  });

  return (
    <div className="space-y-4">
      <Card
        title="API tokens"
        subtitle="Use these tokens to authenticate CLI or integration requests."
        action={
          <Button variant="accent" size="sm" onClick={() => setCreateOpen(true)}>
            <Icon name="Plus" size={12} /> New token
          </Button>
        }
        padding={tokens.length === 0 ? "md" : "none"}
      >
        {isLoading ? (
          <div className="text-[12px] text-text-3">Loading…</div>
        ) : tokens.length === 0 ? (
          <EmptyState
            icon={<Icon name="Key" size={18} />}
            title="No tokens yet"
            description="Create a token to call the API from scripts or external apps."
          />
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <THead>
                <tr>
                  <THCell>Name</THCell>
                  <THCell className="w-40">Token</THCell>
                  <THCell>Scopes</THCell>
                  <THCell className="w-28">Last used</THCell>
                  <THCell className="w-28">Created</THCell>
                  <THCell className="w-24">Status</THCell>
                  <THCell className="w-14" />
                </tr>
              </THead>
              <tbody>
                {tokens.map((t) => {
                  const revoked = !!t.revoked_at;
                  return (
                    <TRow key={t.id}>
                      <TCell>
                        <div className="font-medium text-text truncate">{t.name}</div>
                      </TCell>
                      <TCell>
                        <span className="font-mono text-[11px] text-text-2">
                          doc_{t.prefix}{"·".repeat(8)}
                        </span>
                      </TCell>
                      <TCell>
                        <div className="flex flex-wrap gap-1">
                          {t.scopes.slice(0, 4).map((s) => (
                            <Badge key={s} color="slate">
                              {s}
                            </Badge>
                          ))}
                          {t.scopes.length > 4 && (
                            <span className="text-[10px] text-text-3">
                              +{t.scopes.length - 4}
                            </span>
                          )}
                        </div>
                      </TCell>
                      <TCell className="text-text-3 whitespace-nowrap">
                        {relTime(t.last_used_at)}
                      </TCell>
                      <TCell className="text-text-3 whitespace-nowrap">
                        {relTime(t.created_at)}
                      </TCell>
                      <TCell>
                        <Badge color={revoked ? "danger" : "accent"} dot>
                          {revoked ? "Revoked" : "Active"}
                        </Badge>
                      </TCell>
                      <TCell className="text-right">
                        {!revoked && (
                          <Tooltip content="Revoke">
                            <IconButton
                              size="sm"
                              aria-label="Revoke"
                              onClick={() => {
                                if (confirm(`Revoke token "${t.name}"?`))
                                  revoke.mutate(t.id);
                              }}
                            >
                              <Icon name="Trash2" size={12} />
                            </IconButton>
                          </Tooltip>
                        )}
                      </TCell>
                    </TRow>
                  );
                })}
              </tbody>
            </Table>
          </div>
        )}
      </Card>

      <CreateTokenDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={(res) => {
          setCreateOpen(false);
          setJustCreated(res);
          qc.invalidateQueries({ queryKey: ["api-tokens"] });
        }}
      />

      <TokenRevealDialog
        response={justCreated}
        onClose={() => setJustCreated(null)}
      />
    </div>
  );
}

function CreateTokenDialog({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: (res: CreateApiTokenResponse) => void;
}) {
  const [name, setName] = useState("");
  const [scopes, setScopes] = useState<string[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const { data: available = [] } = useQuery<string[]>({
    queryKey: ["token-scopes"],
    queryFn: () => tokensApi.scopes(),
    enabled: open,
  });

  function reset() {
    setName("");
    setScopes([]);
    setErr(null);
    setLoading(false);
  }

  function toggle(s: string) {
    setScopes((cur) =>
      cur.includes(s) ? cur.filter((x) => x !== s) : [...cur, s],
    );
  }

  async function handle(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      const res = await tokensApi.create({
        name: name.trim(),
        scopes: scopes.length > 0 ? scopes : undefined,
      });
      reset();
      onCreated(res);
    } catch (e) {
      setErr(e instanceof ApiError ? e.message : "Failed to create token");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog
      open={open}
      onClose={() => {
        if (loading) return;
        reset();
        onClose();
      }}
      title="New API token"
      description="The plaintext token is shown once after creation."
      size="md"
      footer={
        <>
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              reset();
              onClose();
            }}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            form="create-token-form"
            variant="accent"
            disabled={loading || !name.trim()}
          >
            {loading ? "Creating…" : "Create token"}
          </Button>
        </>
      }
    >
      <form id="create-token-form" onSubmit={handle} className="space-y-4">
        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">Name</label>
          <Input
            autoFocus
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="CLI, integration-foo, …"
            maxLength={100}
          />
        </div>
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <label className="text-[12px] font-medium text-text-2">Scopes</label>
            <div className="flex items-center gap-1.5 text-[11px] text-text-3">
              <button
                type="button"
                className="hover:text-text"
                onClick={() => setScopes(available)}
              >
                Select all
              </button>
              <span>·</span>
              <button
                type="button"
                className="hover:text-text"
                onClick={() => setScopes([])}
              >
                Clear
              </button>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-x-3 gap-y-1.5">
            {available.map((s) => (
              <label
                key={s}
                className={cn(
                  "flex items-center gap-2 text-[12px] cursor-pointer",
                  scopes.includes(s) ? "text-text" : "text-text-2",
                )}
              >
                <Checkbox
                  checked={scopes.includes(s)}
                  onChange={() => toggle(s)}
                />
                <span className="font-mono">{s}</span>
              </label>
            ))}
          </div>
          <p className="text-[11px] text-text-3">
            Leave empty to grant all scopes. Scopes are stored now and will be
            enforced per-request in a future release.
          </p>
        </div>
        {err && <div className="text-[12px] text-danger">{err}</div>}
      </form>
    </Dialog>
  );
}

function TokenRevealDialog({
  response,
  onClose,
}: {
  response: CreateApiTokenResponse | null;
  onClose: () => void;
}) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    if (!response) return;
    try {
      await navigator.clipboard.writeText(response.plaintext);
      setCopied(true);
      toast("Token copied", "success");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      alert(response.plaintext);
    }
  }

  return (
    <Dialog
      open={!!response}
      onClose={onClose}
      title="Copy your new token"
      description="This is the only time we'll show it. Store it somewhere safe."
      size="md"
      footer={
        <Button variant="accent" onClick={onClose}>
          I&rsquo;ve saved it
        </Button>
      }
    >
      {response && (
        <div className="space-y-3">
          <div className="space-y-1">
            <label className="block text-[11px] font-semibold uppercase tracking-wide text-text-3">
              Name
            </label>
            <div className="text-[13px] text-text font-medium">
              {response.token.name}
            </div>
          </div>
          <div className="space-y-1">
            <label className="block text-[11px] font-semibold uppercase tracking-wide text-text-3">
              Token
            </label>
            <div className="flex gap-2">
              <Input
                readOnly
                value={response.plaintext}
                className="font-mono text-[12px]"
              />
              <Button variant="secondary" onClick={copy}>
                <Icon name={copied ? "Check" : "Copy"} size={14} />
                {copied ? "Copied" : "Copy"}
              </Button>
            </div>
          </div>
          <div className="flex gap-2 rounded border border-amber-200 bg-amber-50 p-3 text-[11px] text-amber-900">
            <Icon name="AlertTriangle" size={14} className="shrink-0 mt-0.5" />
            <span>
              Treat this token like a password. Anyone who has it can act as
              you. Revoke it if you suspect exposure.
            </span>
          </div>
        </div>
      )}
    </Dialog>
  );
}
