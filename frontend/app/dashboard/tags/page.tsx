"use client";

import { FormEvent, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Tag, tagsApi } from "@/lib/api/tags";
import { usePageTitle } from "@/lib/stores/page-title-store";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/v2/button";
import { Card } from "@/components/ui/v2/card";
import { EmptyState } from "@/components/ui/v2/empty-state";
import { Icon } from "@/components/ui/v2/icon";
import { IconButton } from "@/components/ui/v2/icon-button";
import { Input } from "@/components/ui/v2/input";
import { Pill, PillColor } from "@/components/ui/v2/pill";
import { SectionHead } from "@/components/ui/v2/section-head";
import { Table, THead, THCell, TRow, TCell } from "@/components/ui/v2/table";

const PRESET_COLORS = [
  "#10b981",
  "#6366f1",
  "#f59e0b",
  "#f43f5e",
  "#0ea5e9",
  "#8b5cf6",
  "#64748b",
  "#0f172a",
];

const pillColorFor = (hex: string): PillColor => {
  const h = hex.toLowerCase();
  if (h === "#10b981") return "emerald";
  if (h === "#6366f1") return "indigo";
  if (h === "#f59e0b") return "amber";
  if (h === "#f43f5e") return "rose";
  if (h === "#0ea5e9") return "sky";
  if (h === "#8b5cf6") return "violet";
  return "slate";
};

