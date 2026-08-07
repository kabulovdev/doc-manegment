# Phase A — Design foundation

**Effort**: 3–4 days
**Depends on**: nothing
**Blocks**: B, C, D

## Context

The new design is built on CSS custom properties (light/dark/density/accent switchable at runtime) and a small primitives library (Button, Badge, Pill, Card, Icon, etc.). Before touching any page we need these building blocks in place. After this phase, old pages still render and work; new primitives sit alongside and are used by later phases.

Reference: mockup `src/primitives.jsx`, `src/icons.jsx`, `src/tweaks.jsx`, and the CSS at the top of `/Users/zafareshmamatov/Downloads/index (1).html` (lines 10–70).

## Goals

- CSS token system wired into Tailwind (`var(--bg)`, `var(--surface)`, etc.)
- Runtime theme switch (`data-theme="light|dark"`), density (`data-density="compact|default|comfortable"`), accent (`data-accent="emerald|indigo|amber|rose|slate"`)
- Fonts: Inter (sans) + JetBrains Mono (mono) via `next/font/google`
- Primitives in `frontend/components/ui/`: Button (enhanced), Input, Textarea, Select, Checkbox, Toggle, Card, SectionHead, Badge, Pill, Avatar, AvatarStack, Kbd, Progress, SegBar, Sparkbars, Dialog, Tabs, DropdownMenu, ContextMenu, Tooltip, Table, FileIcon, Icon
- `TweaksPanel` component that persists to `localStorage` and mutates the `<html>` data attributes
- Legacy primitives (`ui/button.tsx`, `ui/input.tsx`, `ui/dialog.tsx`) are NOT deleted — later phases migrate pages one by one.

## Tasks

### A1. Fonts

1. Edit `frontend/app/layout.tsx`:

   ```tsx
   import { Inter, JetBrains_Mono } from "next/font/google";

   const inter = Inter({ subsets: ["latin"], variable: "--font-sans" });
   const mono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono" });
   ```

2. Apply both CSS variables on `<html>`:

   ```tsx
   <html lang="en" className={`${inter.variable} ${mono.variable}`}>
   ```

### A2. CSS tokens in `globals.css`

Replace `frontend/app/globals.css` contents with:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
  --bg: #f7f8fa;
  --surface: #ffffff;
  --surface-2: #fafbfc;
  --border: #e6e8ec;
  --border-strong: #d4d7de;
  --text: #0b1020;
  --text-2: #4b5563;
  --text-3: #8a92a3;
  --accent: #10b981;
  --accent-2: #059669;
  --accent-soft: #ecfdf5;
  --accent-border: #a7f3d0;
  --danger: #ef4444;
  --warn: #f59e0b;
  --info: #0ea5e9;
  --violet: #8b5cf6;
  --row-h: 36px;
  --sidebar-w: 232px;
  --topbar-h: 48px;
  --radius: 6px;
  --radius-lg: 10px;
  --shadow-sm: 0 1px 2px rgba(10,14,30,.04), 0 0 0 1px rgba(10,14,30,.04);
  --shadow-md: 0 4px 12px rgba(10,14,30,.06), 0 0 0 1px rgba(10,14,30,.05);
  --shadow-lg: 0 20px 48px -12px rgba(10,14,30,.18), 0 0 0 1px rgba(10,14,30,.06);
}

html[data-theme="dark"] {
  --bg: #0b0f17;
  --surface: #11151f;
  --surface-2: #0e1219;
  --border: #1f2636;
  --border-strong: #2a3245;
  --text: #e7ebf3;
  --text-2: #9aa3b7;
  --text-3: #6b7489;
  --accent-soft: #052e24;
  --accent-border: #064e3b;
  --shadow-sm: 0 1px 2px rgba(0,0,0,.4), 0 0 0 1px rgba(255,255,255,.04);
  --shadow-md: 0 8px 20px rgba(0,0,0,.5), 0 0 0 1px rgba(255,255,255,.05);
  --shadow-lg: 0 24px 60px -12px rgba(0,0,0,.7), 0 0 0 1px rgba(255,255,255,.05);
}

