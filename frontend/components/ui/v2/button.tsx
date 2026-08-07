import { ButtonHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

export type ButtonVariant =
  | "primary"
  | "accent"
  | "secondary"
  | "ghost"
  | "soft"
  | "danger";
export type ButtonSize = "sm" | "md" | "lg";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: boolean;
}

const base =
  "inline-flex items-center justify-center gap-1.5 font-medium rounded transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-accent/60 focus-visible:ring-offset-1 focus-visible:ring-offset-bg disabled:opacity-50 disabled:pointer-events-none select-none whitespace-nowrap";

const variants: Record<ButtonVariant, string> = {
  primary:
    "bg-[var(--text)] text-[var(--surface)] hover:opacity-90 border border-transparent",
  accent:
    "bg-accent text-white hover:bg-accent-2 border border-transparent",
  secondary:
    "bg-surface text-text border border-border hover:bg-surface-2",
  ghost: "text-text-2 hover:bg-surface-2 border border-transparent",
  soft: "bg-accent-soft text-accent-2 border border-accent-border hover:brightness-[0.98]",
  danger: "bg-danger text-white hover:brightness-110 border border-transparent",
};

const sizes: Record<ButtonSize, string> = {
  sm: "h-7 px-2 text-xs",
  md: "h-8 px-3 text-[13px]",
  lg: "h-10 px-4 text-sm",
};

const iconSizes: Record<ButtonSize, string> = {
  sm: "h-7 w-7 p-0",
  md: "h-8 w-8 p-0",
  lg: "h-10 w-10 p-0",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "secondary", size = "md", icon, ...props }, ref) => (
    <button
      ref={ref}
      className={cn(
        base,
        variants[variant],
        icon ? iconSizes[size] : sizes[size],
        className,
      )}
      {...props}
    />
  ),
);
Button.displayName = "Button";
