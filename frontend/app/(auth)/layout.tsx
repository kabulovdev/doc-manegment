"use client";

import { ReactNode } from "react";
import { TweaksPanel } from "@/components/ui/v2/tweaks-panel";
import { Toaster } from "@/components/ui/v2/toaster";
import { Icon } from "@/components/ui/v2/icon";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="relative flex min-h-screen flex-col items-center justify-center bg-bg p-4">
      <div className="mb-6 flex items-center gap-2 text-text-2">
        <span className="inline-flex h-7 w-7 items-center justify-center rounded bg-accent text-white">
          <Icon name="Layers" size={16} />
        </span>
        <span className="text-[15px] font-semibold text-text">Doc Manager</span>
      </div>
      {children}
      <TweaksPanel />
      <Toaster />
    </div>
  );
}
