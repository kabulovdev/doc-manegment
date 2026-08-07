import { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

export function Kbd({ className, children, ...rest }: HTMLAttributes<HTMLElement>) {
  return (
    <kbd
      className={cn(
        "inline-flex items-center justify-center rounded border border-border bg-surface-2",
        "px-1.5 h-5 min-w-5 text-[10px] font-mono text-text-2 shadow-sm",
        className,
      )}
      {...rest}
    >
      {children}
    </kbd>
  );
}
