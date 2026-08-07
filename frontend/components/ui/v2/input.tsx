import { InputHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, invalid, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        "h-8 w-full rounded border bg-surface-2 px-2.5 text-[13px] text-text placeholder:text-text-3",
        "focus:outline-none focus:bg-surface focus:ring-2 focus:ring-accent/30 focus:border-accent",
        "disabled:opacity-50 disabled:cursor-not-allowed",
        invalid ? "border-danger" : "border-border",
        className,
      )}
      {...props}
    />
  ),
);
Input.displayName = "Input";
