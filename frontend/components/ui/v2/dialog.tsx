"use client";

import { ReactNode, useEffect, useRef } from "react";
import { cn } from "@/lib/utils";

export type DialogSize = "sm" | "md" | "lg" | "full";

export interface DialogProps {
  open: boolean;
  onClose: () => void;
  title?: ReactNode;
  description?: ReactNode;
  size?: DialogSize;
  children?: ReactNode;
  footer?: ReactNode;
  closeOnBackdrop?: boolean;
}

const sizeMap: Record<DialogSize, string> = {
  sm: "max-w-sm",
  md: "max-w-lg",
  lg: "max-w-3xl",
  full: "max-w-[calc(100vw-32px)] h-[calc(100vh-32px)]",
};

export function Dialog({
  open,
  onClose,
  title,
  description,
  size = "md",
  children,
  footer,
  closeOnBackdrop = true,
}: DialogProps) {
  const panelRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<Element | null>(null);

  useEffect(() => {
    if (!open) return;
    triggerRef.current = document.activeElement;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKey);
    panelRef.current?.focus();
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = prev;
      if (triggerRef.current instanceof HTMLElement) triggerRef.current.focus();
    };
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget && closeOnBackdrop) onClose();
      }}
    >
      <div
        ref={panelRef}
        tabIndex={-1}
        role="dialog"
        aria-modal="true"
        className={cn(
          "w-full bg-surface border border-border rounded-lg shadow-lg flex flex-col outline-none",
          sizeMap[size],
        )}
      >
        {(title || description) && (
          <div className="px-4 pt-3 pb-3 border-b border-border">
            {title && <div className="text-[13px] font-semibold text-text">{title}</div>}
            {description && (
              <div className="text-xs text-text-2 mt-1">{description}</div>
            )}
          </div>
        )}
        <div className="flex-1 overflow-auto p-4">{children}</div>
        {footer && (
          <div className="px-4 py-3 border-t border-border flex items-center justify-end gap-2">
            {footer}
          </div>
        )}
      </div>
    </div>
  );
}
