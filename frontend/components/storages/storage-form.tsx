"use client";

import { FormEvent, useState } from "react";
import { Button } from "@/components/ui/v2/button";
import { Input } from "@/components/ui/v2/input";
import { Checkbox } from "@/components/ui/v2/checkbox";
import { Icon } from "@/components/ui/v2/icon";
import { CreateStorageInput } from "@/lib/api/storages";
import { ApiError } from "@/lib/api/client";
import { cn } from "@/lib/utils";

interface Props {
  onSubmit: (input: CreateStorageInput) => Promise<void>;
  onCancel: () => void;
}

type Provider = "r2" | "s3" | "minio";

const providers: { value: Provider; label: string; hint: string }[] = [
  { value: "r2", label: "Cloudflare R2", hint: "<account>.r2.cloudflarestorage.com" },
  { value: "s3", label: "AWS S3", hint: "s3.us-east-1.amazonaws.com" },
  { value: "minio", label: "MinIO", hint: "http://minio:9000" },
];

export function StorageForm({ onSubmit, onCancel }: Props) {
  const [provider, setProvider] = useState<Provider>("r2");
  const [displayName, setDisplayName] = useState("");
  const [endpoint, setEndpoint] = useState("");
  const [region, setRegion] = useState("auto");
  const [bucket, setBucket] = useState("");
  const [accessKey, setAccessKey] = useState("");
  const [secretKey, setSecretKey] = useState("");
  const [forcePathStyle, setForcePathStyle] = useState(false);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  async function handle(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      await onSubmit({
        display_name: displayName,
        provider,
        endpoint,
        region,
        bucket,
        access_key: accessKey,
        secret_key: secretKey,
        force_path_style: forcePathStyle,
      });
    } catch (e) {
      setErr(e instanceof ApiError ? e.message : "Failed to add storage");
    } finally {
      setLoading(false);
    }
  }

  function onProviderChange(p: Provider) {
    setProvider(p);
    setForcePathStyle(p === "minio");
    if (p === "r2") setRegion("auto");
    else if (p === "s3") setRegion("us-east-1");
    else setRegion("");
  }

  const currentHint = providers.find((p) => p.value === provider)?.hint ?? "";

  return (
    <form id="storage-form" onSubmit={handle} className="space-y-4">
      <Labeled label="Provider">
        <div className="grid grid-cols-3 gap-1.5">
          {providers.map((p) => (
            <button
              key={p.value}
              type="button"
              onClick={() => onProviderChange(p.value)}
              className={cn(
                "flex flex-col items-start gap-0.5 rounded border px-2.5 py-2 text-left transition-colors",
                provider === p.value
                  ? "bg-accent-soft border-accent-border"
                  : "bg-surface border-border hover:border-border-strong",
              )}
            >
              <span className="flex items-center gap-1.5 text-[12px] font-medium text-text">
                <Icon
                  name="Cloud"
                  size={12}
                  className={provider === p.value ? "text-accent-2" : "text-text-3"}
                />
                {p.label}
              </span>
              <span className="text-[10px] text-text-3 truncate w-full">
                {p.hint}
              </span>
            </button>
          ))}
        </div>
      </Labeled>

      <Labeled label="Display name">
        <Input
          required
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          placeholder="My R2 bucket"
        />
      </Labeled>

      <Labeled label="Endpoint">
        <Input
          required
          value={endpoint}
          onChange={(e) => setEndpoint(e.target.value)}
          placeholder={`https://${currentHint}`}
        />
      </Labeled>

      <div className="grid grid-cols-2 gap-3">
        <Labeled label="Region">
          <Input value={region} onChange={(e) => setRegion(e.target.value)} />
        </Labeled>
        <Labeled label="Bucket">
          <Input
            required
            value={bucket}
            onChange={(e) => setBucket(e.target.value)}
            placeholder="my-bucket"
          />
        </Labeled>
      </div>

      <Labeled label="Access key">
        <Input
          required
          value={accessKey}
          onChange={(e) => setAccessKey(e.target.value)}
          autoComplete="off"
        />
      </Labeled>

      <Labeled label="Secret key">
        <Input
          required
          type="password"
          value={secretKey}
          onChange={(e) => setSecretKey(e.target.value)}
          autoComplete="off"
        />
      </Labeled>

      <label className="flex items-center gap-2 text-[12px] text-text-2">
        <Checkbox
          checked={forcePathStyle}
          onChange={(e) => setForcePathStyle(e.target.checked)}
        />
        Force path-style (MinIO / custom endpoints)
      </label>

      {err && <div className="text-[12px] text-danger">{err}</div>}

      <div className="flex justify-end gap-2 pt-1">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
        <Button type="submit" variant="accent" disabled={loading}>
          {loading ? "Testing + saving…" : "Add storage"}
        </Button>
      </div>
    </form>
  );
}

function Labeled({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1.5">
      <label className="block text-[12px] font-medium text-text-2">{label}</label>
      {children}
    </div>
  );
}
