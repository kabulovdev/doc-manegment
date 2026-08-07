import { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

export type BadgeColor = "accent" | "danger" | "warn" | "info" | "violet" | "slate";

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  color?: BadgeColor;
  dot?: boolean;
}

const colorMap: Record<BadgeColor, { bg: string; text: string; dot: string; border: string }> = {
  accent: { bg: "bg-accent-soft", text: "text-accent-2", dot: "bg-accent", border: "border-accent-border" },
  danger: { bg: "bg-red-50", text: "text-red-700", dot: "bg-danger", border: "border-red-200" },
  warn: { bg: "bg-amber-50", text: "text-amber-700", dot: "bg-warn", border: "border-amber-200" },
  info: { bg: "bg-sky-50", text: "text-sky-700", dot: "bg-info", border: "border-sky-200" },
  violet: { bg: "bg-violet-50", text: "text-violet-700", dot: "bg-violet", border: "border-violet-200" },
  slate: { bg: "bg-surface-2", text: "text-text-2", dot: "bg-text-3", border: "border-border" },
};

export function Badge({ color = "slate", dot, className, children, ...rest }: BadgeProps) {
  const c = colorMap[color];
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[11px] font-medium border",
        c.bg,
        c.text,
        c.border,
        className,
      )}
      {...rest}
    >
      {dot && <span className={cn("h-1.5 w-1.5 rounded-full", c.dot)} />}
      {children}
    </span>
  );
}
