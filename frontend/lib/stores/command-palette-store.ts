"use client";

import { create } from "zustand";

export type PaletteScope = "all" | "files" | "tags" | "folders" | "storages";

interface CommandPaletteState {
  open: boolean;
  scope: PaletteScope;
  openPalette: (scope?: PaletteScope) => void;
  closePalette: () => void;
  togglePalette: () => void;
  setScope: (scope: PaletteScope) => void;
}

export const useCommandPalette = create<CommandPaletteState>((set) => ({
  open: false,
  scope: "all",
  openPalette: (scope) => set({ open: true, scope: scope ?? "all" }),
  closePalette: () => set({ open: false }),
  togglePalette: () => set((s) => ({ open: !s.open })),
  setScope: (scope) => set({ scope }),
}));
