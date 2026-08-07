"use client";

import { ReactNode, useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { DropdownItem } from "./dropdown-menu";

export interface ContextMenuProps {
  items: DropdownItem[];
  children: ReactNode;
  className?: string;
}

export function ContextMenu({ items, children, className }: ContextMenuProps) {
  const [pos, setPos] = useState<{ x: number; y: number } | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!pos) return;
    const close = () => setPos(null);
    document.addEventListener("click", close);
    document.addEventListener("contextmenu", close);
    return () => {
      document.removeEventListener("click", close);
      document.removeEventListener("contextmenu", close);
    };
  }, [pos]);

  return (
    <div
      className={className}
      onContextMenu={(e) => {
        e.preventDefault();
        setPos({ x: e.clientX, y: e.clientY });
      }}
    >
      {children}
      {pos && (
        <div
          ref={menuRef}
          role="menu"
          onContextMenu={(e) => e.preventDefault()}
          className="fixed z-50 min-w-[180px] rounded-lg border border-border bg-surface shadow-lg py-1"
          style={{ top: pos.y, left: pos.x }}
        >
          {items.map((it, i) =>
            it.separator ? (
              <div key={i} className="my-1 h-px bg-border" />
            ) : (
              <button
                key={i}
                role="menuitem"
                disabled={it.disabled}
                onClick={() => {
                  it.onSelect?.();
                  setPos(null);
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
