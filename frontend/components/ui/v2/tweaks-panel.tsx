"use client";

import { useState } from "react";
import { Accent, Density, Theme, useTweaks } from "@/lib/tweaks/provider";
import { Icon } from "./icon";
import { cn } from "@/lib/utils";

const accentSwatches: Record<Accent, string> = {
  emerald: "#10b981",
  indigo: "#6366f1",
  amber: "#f59e0b",
  rose: "#f43f5e",
  slate: "#64748b",
};

const themes: Theme[] = ["light", "dark"];
const densities: Density[] = ["compact", "default", "comfortable"];
const accents: Accent[] = ["emerald", "indigo", "amber", "rose", "slate"];

export function TweaksPanel() {
  const [open, setOpen] = useState(false);
  const { theme, density, accent, setTheme, setDensity, setAccent } = useTweaks();

  return (
    <>
      <button
        type="button"
        aria-label="Open appearance tweaks"
        onClick={() => setOpen((v) => !v)}
        className="fixed bottom-4 right-4 z-40 inline-flex h-10 w-10 items-center justify-center rounded-full bg-surface border border-border text-text-2 shadow-md hover:text-text hover:border-border-strong"
      >
        <Icon name="settings" size={16} />
      </button>

      {open && (
        <div className="fixed inset-0 z-40" onClick={() => setOpen(false)}>
          <div
            role="dialog"
            aria-label="Appearance tweaks"
            onClick={(e) => e.stopPropagation()}
            className="absolute bottom-16 right-4 w-[280px] rounded-lg border border-border bg-surface shadow-lg p-3"
          >
            <div className="text-[11px] font-semibold uppercase tracking-wide text-text-3 mb-1">
              Theme
            </div>
            <div className="grid grid-cols-2 gap-1.5 mb-3">
              {themes.map((t) => (
                <button
                  key={t}
                  onClick={() => setTheme(t)}
                  className={cn(
                    "h-8 rounded border text-[12px] capitalize",
                    theme === t
                      ? "bg-accent-soft border-accent-border text-accent-2 font-medium"
                      : "bg-surface-2 border-border text-text-2 hover:text-text",
                  )}
                >
                  {t}
                </button>
              ))}
            </div>

            <div className="text-[11px] font-semibold uppercase tracking-wide text-text-3 mb-1">
              Density
            </div>
            <div className="grid grid-cols-3 gap-1.5 mb-3">
              {densities.map((d) => (
                <button
                  key={d}
                  onClick={() => setDensity(d)}
                  className={cn(
                    "h-8 rounded border text-[11px] capitalize",
                    density === d
                      ? "bg-accent-soft border-accent-border text-accent-2 font-medium"
                      : "bg-surface-2 border-border text-text-2 hover:text-text",
                  )}
                >
                  {d}
                </button>
              ))}
            </div>

            <div className="text-[11px] font-semibold uppercase tracking-wide text-text-3 mb-1">
              Accent
            </div>
            <div className="flex items-center gap-2">
              {accents.map((a) => (
                <button
                  key={a}
                  aria-label={`Accent ${a}`}
                  onClick={() => setAccent(a)}
                  className={cn(
                    "h-6 w-6 rounded-full border-2 transition-transform",
                    accent === a
                      ? "border-text scale-110"
                      : "border-transparent hover:scale-105",
                  )}
                  style={{ background: accentSwatches[a] }}
                />
              ))}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
