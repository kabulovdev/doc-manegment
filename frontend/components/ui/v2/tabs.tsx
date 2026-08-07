"use client";

import { ReactNode } from "react";
import { cn } from "@/lib/utils";

export interface TabItem {
  value: string;
  label: ReactNode;
  icon?: ReactNode;
  badge?: ReactNode;
}

export interface TabsProps {
  value: string;
  onChange: (value: string) => void;
  items: TabItem[];
  className?: string;
}

export function Tabs({ value, onChange, items, className }: TabsProps) {
  return (
    <div
      role="tablist"
      className={cn(
        "flex items-center gap-0 border-b border-border overflow-x-auto",
        className,
      )}
    >
      {items.map((it) => {
        const active = it.value === value;
        return (
          <button
            key={it.value}
            role="tab"
            aria-selected={active}
            onClick={() => onChange(it.value)}
            className={cn(
              "inline-flex items-center gap-1.5 h-8 px-3 text-[13px] border-b-2 -mb-px transition-colors",
              active
                ? "border-accent text-text font-medium"
                : "border-transparent text-text-2 hover:text-text",
            )}
          >
            {it.icon}
            <span>{it.label}</span>
            {it.badge}
          </button>
        );
      })}
    </div>
  );
}
