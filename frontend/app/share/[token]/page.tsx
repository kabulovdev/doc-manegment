"use client";

import { useParams } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import dynamic from "next/dynamic";
import { ImageViewer } from "@/components/viewer/image-viewer";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { FileIcon } from "@/components/ui/v2/file-icon";
import { Icon } from "@/components/ui/v2/icon";
import { Input } from "@/components/ui/v2/input";

const PDFViewer = dynamic(
  () => import("@/components/viewer/pdf-viewer").then((m) => m.PDFViewer),
  {
    ssr: false,
    loading: () => (
      <div className="p-6 text-[12px] text-text-3">Loading PDF viewer…</div>
    ),
  },
);

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080/api/v1";

interface Metadata {
  target_type: "file" | "folder";
  requires_password: boolean;
  unlocked: boolean;
  name?: string;
  mime_type?: string;
  size_bytes?: number;
  folder_id?: string;
}

interface FolderListing {
  folder?: { id: string; name: string };
  children: { id: string; name: string }[];
  files: { id: string; name: string; mime_type: string; size_bytes: number }[];
}

async function fetchMeta(
  token: string,
): Promise<{ status: number; body: Metadata | null }> {
  const res = await fetch(`${API_BASE}/share/${token}`, {
    credentials: "include",
  });
  if (!res.ok) return { status: res.status, body: null };
  return { status: res.status, body: await res.json() };
}

export default function PublicSharePage() {
  const params = useParams<{ token: string }>();
  const token = params.token;
  const [meta, setMeta] = useState<Metadata | null>(null);
  const [status, setStatus] = useState<number>(0);
  const [password, setPassword] = useState("");
  const [unlockErr, setUnlockErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  async function reload() {
    setLoading(true);
    const { status, body } = await fetchMeta(token);
    setStatus(status);
    setMeta(body);
    setLoading(false);
  }

  useEffect(() => {
    reload();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  async function onUnlock(e: FormEvent) {
    e.preventDefault();
    setUnlockErr(null);
    const res = await fetch(`${API_BASE}/share/${token}/unlock`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ password }),
    });
    if (!res.ok) {
      setUnlockErr("Wrong password or link no longer available.");
      return;
    }
    setPassword("");
    await reload();
  }

  if (loading) {
    return (
      <Shell>
        <div className="text-[12px] text-text-3">Loading…</div>
      </Shell>
    );
  }

  if (status === 410) {
    return (
      <Shell>
        <Card className="mx-auto max-w-sm">
          <EmptyState
            icon={<Icon name="ShieldAlert" size={20} className="text-warn" />}
            title="Link no longer available"
            description="This link has expired, been revoked, or was already consumed."
          />
        </Card>
      </Shell>
    );
  }
  if (status === 404 || !meta) {
    return (
      <Shell>
        <Card className="mx-auto max-w-sm">
          <EmptyState
            icon={<Icon name="ShieldAlert" size={20} className="text-danger" />}
            title="Not found"
          />
        </Card>
      </Shell>
    );
  }

  if (meta.requires_password && !meta.unlocked) {
    return (
      <Shell>
        <Card className="mx-auto w-full max-w-sm" padding="lg">
          <div className="flex flex-col items-center text-center mb-4">
            <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-surface-2 text-text-2">
              <Icon name="Lock" size={18} />
            </span>
            <h1 className="mt-2 text-[15px] font-semibold text-text">
              Password required
            </h1>
            <p className="text-[12px] text-text-2 mt-1">
              Enter the password to view {meta.name ?? "this share"}.
            </p>
          </div>
          <form onSubmit={onUnlock} className="space-y-3">
            <Input
              autoFocus
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Password"
            />
            {unlockErr && (
              <div className="text-[12px] text-danger">{unlockErr}</div>
            )}
            <Button
              type="submit"
              variant="accent"
              className="w-full"
              disabled={!password}
            >
              Unlock
            </Button>
          </form>
        </Card>
      </Shell>
    );
  }

  if (meta.target_type === "file") {
    return (
      <FileShareView
        token={token}
        name={meta.name ?? "file"}
        mime={meta.mime_type ?? ""}
      />
    );
  }

  return (
    <FolderShareView
      token={token}
      rootID={meta.folder_id ?? ""}
      rootName={meta.name ?? "folder"}
    />
  );
}

function FileShareView({
  token,
  name,
  mime,
}: {
  token: string;
  name: string;
  mime: string;
}) {
  const url = `${API_BASE}/share/${token}/content`;
  return (
    <Shell>
      <Warning />
      <div className="flex items-center gap-3 mb-4">
        <FileIcon mime={mime} name={name} size="lg" />
        <div className="min-w-0">
          <h1 className="text-[15px] font-semibold text-text truncate">{name}</h1>
          <p className="text-[11px] text-text-3">{mime}</p>
        </div>
      </div>
      <FileRenderer url={url} mime={mime} name={name} />
    </Shell>
  );
}

