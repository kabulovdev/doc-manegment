import { ReactNode } from "react";
import { cn } from "@/lib/utils";

export interface SectionHeadProps {
  title: ReactNode;
  subtitle?: ReactNode;
  action?: ReactNode;
  level?: "h1" | "h2" | "h3";
  className?: string;
}

const titleSize = {
  h1: "text-lg font-semibold",
  h2: "text-base font-semibold",
  h3: "text-[13px] font-semibold",
} as const;

export function SectionHead({
  title,
  subtitle,
  action,
  level = "h2",
  className,
}: SectionHeadProps) {
  const Tag = level;
  return (
    <div className={cn("flex items-start justify-between gap-3", className)}>
      <div className="min-w-0">
        <Tag className={cn(titleSize[level], "text-text truncate")}>{title}</Tag>
        {subtitle && (
          <div className="text-xs text-text-2 mt-1">{subtitle}</div>
        )}
      </div>
      {action && <div className="flex items-center gap-1.5">{action}</div>}
    </div>
  );
}
