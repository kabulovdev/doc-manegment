import { cn } from "@/lib/utils";

export interface SegBarSegment {
  value: number;
  color: string;
  label?: string;
}

export interface SegBarProps {
  segments: SegBarSegment[];
  className?: string;
}

export function SegBar({ segments, className }: SegBarProps) {
  const total = segments.reduce((s, x) => s + x.value, 0) || 1;
  return (
    <div
      className={cn(
        "flex h-2 w-full overflow-hidden rounded-full bg-surface-2 border border-border",
        className,
      )}
    >
      {segments.map((seg, i) => (
        <div
          key={i}
          title={seg.label}
          style={{ width: `${(seg.value / total) * 100}%`, background: seg.color }}
          className="h-full"
        />
      ))}
    </div>
  );
}
