import { HTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/utils";

export interface CardProps extends Omit<HTMLAttributes<HTMLDivElement>, "title"> {
  title?: ReactNode;
  subtitle?: ReactNode;
  action?: ReactNode;
  padding?: "none" | "sm" | "md" | "lg";
}

const padMap = {
  none: "",
  sm: "p-3",
  md: "p-4",
  lg: "p-5",
} as const;

export function Card({
  title,
  subtitle,
  action,
  padding = "md",
  className,
  children,
  ...rest
}: CardProps) {
  const hasHeader = title || subtitle || action;
  return (
    <div
      className={cn(
        "bg-surface border border-border rounded-lg shadow-sm",
        className,
      )}
      {...rest}
    >
      {hasHeader && (
        <div className="flex items-start justify-between gap-3 px-4 pt-3 pb-2 border-b border-border">
          <div className="min-w-0">
            {title && (
              <div className="text-[13px] font-semibold text-text truncate">
                {title}
              </div>
            )}
            {subtitle && (
              <div className="text-xs text-text-2 mt-0.5">{subtitle}</div>
            )}
          </div>
          {action && <div className="flex items-center gap-1.5">{action}</div>}
        </div>
      )}
      <div className={cn(padMap[padding])}>{children}</div>
    </div>
  );
}
