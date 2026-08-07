import { ReactNode } from "react";
import { cn } from "@/lib/utils";

export interface EmptyStateProps {
  icon?: ReactNode;
  title: ReactNode;
  description?: ReactNode;
  action?: ReactNode;
  className?: string;
}

export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center text-center py-12 px-6",
        className,
      )}
    >
      {icon && (
        <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-surface-2 border border-border text-text-3">
          {icon}
        </div>
      )}
      <div className="text-[13px] font-semibold text-text">{title}</div>
      {description && (
        <div className="mt-1 max-w-sm text-xs text-text-2">{description}</div>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}
