# Phase B — Shell + navigation

**Effort**: 2–3 days
**Depends on**: Phase A (primitives)
**Blocks**: C, D

## Context

The mockup's `shell.jsx` defines a new sidebar (expand/collapse, workspace switcher, storage usage bar, user footer) and topbar (title, Cmd+K search trigger, notifications, help, upload). We replace the current simple sidebar with this, and introduce a Cmd+K palette that searches across existing backend endpoints (no new endpoints).

Reference: `/tmp/docmgmt-redesign/src/shell.jsx`, `page-search.jsx` (the palette pattern).

## Goals

- New `AppShell` component: sidebar + topbar + main content
- Sidebar: expand/collapse, 6 primary + 2 secondary nav items, workspace switcher (single workspace stub), aggregate storage usage bar, user profile footer
- Topbar: page title, global search trigger (Cmd+K), notifications bell (stub badge), help (`?`), Upload button
- Cmd+K palette: keyboard shortcut (Cmd/Ctrl+K), scope filter (All / Files / Tags / Folders / Storages), searches via existing API
- Tweaks panel mounted in the shell (bottom-right floating button)
- Mobile: sidebar collapses to icon-only on narrow viewports; becomes a drawer below 768px

## Tasks

### B1. `AppShell` component

Create `frontend/components/shell/app-shell.tsx`:

```tsx
"use client";
export function AppShell({ children }: { children: React.ReactNode }) {
  // sidebar open/collapsed state with localStorage persistence
  // layout: flex, sidebar on left, topbar + main on right
}
```

### B2. Sidebar

Create `frontend/components/shell/sidebar.tsx`:

- Props: `collapsed: boolean`, `onToggle: () => void`
- Expanded width: `var(--sidebar-w)` = 232px. Collapsed: 56px. Smooth transition.
- Sections:
  - Workspace switcher (expanded only) — select with current workspace name; dropdown is stubbed but renders
  - Primary nav:
    - Overview → `/dashboard`
    - Files → `/dashboard/files`
    - Storages → `/dashboard/storages`
    - Tags → `/dashboard/tags`
    - Shares → `/dashboard/shares`
    - Search → `/dashboard/search` (route added in Phase D; for now render link that 404s is OK)
  - Divider
  - Secondary nav:
    - Team → `/dashboard/team` (stub, 404 until Phase E)
    - Settings → `/dashboard/settings` (added in Phase D)
  - Spacer (flex-1)
  - Storage usage bar (expanded only): aggregate `used_bytes / max` across all user storages (use `storagesApi.list()` + compute); progress bar using `Progress` primitive
  - User footer: Avatar + name + email + logout (keep existing logout logic)
- Active item: `data-active` attr → `bg-accent-soft text-accent` via Tailwind variants
- Collapsed mode: hide labels, show only icons; tooltip appears on hover (uses `v2/tooltip`)

### B3. Topbar

Create `frontend/components/shell/topbar.tsx`:

- Height: `var(--topbar-h)` (48px default)
- Layout: left (page title slot), center (search trigger button, width 480px, shows "Search… ⌘K"), right (icon buttons: bell, help, upload)
- `title` prop passed from the page via a context or a `usePageTitle()` hook
- Clicking the search button opens the command palette
- Clicking Upload opens the upload dialog (reuses Phase C's upload dialog; until then opens a stub)

### B4. Command palette (Cmd+K)

Create `frontend/components/shell/command-palette.tsx`:

- Opens on `Cmd/Ctrl+K` from anywhere inside `AppShell`
- Backdrop blur, centered panel (width 640px, max-height 80vh)
- Input auto-focus; types to filter
- Scope chips: All (default) / Files / Tags / Folders / Storages (click to narrow)
- Results sections, each uses the appropriate API:
  - Files: `filesApi.list({ q })` — show file icon, name, folder path, size
  - Tags: `tagsApi.list()` then client-filter by name
  - Folders: `foldersApi.list()` recursively — or reuse existing full-list; filter by name
  - Storages: `storagesApi.list()` — filter by display_name or bucket
- Arrow keys + Enter navigate and open; Esc closes
- Empty state if no query and no scope

Implementation notes:
- Use a single state hook `useCommandPalette()` (Zustand) for open/close, so any component can open it (e.g. topbar search button, shortcut)
- Group results with section headers; max 5 per group
- Highlight matched substring in result labels (simple `<mark>`)

### B5. Mount tweaks panel

Add `<TweaksPanel />` to `AppShell` as a floating button (bottom-right, 16px from edges, z-40).

### B6. Replace `app/dashboard/layout.tsx`

The current `app/dashboard/layout.tsx` renders its own sidebar + auth guard. Replace its JSX so it uses the new `AppShell`:

```tsx
"use client";
export default function DashboardLayout({ children }) {
  // keep existing auth guard (redirect to /login if no accessToken)
  return <AppShell>{children}</AppShell>;
}
```

Keep the auth redirect behaviour. Move logout into the new sidebar's user footer.

### B7. Mobile responsiveness

- Below 768px: sidebar becomes a drawer; a hamburger button in the topbar opens it
- Topbar compresses: hide help/bell, keep search + upload
- Tailwind classes: use `md:` prefix gates

## Acceptance criteria

- [ ] Every existing dashboard route (`/dashboard`, `/dashboard/files`, `/dashboard/files/[folderId]`, `/dashboard/storages`, `/dashboard/tags`, `/dashboard/shares`) still renders the same page content, but inside the new shell
- [ ] Sidebar collapses/expands, state persists across reloads
- [ ] Cmd+K opens the palette; typing filters across files/tags/folders/storages
- [ ] Logout works (refresh cookie cleared, redirect to /login)
- [ ] Tweaks panel: changing theme/density/accent updates visuals live
- [ ] `tsc --noEmit` + `next lint --quiet` clean
- [ ] `docker compose build frontend` passes
- [ ] Responsive: at 375px width, sidebar is a drawer toggled by a hamburger

## Verification commands

```bash
cd /Users/zafareshmamatov/go/src/gitlab.com/docuemnt_manegment/frontend
./node_modules/.bin/tsc --noEmit && ./node_modules/.bin/next lint --quiet
cd .. && docker compose --profile dev up -d --build
# Manual: load each dashboard page in the browser, test Cmd+K, toggle theme
```

## Out of scope

- Don't migrate the page contents (that's Phase C). Pages still use their OLD internals; only the OUTER shell changes.
- No Team page content (Phase E)
- No Settings page content (Phase D3)
- No full-text search index — palette uses simple substring filter on existing endpoints
- No notifications list — the bell is a stub with a fake "0" badge
- Don't delete `app/(dashboard)` remnants — already removed
