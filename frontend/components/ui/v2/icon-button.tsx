import { ButtonHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

export type IconButtonSize = "sm" | "md" | "lg";

export interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  size?: IconButtonSize;
  active?: boolean;
}

const sizes: Record<IconButtonSize, string> = {
  sm: "h-7 w-7",
  md: "h-8 w-8",
  lg: "h-10 w-10",
};

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ className, size = "md", active, ...props }, ref) => (
    <button
      ref={ref}
      aria-pressed={active}
      className={cn(
        "inline-flex items-center justify-center rounded border transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-accent/60 disabled:opacity-50 disabled:pointer-events-none",
        sizes[size],
        active
          ? "bg-accent-soft border-accent-border text-accent-2"
          : "bg-surface border-border text-text-2 hover:bg-surface-2 hover:text-text",
        className,
      )}
      {...props}
    />
  ),
);
IconButton.displayName = "IconButton";