function FolderShareView({
  token,
  rootID,
  rootName,
}: {
  token: string;
  rootID: string;
  rootName: string;
}) {
  const [currentID, setCurrentID] = useState(rootID);
  const [listing, setListing] = useState<FolderListing | null>(null);
  const [selected, setSelected] = useState<
    { id: string; name: string; mime: string } | null
  >(null);
  const [stack, setStack] = useState<{ id: string; name: string }[]>([
    { id: rootID, name: rootName },
  ]);

  useEffect(() => {
    (async () => {
      const res = await fetch(
        `${API_BASE}/share/${token}/children?folder_id=${currentID}`,
        { credentials: "include" },
      );
      if (res.ok) setListing(await res.json());
    })();
  }, [token, currentID]);

  function enterFolder(id: string, name: string) {
    setStack((s) => [...s, { id, name }]);
    setCurrentID(id);
    setSelected(null);
  }

  function jumpTo(idx: number) {
    const newStack = stack.slice(0, idx + 1);
    setStack(newStack);
    setCurrentID(newStack[newStack.length - 1].id);
    setSelected(null);
  }

  if (selected) {
    return (
      <Shell>
        <Warning />
        <button
          onClick={() => setSelected(null)}
          className="mb-3 inline-flex items-center gap-1 text-[12px] text-text-2 hover:text-text"
        >
          <Icon name="ArrowLeft" size={12} /> Back to folder
        </button>
        <div className="flex items-center gap-3 mb-4">
          <FileIcon mime={selected.mime} name={selected.name} size="lg" />
          <div className="min-w-0">
            <h1 className="text-[15px] font-semibold text-text truncate">
              {selected.name}
            </h1>
            <p className="text-[11px] text-text-3">{selected.mime}</p>
          </div>
        </div>
        <FileRenderer
          url={`${API_BASE}/share/${token}/content?file_id=${selected.id}`}
          mime={selected.mime}
          name={selected.name}
        />
      </Shell>
    );
  }

  return (
    <Shell>
      <Warning />
      <nav className="flex items-center gap-1 text-[12px] text-text-3 mb-4 flex-wrap">
        {stack.map((c, i) => (
          <span key={c.id} className="flex items-center gap-1">
            {i > 0 && <Icon name="ChevronRight" size={10} />}
            <button
              onClick={() => jumpTo(i)}
              className={
                i === stack.length - 1
                  ? "text-text font-medium"
                  : "hover:text-text"
              }
            >
              {c.name}
            </button>
          </span>
        ))}
      </nav>
      <Card padding="none">
        {listing?.children.map((ch) => (
          <button
            key={ch.id}
            onClick={() => enterFolder(ch.id, ch.name)}
            className="flex items-center gap-2 w-full text-left px-3 py-2 border-b border-border hover:bg-surface-2 text-[13px]"
          >
            <Icon name="Folder" size={14} className="text-text-3" />
            <span className="truncate">{ch.name}</span>
          </button>
        ))}
        {listing?.files.map((f) => (
          <button
            key={f.id}
            onClick={() =>
              setSelected({ id: f.id, name: f.name, mime: f.mime_type })
            }
            className="flex items-center gap-2 w-full text-left px-3 py-2 border-b border-border hover:bg-surface-2 text-[13px]"
          >
            <FileIcon mime={f.mime_type} name={f.name} size="sm" />
            <span className="flex-1 truncate">{f.name}</span>
            <span className="text-[11px] text-text-3">{f.mime_type}</span>
          </button>
        ))}
        {listing && listing.children.length === 0 && listing.files.length === 0 && (
          <div className="px-4 py-8 text-center text-[12px] text-text-3">
            Folder is empty.
          </div>
        )}
      </Card>
    </Shell>
  );
}

function FileRenderer({
  url,
  mime,
  name,
}: {
  url: string;
  mime: string;
  name: string;
}) {
  if (mime === "application/pdf") return <PDFViewer url={url} withCredentials />;
  if (mime.startsWith("image/"))
    return <ImageViewer url={url} alt={name} withCredentials />;
  return (
    <Card>
      <EmptyState
        icon={<Icon name="FileX" size={18} />}
        title="Preview not available"
        description={`Only PDF and images can be previewed view-only (mime: ${mime}).`}
      />
    </Card>
  );
}

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-bg py-8 px-4">
      <div className="mx-auto max-w-5xl">{children}</div>
    </div>
  );
}

function Warning() {
  return (
    <div className="mx-auto mb-4 max-w-2xl rounded border border-amber-200 bg-amber-50 px-3 py-2 text-[11px] text-amber-900 flex items-center gap-2">
      <Icon name="AlertTriangle" size={14} className="shrink-0" />
      <span>
        This document is view-only. Downloading is disabled, but screenshots
        can&rsquo;t be prevented. Please treat the content as confidential.
      </span>
    </div>
  );
}
