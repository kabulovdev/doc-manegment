import { SelectHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  invalid?: boolean;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, invalid, children, ...props }, ref) => (
    <select
      ref={ref}
      className={cn(
        "h-8 w-full rounded border bg-surface-2 px-2 pr-7 text-[13px] text-text appearance-none",
        "focus:outline-none focus:bg-surface focus:ring-2 focus:ring-accent/30 focus:border-accent",
        "disabled:opacity-50 disabled:cursor-not-allowed",
        invalid ? "border-danger" : "border-border",
        "bg-no-repeat bg-[right_0.5rem_center]",
        className,
      )}
      style={{
        backgroundImage:
          "url(\"data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%238a92a3' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><polyline points='6 9 12 15 18 9'/></svg>\")",
        backgroundSize: "12px",
      }}
      {...props}
    >
      {children}
    </select>
  ),
);
Select.displayName = "Select";
