# Phase G — Polish

**Effort**: 3–5 days
**Depends on**: Phases A–D complete (and ideally key E items shipped)
**Blocks**: nothing

## Context

After the redesign and key features ship, we tighten the app for production: performance on large datasets, accessibility, mobile responsiveness, empty/error states, animations, and an end-to-end test suite.

## Tasks

### G1 — Table virtualization

Large file/tag lists need virtual scrolling to stay smooth at 10k+ rows.

- Install `@tanstack/react-virtual`
- Wrap the Files list and Tags table with a virtualizer
- Keep row height = `var(--row-h)` so density still works
- Lazy-load next pages when near the bottom (`filesApi.list({ skip, limit: 100 })`)

### G2 — Accessibility audit

- Focus-visible rings on every interactive element (Tailwind `focus-visible:ring-2 focus-visible:ring-accent`)
- ARIA labels on icon-only buttons
- Sidebar: `aria-current="page"` on active link
- Dialog: focus trap + restore on close (already partial — verify)
- Color contrast: run axe or Lighthouse; fix any AA failures (common culprits: `--text-3` on `--surface-2`)
- Keyboard shortcuts documented in a `?` help panel

### G3 — Mobile responsiveness

- Sidebar → drawer below 768px (already begun in Phase B)
- Topbar compresses, collapses search to icon-only
- Files table → card list below 640px (no horizontal scroll)
- File Detail: AI sidebar becomes a bottom sheet on mobile; tabs become a scrollable row

### G4 — Empty and error states

Audit every page for:
- **Empty state**: no data yet (clear CTA + link to relevant page)
- **Loading state**: Skeleton primitive (already built in Phase A)
- **Error state**: error message with retry button (not a page-level crash)

Pages to cover: Overview (no storages / no files), Files (folder empty), File Detail (file deleted since URL cached), Storages, Tags, Shares, Search (no results).

### G5 — Subtle animations

- Sidebar collapse transition: 180ms ease-out
- Dialog open: fade + scale from 0.96 to 1
- Tabs: underline slide
- Toasts: slide-in from top-right, dismiss after 4s
- **No bounce, no rubber-band, no parallax** — keep it tight

Use `framer-motion` only if native CSS transitions aren't enough; prefer CSS.

### G6 — End-to-end tests

Install Playwright: `cd frontend && npm install -D @playwright/test && npx playwright install`.

Create `frontend/e2e/` with:

- `auth.spec.ts` — register → login → logout
- `files.spec.ts` — add storage → create folder → upload PDF → tag → share → view via share link (incognito context)
- `search.spec.ts` — upload files → open Cmd+K → find by name
- `tweaks.spec.ts` — toggle theme/density/accent, reload, verify persistence

Add `make test-e2e` target:

```makefile
test-e2e:
	cd frontend && npx playwright test
```

Run in CI against a live docker compose stack; or locally via `make up-dev && make test-e2e`.

### G7 — Logging and observability

- Backend: structured request log (already via chi middleware) — ensure it includes `request_id`, `user_id`, `workspace_id`
- Frontend: surface fetch errors with a global error boundary + toast (install `sonner` or build on existing Dialog primitive)
- Optional: Sentry SDK stubbed (off by default) for crash reporting

### G8 — Performance budget

Measure + enforce:

- Largest Contentful Paint < 2.5s on slow 4G (Lighthouse)
- Bundle size ceiling: set `next.config.mjs` to fail build if main bundle > 400 KB gzipped
- Install `@next/bundle-analyzer`, inspect, lazy-load heavy deps (PDF viewer is already dynamic)

### G9 — Documentation updates

- Update `README.md` with:
  - Screenshots of the redesigned UI (light + dark)
  - Updated feature list (everything that shipped in A–E)
  - Updated API reference
- Update `plan/README.md` status column to mark completed phases

## Acceptance criteria

- [ ] Lighthouse scores ≥ 90 for Performance, Accessibility, Best Practices on Overview and Files pages (light + dark)
- [ ] axe-core: 0 critical issues on every authenticated page
- [ ] 10,000 file fixtures render smoothly (no frame drops scrolling)
- [ ] 4 Playwright specs green in CI
- [ ] Main JS bundle ≤ 400 KB gzipped
- [ ] Mobile usability: every feature reachable with touch on a 375px viewport

## Verification commands

```bash
cd frontend && npx lighthouse http://localhost:3001/dashboard --view
cd frontend && npx playwright test
cd frontend && ANALYZE=true npm run build
```

## Out of scope

- Offline / PWA (separate track)
- Multi-language i18n (separate track)
- Server-rendered optimisations beyond what Next.js gives us
