"use client";

import { useEffect } from "react";
import { create } from "zustand";

interface PageTitleState {
  title: string;
  subtitle?: string;
  setPageTitle: (title: string, subtitle?: string) => void;
}

export const usePageTitleStore = create<PageTitleState>((set) => ({
  title: "",
  subtitle: undefined,
  setPageTitle: (title, subtitle) => set({ title, subtitle }),
}));

export function usePageTitle(title: string, subtitle?: string) {
  const setPageTitle = usePageTitleStore((s) => s.setPageTitle);
  useEffect(() => {
    setPageTitle(title, subtitle);
  }, [title, subtitle, setPageTitle]);
}