export default function TagsPage() {
  usePageTitle("Tags", "Organize files with labels");
  const qc = useQueryClient();

  const { data: tags = [], isLoading } = useQuery<Tag[]>({
    queryKey: ["tags"],
    queryFn: () => tagsApi.list(),
  });

  const [name, setName] = useState("");
  const [color, setColor] = useState(PRESET_COLORS[0]);
  const [filter, setFilter] = useState("");
  const [editID, setEditID] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [editColor, setEditColor] = useState("");

  const create = useMutation({
    mutationFn: () => tagsApi.create(name.trim(), color),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tags"] });
      setName("");
    },
  });
  const update = useMutation({
    mutationFn: ({ id, ...patch }: { id: string; name?: string; color?: string }) =>
      tagsApi.update(id, patch),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tags"] });
      qc.invalidateQueries({ queryKey: ["files"] });
      setEditID(null);
    },
  });
  const del = useMutation({
    mutationFn: (id: string) => tagsApi.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tags"] });
      qc.invalidateQueries({ queryKey: ["files"] });
    },
  });

  const filtered = useMemo(() => {
    const q = filter.trim().toLowerCase();
    if (!q) return tags;
    return tags.filter((t) => t.name.toLowerCase().includes(q));
  }, [tags, filter]);

  function onCreate(e: FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    create.mutate();
  }

  function beginEdit(t: Tag) {
    setEditID(t.id);
    setEditName(t.name);
    setEditColor(t.color || PRESET_COLORS[0]);
  }

  function saveEdit(t: Tag) {
    const patch: { id: string; name?: string; color?: string } = { id: t.id };
    if (editName.trim() && editName.trim() !== t.name) patch.name = editName.trim();
    if (editColor !== t.color) patch.color = editColor;
    if (patch.name === undefined && patch.color === undefined) {
      setEditID(null);
      return;
    }
    update.mutate(patch);
  }

  return (
    <div className="space-y-6">
      <SectionHead
        level="h1"
        title="Tags"
        subtitle="Create tags and attach them to files or folders."
      />

      <Card title="Create tag">
        <form onSubmit={onCreate} className="flex flex-wrap items-end gap-3">
          <div className="min-w-[200px] flex-1 space-y-1.5">
            <label className="block text-[12px] font-medium text-text-2">Name</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="passport, invoice…"
            />
          </div>
          <div className="space-y-1.5">
            <label className="block text-[12px] font-medium text-text-2">Color</label>
            <ColorPicker value={color} onChange={setColor} />
          </div>
          <Button
            type="submit"
            variant="accent"
            disabled={create.isPending || !name.trim()}
          >
            <Icon name="Plus" size={14} /> Create
          </Button>
        </form>
        {create.isError && (
          <div className="text-[12px] text-danger mt-2">
            {(create.error as Error).message || "Failed to create tag"}
          </div>
        )}
      </Card>

      <div className="flex items-center justify-between gap-3">
        <div className="relative w-full max-w-xs">
          <Icon
            name="Search"
            size={14}
            className="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-3"
          />
          <Input
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="Filter tags…"
            className="pl-8"
          />
        </div>
        <div className="text-[11px] text-text-3">
          {tags.length} tag{tags.length === 1 ? "" : "s"}
        </div>
      </div>

      {isLoading ? (
        <Card><div className="text-[12px] text-text-3">Loading…</div></Card>
      ) : tags.length === 0 ? (
        <Card>
          <EmptyState
            icon={<Icon name="Tag" size={18} />}
            title="No tags yet"
            description="Create a tag above to start labeling files."
          />
        </Card>
      ) : filtered.length === 0 ? (
        <Card>
          <div className="py-6 text-center text-[12px] text-text-3">
            No tags match “{filter}”.
          </div>
        </Card>
      ) : (
        <Card padding="none">
          <div className="overflow-x-auto">
            <Table>
              <THead>
                <tr>
                  <THCell>Tag</THCell>
                  <THCell className="w-32">Color</THCell>
                  <THCell className="w-20">Files</THCell>
                  <THCell className="w-32">Created</THCell>
                  <THCell className="w-24 text-right">Actions</THCell>
                </tr>
              </THead>
              <tbody>
                {filtered.map((t) => {
                  const editing = editID === t.id;
                  return (
                    <TRow key={t.id}>
                      <TCell>
                        {editing ? (
                          <Input
                            autoFocus
                            value={editName}
                            onChange={(e) => setEditName(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === "Enter") saveEdit(t);
                              if (e.key === "Escape") setEditID(null);
                            }}
                            className="max-w-[220px]"
                          />
                        ) : (
                          <Pill name={t.name} color={pillColorFor(t.color)} />
                        )}
                      </TCell>
                      <TCell>
                        {editing ? (
                          <ColorPicker value={editColor} onChange={setEditColor} compact />
                        ) : (
                          <span className="inline-flex items-center gap-1.5">
                            <span
                              className="h-3 w-3 rounded-full border border-border"
                              style={{ background: t.color || "transparent" }}
                            />
                            <span className="font-mono text-[11px] text-text-2">
                              {t.color || "—"}
                            </span>
                          </span>
                        )}
                      </TCell>
                      <TCell className="text-text-3">—</TCell>
                      <TCell className="text-text-3 whitespace-nowrap">
                        {new Date(t.created_at).toLocaleDateString()}
                      </TCell>
                      <TCell>
                        <div className="flex items-center justify-end gap-1">
                          {editing ? (
                            <>
                              <Button
                                size="sm"
                                variant="accent"
                                onClick={() => saveEdit(t)}
                                disabled={update.isPending}
                              >
                                Save
                              </Button>
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => setEditID(null)}
                              >
                                Cancel
                              </Button>
                            </>
                          ) : (
                            <>
                              <IconButton
                                size="sm"
                                aria-label="Edit"
                                onClick={() => beginEdit(t)}
                              >
                                <Icon name="Pencil" size={12} />
                              </IconButton>
                              <IconButton
                                size="sm"
                                aria-label="Delete"
                                onClick={() => {
                                  if (confirm(`Delete tag "${t.name}"?`)) del.mutate(t.id);
                                }}
                              >
                                <Icon name="Trash2" size={12} />
                              </IconButton>
                            </>
                          )}
                        </div>
                      </TCell>
                    </TRow>
                  );
                })}
              </tbody>
            </Table>
          </div>
        </Card>
      )}
    </div>
  );
}

function ColorPicker({
  value,
  onChange,
  compact,
}: {
  value: string;
  onChange: (v: string) => void;
  compact?: boolean;
}) {
  return (
    <div className="flex items-center gap-1">
      {PRESET_COLORS.map((c) => (
        <button
          key={c}
          type="button"
          onClick={() => onChange(c)}
          aria-label={`Color ${c}`}
          className={cn(
            "rounded-full border-2 transition-all",
            compact ? "h-5 w-5" : "h-7 w-7",
            value.toLowerCase() === c.toLowerCase()
              ? "border-text scale-110"
              : "border-transparent hover:scale-105",
          )}
          style={{ background: c }}
        />
      ))}
      <label
        className={cn(
          "relative inline-flex items-center justify-center rounded-full border border-border bg-surface-2 text-text-3 cursor-pointer",
          compact ? "h-5 w-5" : "h-7 w-7",
        )}
        aria-label="Custom color"
      >
        <Icon name="Palette" size={compact ? 10 : 12} />
        <input
          type="color"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="absolute inset-0 opacity-0 cursor-pointer"
        />
      </label>
    </div>
  );
}
