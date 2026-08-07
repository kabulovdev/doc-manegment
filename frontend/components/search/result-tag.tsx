"use client";

import Link from "next/link";
import { Tag } from "@/lib/api/tags";
import { Pill, PillColor } from "@/components/ui/v2/pill";

const pillColorFor = (hex: string): PillColor => {
  const h = hex.toLowerCase();
  if (h === "#10b981") return "emerald";
  if (h === "#6366f1") return "indigo";
  if (h === "#f59e0b") return "amber";
  if (h === "#f43f5e") return "rose";
  if (h === "#0ea5e9") return "sky";
  if (h === "#8b5cf6") return "violet";
  return "slate";
};

export interface ResultTagProps {
  tag: Tag;
}

export function ResultTag({ tag }: ResultTagProps) {
  return (
    <Link href="/dashboard/tags" className="inline-flex">
      <Pill name={tag.name} color={pillColorFor(tag.color)} />
    </Link>
  );
}
