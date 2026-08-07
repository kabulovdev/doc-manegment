"use client";

import { FormEvent, useState } from "react";
import { sharesApi, Share } from "@/lib/api/shares";
import { ApiError } from "@/lib/api/client";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/v2/button";
import { Checkbox } from "@/components/ui/v2/checkbox";
import { Dialog } from "@/components/ui/v2/dialog";
import { Icon } from "@/components/ui/v2/icon";
import { Input } from "@/components/ui/v2/input";

interface Props {
  open: boolean;
  onClose: () => void;
  targetType: "file" | "folder";
  targetID: string;
  targetName: string;
}

const expiryOptions = [
  { value: "1h", label: "1h" },
  { value: "24h", label: "24h" },
  { value: "7d", label: "7d" },
  { value: "30d", label: "30d" },
  { value: "none", label: "Never" },
] as const;

type Expiry = (typeof expiryOptions)[number]["value"];

const expiryMinutes: Record<Exclude<Expiry, "none">, number> = {
  "1h": 60,
  "24h": 60 * 24,
  "7d": 60 * 24 * 7,
  "30d": 60 * 24 * 30,
};

export function ShareDialog({
  open,
  onClose,
  targetType,
  targetID,
  targetName,
}: Props) {
  const [expiry, setExpiry] = useState<Expiry>("7d");
  const [password, setPassword] = useState("");
  const [oneTime, setOneTime] = useState(false);
  const [share, setShare] = useState<Share | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  async function handle(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      let expiresAt: string | null = null;
      if (expiry !== "none") {
        const mins = expiryMinutes[expiry];
        expiresAt = new Date(Date.now() + mins * 60 * 1000).toISOString();
      }
      const s = await sharesApi.create({
        target_type: targetType,
        target_id: targetID,
        expires_at: expiresAt,
        password: password || undefined,
        one_time_use: oneTime,
      });
      setShare(s);
    } catch (e) {
      setErr(e instanceof ApiError ? e.message : "Failed to create share");
    } finally {
      setLoading(false);
    }
  }

  function reset() {
    setShare(null);
    setPassword("");
    setExpiry("7d");
    setOneTime(false);
    setCopied(false);
    setErr(null);
  }

  function handleClose() {
    reset();
    onClose();
  }

  const shareURL = share
    ? `${typeof window !== "undefined" ? window.location.origin : ""}/share/${share.token}`
    : "";

  async function copy() {
    try {
      await navigator.clipboard.writeText(shareURL);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      alert(shareURL);
    }
  }

  if (share) {
    return (
      <Dialog
        open={open}
        onClose={handleClose}
        title={`Share ${targetType}`}
        description={`Link ready for ${targetName}.`}
        size="md"
        footer={
          <Button variant="secondary" onClick={handleClose}>
            Done
          </Button>
        }
      >
        <div className="space-y-3">
          <div className="flex gap-2">
            <Input
              readOnly
              value={shareURL}
              className="font-mono text-[11px]"
            />
            <Button type="button" variant="accent" onClick={copy}>
              <Icon name={copied ? "Check" : "Copy"} size={14} />
              {copied ? "Copied" : "Copy"}
            </Button>
          </div>
          <div className="flex gap-2 rounded border border-amber-200 bg-amber-50 p-3 text-[11px] text-amber-900">
            <Icon name="AlertTriangle" size={14} className="shrink-0 mt-0.5" />
            <span>
              The viewer can&rsquo;t download the file, but screenshots can&rsquo;t be
              prevented. Treat this as a deterrent, not a guarantee.
            </span>
          </div>
        </div>
      </Dialog>
    );
  }

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      title={`Share ${targetType}`}
      description={targetName}
      size="md"
      footer={
        <>
          <Button
            type="button"
            variant="secondary"
            onClick={handleClose}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            form="share-form"
            variant="accent"
            disabled={loading}
          >
            {loading ? "Creating…" : "Create link"}
          </Button>
        </>
      }
    >
      <form id="share-form" onSubmit={handle} className="space-y-4">
        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">
            Expires in
          </label>
          <div className="grid grid-cols-5 gap-1.5">
            {expiryOptions.map((k) => (
              <button
                key={k.value}
                type="button"
                onClick={() => setExpiry(k.value)}
                className={cn(
                  "rounded border h-8 text-[12px] transition-colors",
                  expiry === k.value
                    ? "bg-accent-soft border-accent-border text-accent-2 font-medium"
                    : "bg-surface border-border text-text-2 hover:border-border-strong",
                )}
              >
                {k.label}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-1.5">
          <label className="block text-[12px] font-medium text-text-2">
            Password (optional)
          </label>
          <Input
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Leave empty for no password"
          />
        </div>

        <label className="flex items-center gap-2 text-[12px] text-text-2">
          <Checkbox
            checked={oneTime}
            onChange={(e) => setOneTime(e.target.checked)}
          />
          One-time use (link expires after first open)
        </label>

        {err && <div className="text-[12px] text-danger">{err}</div>}
      </form>
    </Dialog>
  );
}
