"use client";

import Link from "next/link";
import { Folder } from "@/lib/api/folders";
import { Icon } from "@/components/ui/v2/icon";
import { highlight } from "./highlight";

export interface ResultFolderProps {
  folder: Folder;
  query: string;
}

export function ResultFolder({ folder, query }: ResultFolderProps) {
  return (
    <Link
      href={`/dashboard/files/${folder.id}`}
      className="flex items-center gap-3 rounded-lg border border-border bg-surface px-3 py-2.5 hover:border-border-strong transition-colors"
    >
      <span className="inline-flex h-8 w-8 items-center justify-center rounded bg-surface-2 text-text-2">
        <Icon name="Folder" size={16} />
      </span>
      <div className="min-w-0 flex-1">
        <div className="truncate text-[13px] font-medium text-text">
          {highlight(folder.name, query)}
        </div>
        <div className="text-[11px] text-text-3 mt-0.5">
          Depth {folder.depth}
        </div>
      </div>
      <Icon name="ChevronRight" size={12} className="text-text-3" />
    </Link>
  );
}
