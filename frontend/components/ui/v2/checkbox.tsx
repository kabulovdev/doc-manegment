import { InputHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

export type CheckboxProps = InputHTMLAttributes<HTMLInputElement>;

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ className, style, ...props }, ref) => (
    <input
      ref={ref}
      type="checkbox"
      className={cn(
        "h-4 w-4 rounded border-border bg-surface-2 cursor-pointer",
        "focus:outline-none focus-visible:ring-2 focus-visible:ring-accent/50",
        "disabled:opacity-50 disabled:cursor-not-allowed",
        className,
      )}
      style={{ accentColor: "var(--accent)", ...style }}
      {...props}
    />
  ),
);
Checkbox.displayName = "Checkbox";
