import { Avatar, AvatarSize } from "./avatar";
import { cn } from "@/lib/utils";

export interface AvatarStackProps {
  names: string[];
  max?: number;
  size?: AvatarSize;
  className?: string;
}

export function AvatarStack({ names, max = 3, size = "sm", className }: AvatarStackProps) {
  const visible = names.slice(0, max);
  const extra = names.length - visible.length;
  return (
    <div className={cn("inline-flex items-center -space-x-1.5", className)}>
      {visible.map((n) => (
        <Avatar key={n} name={n} size={size} />
      ))}
      {extra > 0 && (
        <span className="inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-surface-2 border border-border text-[10px] font-medium text-text-2 px-1.5 ring-2 ring-surface">
          +{extra}
        </span>
      )}
    </div>
  );
}
