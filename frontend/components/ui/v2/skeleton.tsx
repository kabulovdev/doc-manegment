import { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

export function Skeleton({ className, ...rest }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "animate-pulse rounded bg-surface-2 border border-border/60",
        className,
      )}
      {...rest}
    />
  );
}
