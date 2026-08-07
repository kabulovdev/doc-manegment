import { cn } from "@/lib/utils";

export type PillColor =
  | "emerald"
  | "indigo"
  | "amber"
  | "rose"
  | "slate"
  | "sky"
  | "violet";

const pillColors: Record<PillColor, string> = {
  emerald: "#10b981",
  indigo: "#6366f1",
  amber: "#f59e0b",
  rose: "#f43f5e",
  slate: "#64748b",
  sky: "#0ea5e9",
  violet: "#8b5cf6",
};

export interface PillProps {
  name: string;
  color?: PillColor;
  onRemove?: () => void;
  onClick?: () => void;
  active?: boolean;
  className?: string;
}

export function Pill({ name, color = "slate", onRemove, onClick, active, className }: PillProps) {
  const dot = pillColors[color];
  const interactive = onClick || onRemove;
  return (
    <span
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[11px] font-medium",
        active
          ? "bg-accent-soft border-accent-border text-accent-2"
          : "bg-surface border-border text-text-2",
        interactive && "cursor-pointer hover:border-border-strong",
        className,
      )}
    >
      <span
        className="h-1.5 w-1.5 rounded-full"
        style={{ background: dot }}
      />
      <span>{name}</span>
      {onRemove && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onRemove();
          }}
          className="ml-0.5 text-text-3 hover:text-text"
          aria-label={`Remove ${name}`}
        >
          ×
        </button>
      )}
    </span>
  );
}