html[data-density="compact"] { --row-h: 30px; --topbar-h: 44px; }
html[data-density="comfortable"] { --row-h: 44px; --topbar-h: 56px; }

html[data-accent="indigo"] { --accent: #6366f1; --accent-2: #4f46e5; --accent-soft: #eef2ff; --accent-border: #c7d2fe; }
html[data-accent="amber"]  { --accent: #f59e0b; --accent-2: #d97706; --accent-soft: #fffbeb; --accent-border: #fcd34d; }
html[data-accent="rose"]   { --accent: #f43f5e; --accent-2: #e11d48; --accent-soft: #fff1f2; --accent-border: #fecdd3; }
html[data-accent="slate"]  { --accent: #64748b; --accent-2: #475569; --accent-soft: #f1f5f9; --accent-border: #cbd5e1; }

html, body {
  background: var(--bg);
  color: var(--text);
  font-family: var(--font-sans), system-ui, sans-serif;
  font-size: 13px;
  line-height: 1.45;
  -webkit-font-smoothing: antialiased;
}

::-webkit-scrollbar { width: 10px; height: 10px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--border-strong); border-radius: 10px; border: 2px solid var(--bg); }
::-webkit-scrollbar-thumb:hover { background: var(--text-3); }

.font-mono { font-family: var(--font-mono), ui-monospace, Menlo, monospace; }
```

### A3. Tailwind config

Replace `frontend/tailwind.config.ts` with:

```ts
import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./lib/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "var(--bg)",
        surface: "var(--surface)",
        "surface-2": "var(--surface-2)",
        border: "var(--border)",
        "border-strong": "var(--border-strong)",
        text: "var(--text)",
        "text-2": "var(--text-2)",
        "text-3": "var(--text-3)",
        accent: "var(--accent)",
        "accent-2": "var(--accent-2)",
        "accent-soft": "var(--accent-soft)",
        "accent-border": "var(--accent-border)",
        danger: "var(--danger)",
        warn: "var(--warn)",
        info: "var(--info)",
        violet: "var(--violet)",
      },
      borderRadius: {
        DEFAULT: "var(--radius)",
        lg: "var(--radius-lg)",
      },
      boxShadow: {
        sm: "var(--shadow-sm)",
        md: "var(--shadow-md)",
        lg: "var(--shadow-lg)",
      },
      fontFamily: {
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "ui-monospace", "Menlo", "monospace"],
      },
    },
  },
  plugins: [],
};

