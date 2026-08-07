"use client";

import { ToastKind, useToastStore } from "@/lib/stores/toast-store";
import { cn } from "@/lib/utils";
import { Icon } from "./icon";

const kindStyles: Record<ToastKind, string> = {
  success: "bg-accent-soft border-accent-border text-accent-2",
  error: "bg-red-50 border-red-200 text-red-700",
  info: "bg-surface border-border text-text",
};

const kindIcons: Record<ToastKind, string> = {
  success: "CheckCircle2",
  error: "AlertCircle",
  info: "Info",
};

export function Toaster() {
  const { toasts, dismiss } = useToastStore();
  if (toasts.length === 0) return null;
  return (
    <div className="pointer-events-none fixed bottom-4 left-1/2 z-50 flex -translate-x-1/2 flex-col gap-1.5">
      {toasts.map((t) => (
        <div
          key={t.id}
          role="status"
          className={cn(
            "pointer-events-auto flex items-center gap-2 rounded-lg border px-3 py-2 text-[12px] shadow-md min-w-[220px] max-w-sm",
            kindStyles[t.kind],
          )}
        >
          <Icon name={kindIcons[t.kind]} size={14} />
          <span className="flex-1">{t.message}</span>
          <button
            type="button"
            onClick={() => dismiss(t.id)}
            aria-label="Dismiss"
            className="text-current opacity-60 hover:opacity-100"
          >
            <Icon name="X" size={12} />
          </button>
        </div>
      ))}
    </div>
  );
}
