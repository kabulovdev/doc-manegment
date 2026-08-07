"use client";

import { ReactNode, useEffect, useState } from "react";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";
import { CommandPalette } from "./command-palette";
import { UploadDialogStub } from "./upload-dialog-stub";
import { TweaksPanel } from "@/components/ui/v2/tweaks-panel";
import { Toaster } from "@/components/ui/v2/toaster";
import { cn } from "@/lib/utils";

const STORAGE_KEY = "docmgmt:sidebar-collapsed";

function loadCollapsed(): boolean {
  if (typeof window === "undefined") return false;
  try {
    return window.localStorage.getItem(STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}

function persistCollapsed(v: boolean) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(STORAGE_KEY, v ? "1" : "0");
  } catch {
    /* ignore */
  }
}

export function AppShell({ children }: { children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [uploadOpen, setUploadOpen] = useState(false);

  useEffect(() => {
    setCollapsed(loadCollapsed());
  }, []);

  const toggleSidebar = () => {
    setCollapsed((v) => {
      const next = !v;
      persistCollapsed(next);
      return next;
    });
  };

  return (
    <div className="flex h-screen min-h-screen overflow-hidden bg-bg text-text">
      <div className="hidden md:flex shrink-0">
        <Sidebar collapsed={collapsed} onToggle={toggleSidebar} />
      </div>

      {drawerOpen && (
        <div
          className="fixed inset-0 z-30 md:hidden"
          onClick={() => setDrawerOpen(false)}
        >
          <div className="absolute inset-0 bg-black/40 backdrop-blur-sm" />
          <div
            className="absolute inset-y-0 left-0 z-10"
            onClick={(e) => e.stopPropagation()}
          >
            <Sidebar
              collapsed={false}
              onToggle={() => {}}
              mobile
              onNavigate={() => setDrawerOpen(false)}
            />
          </div>
        </div>
      )}

      <div className="flex min-w-0 flex-1 flex-col">
        <Topbar
          onToggleSidebar={() => setDrawerOpen(true)}
          onOpenUpload={() => setUploadOpen(true)}
          showHamburger
        />
        <main className={cn("flex-1 overflow-y-auto p-6 md:p-8")}>{children}</main>
      </div>

      <CommandPalette />
      <UploadDialogStub open={uploadOpen} onClose={() => setUploadOpen(false)} />
      <TweaksPanel />
      <Toaster />
    </div>
  );
}
