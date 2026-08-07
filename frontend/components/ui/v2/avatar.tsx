import { cn } from "@/lib/utils";

export type AvatarSize = "sm" | "md" | "lg";

export interface AvatarProps {
  name: string;
  color?: string;
  size?: AvatarSize;
  className?: string;
}

const sizeMap: Record<AvatarSize, string> = {
  sm: "h-5 w-5 text-[10px]",
  md: "h-7 w-7 text-[11px]",
  lg: "h-9 w-9 text-[13px]",
};

const palette = ["#10b981", "#6366f1", "#f59e0b", "#f43f5e", "#0ea5e9", "#8b5cf6", "#64748b"];

function colorFor(name: string) {
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) | 0;
  return palette[Math.abs(hash) % palette.length];
}

function initials(name: string) {
  const parts = name.trim().split(/\s+/).slice(0, 2);
  return parts.map((p) => p[0]?.toUpperCase() ?? "").join("") || "?";
}

export function Avatar({ name, color, size = "md", className }: AvatarProps) {
  const bg = color ?? colorFor(name);
  return (
    <span
      title={name}
      className={cn(
        "inline-flex items-center justify-center rounded-full font-semibold text-white ring-2 ring-surface",
        sizeMap[size],
        className,
      )}
      style={{ background: bg }}
    >
      {initials(name)}
    </span>
  );
}
