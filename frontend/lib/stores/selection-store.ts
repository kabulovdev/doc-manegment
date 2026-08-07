"use client";

import { create } from "zustand";

interface SelectionState {
  selectedFileIDs: Set<string>;
  toggle: (id: string) => void;
  add: (id: string) => void;
  remove: (id: string) => void;
  setMany: (ids: string[]) => void;
  clear: () => void;
  isSelected: (id: string) => boolean;
  count: () => number;
}

export const useSelectionStore = create<SelectionState>((set, get) => ({
  selectedFileIDs: new Set<string>(),
  toggle: (id) =>
    set((s) => {
      const next = new Set(s.selectedFileIDs);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return { selectedFileIDs: next };
    }),
  add: (id) =>
    set((s) => {
      if (s.selectedFileIDs.has(id)) return s;
      const next = new Set(s.selectedFileIDs);
      next.add(id);
      return { selectedFileIDs: next };
    }),
  remove: (id) =>
    set((s) => {
      if (!s.selectedFileIDs.has(id)) return s;
      const next = new Set(s.selectedFileIDs);
      next.delete(id);
      return { selectedFileIDs: next };
    }),
  setMany: (ids) => set({ selectedFileIDs: new Set(ids) }),
  clear: () => set({ selectedFileIDs: new Set() }),
  isSelected: (id) => get().selectedFileIDs.has(id),
  count: () => get().selectedFileIDs.size,
}));
