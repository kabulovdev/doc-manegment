"use client";

import { FormEvent, useState } from "react";
import { Dialog } from "@/components/ui/v2/dialog";
import { Button } from "@/components/ui/v2/button";
import { Input } from "@/components/ui/v2/input";

interface Props {
  open: boolean;
  onClose: () => void;
  onCreate: (name: string) => Promise<void>;
}

export function NewFolderDialog({ open, onClose, onCreate }: Props) {
  const [name, setName] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handle(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      await onCreate(name.trim());
      setName("");
      onClose();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="New folder"
      footer={
        <>
          <Button
            type="button"
            variant="secondary"
            onClick={onClose}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            form="new-folder-form"
            variant="accent"
            disabled={loading || !name.trim()}
          >
            Create
          </Button>
        </>
      }
    >
      <form id="new-folder-form" onSubmit={handle} className="space-y-3">
        <div className="space-y-1.5">
          <label className="text-[12px] font-medium text-text-2">Name</label>
          <Input
            autoFocus
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="2026 taxes"
          />
        </div>
        {err && <div className="text-[12px] text-danger">{err}</div>}
      </form>
    </Dialog>
  );
}