export default config;
```

### A4. Primitives (create all new files under `frontend/components/ui/v2/`)

Using a `v2/` subfolder avoids name collisions. Later phases will re-export from `v2/` once the migration is complete.

Each primitive is a **small React component** with typed props and Tailwind + CSS-var classes. Structure:

- `v2/button.tsx` — variants: `primary` (slate/accent), `accent`, `secondary`, `ghost`, `soft`, `danger`; sizes: `sm`, `md`, `lg`; `icon` prop for icon-only
- `v2/icon-button.tsx` — square icon button with active state
- `v2/icon.tsx` — wraps `lucide-react` with consistent sizing; exports a `getIcon(name)` helper matching the mockup's icon names
- `v2/input.tsx`, `v2/textarea.tsx`, `v2/select.tsx` — styled form controls using `bg-surface-2`, `border-border`, `focus:ring-accent`
- `v2/checkbox.tsx`, `v2/toggle.tsx` — boolean controls with `accent-color`
- `v2/card.tsx` — wrapper with `bg-surface border-border shadow-sm rounded-lg`; props: `title`, `subtitle`, `action`, `padding`
- `v2/section-head.tsx` — page/section heading with subtitle and action slot
- `v2/badge.tsx` — small rounded label, `color` prop (`accent|danger|warn|info|violet|slate`), optional `dot`
- `v2/pill.tsx` — tag pill with colored dot; props: `name`, `color`, `onRemove`, `active`, `onClick`
- `v2/avatar.tsx`, `v2/avatar-stack.tsx` — initials + color, size sm/md/lg
- `v2/kbd.tsx` — keyboard-key chip
- `v2/progress.tsx`, `v2/seg-bar.tsx`, `v2/sparkbars.tsx` — bars/charts (SVG)
- `v2/dialog.tsx` — backdrop blur, centered panel, size sm/md/lg/full; returns focus on close
- `v2/tabs.tsx` — underline tabs; value/onChange controlled
- `v2/dropdown-menu.tsx` — headless menu positioned under an anchor; keyboard nav
- `v2/context-menu.tsx` — right-click menu
- `v2/tooltip.tsx` — hover tooltip; delay 200ms
- `v2/table.tsx` — headless `Table`, `THead`, `TRow`, `TCell`; respects `var(--row-h)`
- `v2/file-icon.tsx` — MIME-aware colored badge (PDF red, IMG blue, XLS green, etc.)
- `v2/skeleton.tsx` — shimmer placeholder
- `v2/empty-state.tsx` — centered empty state with icon/title/action

Each file ≤120 lines. Mirror the JSX in `/tmp/docmgmt-redesign/src/primitives.jsx` for exact visual behaviour, but translate classes from inline style maps to Tailwind utilities using the new `--bg`, `--accent`, etc. tokens.

### A5. TweaksProvider

Create `frontend/lib/tweaks/provider.tsx`:

- `TweaksProvider` — reads `theme`, `density`, `accent` from `localStorage` on mount; sets `data-theme`, `data-density`, `data-accent` on `<html>`; exposes `useTweaks()` hook with setters
- Default: `theme=light`, `density=default`, `accent=emerald`
- Add `<TweaksProvider>` inside `app/providers.tsx` wrapping `<QueryClientProvider>`

Create `frontend/components/ui/v2/tweaks-panel.tsx`:

- Small floating button bottom-right (gear icon) that opens a side panel
- Radio groups: Theme (Light / Dark), Density (Compact / Default / Comfortable), Accent (5 colors as swatches)
- Changes apply live and persist

Do NOT mount the panel globally yet — Phase B will add it to the new shell.

### A6. Update `app/layout.tsx`

Wrap children in `TweaksProvider`. Final layout tree:

```
<html>
  <body>
    <TweaksProvider>
      <QueryClientProvider>
        {children}
      </QueryClientProvider>
    </TweaksProvider>
  </body>
</html>
```

## Acceptance criteria

- [ ] `cd frontend && ./node_modules/.bin/tsc --noEmit` passes
- [ ] `./node_modules/.bin/next lint --quiet` passes
- [ ] Opening http://localhost:3001 still shows the existing dashboard (no page migrated yet)
- [ ] Every new primitive has a working default export and typed props
- [ ] `document.documentElement.dataset.theme = "dark"` in DevTools instantly darkens the UI (tokens wired)
- [ ] Importing any `v2/*` primitive from a test page renders correctly in both light and dark
- [ ] Docker image builds: `docker compose build frontend` succeeds
- [ ] A scratch page (e.g. `app/_sandbox/primitives/page.tsx` — delete before PR) showing every primitive renders without errors

## Verification commands

```bash
cd /Users/zafareshmamatov/go/src/gitlab.com/docuemnt_manegment/frontend
./node_modules/.bin/tsc --noEmit
./node_modules/.bin/next lint --quiet
cd .. && docker compose build frontend
```

## Out of scope

- Don't touch `app/dashboard/**` pages yet — those are migrated in Phase C
- Don't delete old `components/ui/button.tsx`, `input.tsx`, `dialog.tsx` — still used by current pages
- Don't build a Cmd+K palette (Phase B)
- No new routes
- No backend changes
