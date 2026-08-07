"use client";

import { useEffect, useState } from "react";
import { usePageTitleStore } from "@/lib/stores/page-title-store";
import { useCommandPalette } from "@/lib/stores/command-palette-store";
import { Button } from "@/components/ui/v2/button";
import { IconButton } from "@/components/ui/v2/icon-button";
import { Icon } from "@/components/ui/v2/icon";
import { Kbd } from "@/components/ui/v2/kbd";
import { Tooltip } from "@/components/ui/v2/tooltip";
import { Badge } from "@/components/ui/v2/badge";
import { cn } from "@/lib/utils";

export interface TopbarProps {
  onToggleSidebar: () => void;
  onOpenUpload: () => void;
  showHamburger?: boolean;
}

export function Topbar({ onToggleSidebar, onOpenUpload, showHamburger }: TopbarProps) {
  const { title, subtitle } = usePageTitleStore();
  const openPalette = useCommandPalette((s) => s.openPalette);
  const [mod, setMod] = useState("⌘");

  useEffect(() => {
    if (typeof navigator !== "undefined" && !/Mac|iPhone|iPad|iPod/.test(navigator.platform)) {
      setMod("Ctrl");
    }
  }, []);

  return (
    <header
      className="flex items-center gap-3 border-b border-border bg-surface px-3 md:px-4"
      style={{ height: "var(--topbar-h)" }}
    >
      {showHamburger && (
        <button
          type="button"
          onClick={onToggleSidebar}
          aria-label="Open navigation"
          className="inline-flex h-8 w-8 items-center justify-center rounded text-text-2 hover:bg-surface-2 md:hidden"
        >
          <Icon name="Menu" size={18} />
        </button>
      )}

      <div className="min-w-0 flex-1 md:flex-none md:w-48">
        <div className="text-[13px] font-semibold text-text truncate">
          {title || "Doc Manager"}
        </div>
        {subtitle && (
          <div className="text-[11px] text-text-3 truncate">{subtitle}</div>
        )}
      </div>

      <div className="hidden md:flex flex-1 justify-center">
        <button
          type="button"
          onClick={() => openPalette()}
          className={cn(
            "flex items-center gap-2 h-8 w-full max-w-[480px] rounded border border-border bg-surface-2",
            "px-2.5 text-[12px] text-text-3 hover:border-border-strong hover:text-text-2",
          )}
        >
          <Icon name="Search" size={14} />
          <span className="flex-1 text-left">Search files, tags, storages…</span>
          <span className="flex items-center gap-0.5">
            <Kbd>{mod}</Kbd>
            <Kbd>K</Kbd>
          </span>
        </button>
      </div>

      <div className="flex items-center gap-1">
        <button
          type="button"
          onClick={() => openPalette()}
          aria-label="Search"
          className="inline-flex h-8 w-8 items-center justify-center rounded text-text-2 hover:bg-surface-2 md:hidden"
        >
          <Icon name="Search" size={16} />
        </button>

        <Tooltip content="Notifications" side="bottom">
          <IconButton aria-label="Notifications" className="relative hidden md:inline-flex">
            <Icon name="Bell" size={16} />
            <Badge
              color="slate"
              className="absolute -top-1 -right-1 h-4 min-w-4 justify-center px-1 text-[9px]"
            >
              0
            </Badge>
          </IconButton>
        </Tooltip>

        <Tooltip content="Help" side="bottom">
          <IconButton aria-label="Help" className="hidden md:inline-flex">
            <Icon name="HelpCircle" size={16} />
          </IconButton>
        </Tooltip>

        <Button variant="accent" size="md" onClick={onOpenUpload}>
          <Icon name="Upload" size={14} />
          <span className="hidden sm:inline">Upload</span>
        </Button>
      </div>
    </header>
  );
}
