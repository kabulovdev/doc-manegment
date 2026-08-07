"use client";

import { FormEvent, useEffect, useState } from "react";
import { authApi, ApiError } from "@/lib/api/client";
import { useAuthStore } from "@/lib/stores/auth-store";
import { toast } from "@/lib/stores/toast-store";
import { Avatar } from "@/components/ui/v2/avatar";
import { Badge } from "@/components/ui/v2/badge";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { Input } from "@/components/ui/v2/input";
import { Select } from "@/components/ui/v2/select";

export function ProfileTab() {
  const user = useAuthStore((s) => s.user);
  const accessToken = useAuthStore((s) => s.accessToken);
  const setAuth = useAuthStore((s) => s.setAuth);

  const [displayName, setDisplayName] = useState(user?.display_name ?? "");
  const [language, setLanguage] = useState("en");
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    setDisplayName(user?.display_name ?? "");
  }, [user?.display_name]);

  const dirty = displayName.trim() !== (user?.display_name ?? "");

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!dirty) return;
    setErr(null);
    setSaving(true);
    try {
      const updated = await authApi.updateMe({ display_name: displayName.trim() });
      if (accessToken) setAuth(accessToken, updated);
      toast("Profile updated", "success");
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : "Failed to update profile";
      setErr(msg);
      toast(msg, "error");
    } finally {
      setSaving(false);
    }
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <Card title="Profile">
        <div className="flex items-start gap-4">
          <Avatar
            name={user?.display_name || user?.email || "?"}
            size="lg"
          />
          <div className="flex-1 grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-3 min-w-0">
            <Field label="Display name">
              <Input
                value={displayName}
                maxLength={100}
                onChange={(e) => setDisplayName(e.target.value)}
              />
            </Field>
            <Field label="Email">
              <Input value={user?.email ?? ""} readOnly className="opacity-80" />
            </Field>
            <Field label="Role">
              <div>
                <Badge color="accent">Owner</Badge>
              </div>
            </Field>
            <Field label="Language">
              <Select
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
              >
                <option value="en">English</option>
                <option value="uz" disabled>
                  O&rsquo;zbek — coming soon
                </option>
                <option value="ru" disabled>
                  Русский — coming soon
                </option>
              </Select>
            </Field>
          </div>
        </div>
      </Card>

      {err && <div className="text-[12px] text-danger">{err}</div>}

      <div className="flex items-center justify-end gap-2">
        <span className="text-[11px] text-text-3 mr-auto">
          {dirty ? "Unsaved changes" : "All changes saved"}
        </span>
        <Button
          type="button"
          variant="ghost"
          disabled={!dirty || saving}
          onClick={() => setDisplayName(user?.display_name ?? "")}
        >
          Discard
        </Button>
        <Button
          type="submit"
          variant="accent"
          disabled={!dirty || saving || !displayName.trim()}
        >
          {saving ? "Saving…" : "Save"}
        </Button>
      </div>
    </form>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="min-w-0 space-y-1.5">
      <div className="text-[10px] font-semibold uppercase tracking-wide text-text-3">
        {label}
      </div>
      {children}
    </div>
  );
}
