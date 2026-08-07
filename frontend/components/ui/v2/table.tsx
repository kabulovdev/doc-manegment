import { HTMLAttributes, TdHTMLAttributes, ThHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

export function Table({ className, ...rest }: HTMLAttributes<HTMLTableElement>) {
  return (
    <table
      className={cn("w-full border-collapse text-[13px]", className)}
      {...rest}
    />
  );
}

export function THead({ className, ...rest }: HTMLAttributes<HTMLTableSectionElement>) {
  return (
    <thead
      className={cn(
        "text-left text-[11px] font-medium uppercase tracking-wide text-text-3",
        className,
      )}
      {...rest}
    />
  );
}

export interface TRowProps extends HTMLAttributes<HTMLTableRowElement> {
  selected?: boolean;
  interactive?: boolean;
}

export function TRow({ selected, interactive, className, style, ...rest }: TRowProps) {
  return (
    <tr
      className={cn(
        "border-b border-border",
        selected && "bg-accent-soft",
        interactive && "hover:bg-surface-2 cursor-pointer",
        className,
      )}
      style={{ height: "var(--row-h)", ...style }}
      {...rest}
    />
  );
}

export function TCell({
  className,
  ...rest
}: TdHTMLAttributes<HTMLTableCellElement>) {
  return (
    <td
      className={cn("px-3 align-middle text-text", className)}
      {...rest}
    />
  );
}

export function THCell({
  className,
  ...rest
}: ThHTMLAttributes<HTMLTableCellElement>) {
  return (
    <th
      className={cn(
        "px-3 py-2 align-middle font-medium border-b border-border bg-surface-2",
        className,
      )}
      {...rest}
    />
  );
}
