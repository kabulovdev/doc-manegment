"use client";

import { ReactNode, useRef, useState } from "react";
import { cn } from "@/lib/utils";

export interface TooltipProps {
  content: ReactNode;
  children: ReactNode;
  side?: "top" | "bottom" | "left" | "right";
  delay?: number;
  className?: string;
}

const sideMap = {
  top: "bottom-full mb-1.5 left-1/2 -translate-x-1/2",
  bottom: "top-full mt-1.5 left-1/2 -translate-x-1/2",
  left: "right-full mr-1.5 top-1/2 -translate-y-1/2",
  right: "left-full ml-1.5 top-1/2 -translate-y-1/2",
} as const;

export function Tooltip({
  content,
  children,
  side = "top",
  delay = 200,
  className,
}: TooltipProps) {
  const [visible, setVisible] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const show = () => {
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => setVisible(true), delay);
  };
  const hide = () => {
    if (timer.current) clearTimeout(timer.current);
    setVisible(false);
  };

  return (
    <span
      className="relative inline-flex"
      onMouseEnter={show}
      onMouseLeave={hide}
      onFocus={show}
      onBlur={hide}
    >
      {children}
      {visible && (
        <span
          role="tooltip"
          className={cn(
            "absolute z-50 px-2 py-1 rounded bg-[var(--text)] text-[var(--surface)] text-[11px] font-medium whitespace-nowrap shadow-md pointer-events-none",
            sideMap[side],
            className,
          )}
        >
          {content}
        </span>
      )}
    </span>
  );
}
