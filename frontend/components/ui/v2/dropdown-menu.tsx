"use client";

import { ReactNode, useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";

export interface DropdownItem {
  label: ReactNode;
  icon?: ReactNode;
  onSelect?: () => void;
  disabled?: boolean;
  danger?: boolean;
  separator?: boolean;
}

export interface DropdownMenuProps {
  trigger: (props: { open: boolean; toggle: () => void }) => ReactNode;
  items: DropdownItem[];
  align?: "start" | "end";
  className?: string;
}

export function DropdownMenu({
  trigger,
  items,
  align = "start",
  className,
}: DropdownMenuProps) {
  const [open, setOpen] = useState(false);
  const [focus, setFocus] = useState(-1);
  const rootRef = useRef<HTMLDivElement>(null);
  const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);

  useEffect(() => {
    if (!open) return;
    const onClickOutside = (e: MouseEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setFocus((f) => Math.min(items.length - 1, f + 1));
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        setFocus((f) => Math.max(0, f - 1));
      }
    };
    document.addEventListener("mousedown", onClickOutside);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onClickOutside);
      document.removeEventListener("keydown", onKey);
    };
  }, [open, items.length]);

  useEffect(() => {
    if (open && focus >= 0) itemRefs.current[focus]?.focus();
  }, [open, focus]);

  const toggle = () => {
    setOpen((v) => !v);
    setFocus(-1);
  };

  return (
    <div ref={rootRef} className={cn("relative inline-block", className)}>
      {trigger({ open, toggle })}
      {open && (
        <div
          role="menu"
          className={cn(
            "absolute z-40 mt-1 min-w-[180px] rounded-lg border border-border bg-surface shadow-lg py-1",
            align === "end" ? "right-0" : "left-0",
          )}
        >
          {items.map((it, i) =>
            it.separator ? (
              <div key={i} className="my-1 h-px bg-border" />
            ) : (
              <button
                key={i}
                ref={(el) => {
                  itemRefs.current[i] = el;
                }}
                role="menuitem"
                disabled={it.disabled}
                onClick={() => {
                  it.onSelect?.();
                  setOpen(false);
                }}
                className={cn(
                  "flex w-full items-center gap-2 px-2.5 h-8 text-[13px] text-left",
                  "hover:bg-surface-2 focus:bg-surface-2 focus:outline-none",
                  it.danger ? "text-danger" : "text-text",
                  it.disabled && "opacity-50 cursor-not-allowed",
                )}
              >
                {it.icon && <span className="text-text-3">{it.icon}</span>}
                <span className="flex-1 truncate">{it.label}</span>
              </button>
            ),
          )}
        </div>
      )}
    </div>
  );
}
