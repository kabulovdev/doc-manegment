"use client";

import {
  ReactNode,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";

export type Theme = "light" | "dark";
export type Density = "compact" | "default" | "comfortable";
export type Accent = "emerald" | "indigo" | "amber" | "rose" | "slate";

export interface TweaksState {
  theme: Theme;
  density: Density;
  accent: Accent;
}

export interface TweaksContextValue extends TweaksState {
  setTheme: (t: Theme) => void;
  setDensity: (d: Density) => void;
  setAccent: (a: Accent) => void;
}

const STORAGE_KEY = "docmgmt:tweaks:v1";

const DEFAULTS: TweaksState = {
  theme: "light",
  density: "default",
  accent: "emerald",
};

const TweaksContext = createContext<TweaksContextValue | null>(null);

function applyToDOM(state: TweaksState) {
  if (typeof document === "undefined") return;
  const html = document.documentElement;
  html.dataset.theme = state.theme;
  html.dataset.density = state.density;
  html.dataset.accent = state.accent;
}

function load(): TweaksState {
  if (typeof window === "undefined") return DEFAULTS;
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return DEFAULTS;
    const parsed = JSON.parse(raw) as Partial<TweaksState>;
    return {
      theme: parsed.theme === "dark" ? "dark" : "light",
      density:
        parsed.density === "compact" || parsed.density === "comfortable"
          ? parsed.density
          : "default",
      accent:
        parsed.accent === "indigo" ||
        parsed.accent === "amber" ||
        parsed.accent === "rose" ||
        parsed.accent === "slate"
          ? parsed.accent
          : "emerald",
    };
  } catch {
    return DEFAULTS;
  }
}

function persist(state: TweaksState) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } catch {
    /* ignore quota */
  }
}

export function TweaksProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<TweaksState>(DEFAULTS);

  useEffect(() => {
    const loaded = load();
    setState(loaded);
    applyToDOM(loaded);
  }, []);

  const update = useCallback((patch: Partial<TweaksState>) => {
    setState((prev) => {
      const next = { ...prev, ...patch };
      applyToDOM(next);
      persist(next);
      return next;
    });
  }, []);

  const value: TweaksContextValue = {
    ...state,
    setTheme: (theme) => update({ theme }),
    setDensity: (density) => update({ density }),
    setAccent: (accent) => update({ accent }),
  };

  return (
    <TweaksContext.Provider value={value}>{children}</TweaksContext.Provider>
  );
}

export function useTweaks(): TweaksContextValue {
  const ctx = useContext(TweaksContext);
  if (!ctx) throw new Error("useTweaks must be used inside <TweaksProvider>");
  return ctx;
}
