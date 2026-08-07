import { cn } from "@/lib/utils";

export interface SparkbarsProps {
  data: number[];
  height?: number;
  width?: number;
  className?: string;
  color?: string;
}

export function Sparkbars({
  data,
  height = 28,
  width = 80,
  className,
  color = "var(--accent)",
}: SparkbarsProps) {
  if (!data.length) return null;
  const max = Math.max(...data, 1);
  const gap = 2;
  const barW = Math.max(1, (width - gap * (data.length - 1)) / data.length);
  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      className={cn("overflow-visible", className)}
      aria-hidden
    >
      {data.map((v, i) => {
        const h = Math.max(1, (v / max) * height);
        const x = i * (barW + gap);
        const y = height - h;
        return <rect key={i} x={x} y={y} width={barW} height={h} fill={color} rx={1} />;
      })}
    </svg>
  );
}
