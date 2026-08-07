# Phase C — Existing pages redesign

**Effort**: 7–10 days
**Depends on**: Phase A (primitives), Phase B (shell)
**Blocks**: Phase D1 (Search) reuses Files filter logic

## Context

Every existing authenticated page gets rebuilt using v2 primitives and the new layout. Backend endpoints and data shape stay the same — we only touch `frontend/app/dashboard/**/*.tsx` and `frontend/components/**/*`.

Reference: `/tmp/docmgmt-redesign/src/page-overview.jsx`, `page-files.jsx`, `page-file-detail.jsx`, `page-storages.jsx`, `page-tags.jsx`, `page-shares.jsx`.

Work is sub-phased C1–C7; each sub-phase is independently shippable. Order matters only where a shared component (e.g. upload dialog) is built.

## Sub-phases

- [C1 — Overview](#c1--overview-dashboard) — 1 day
- [C2 — Files page](#c2--files-page-list--grid--folder-tree--bulk) — 2–3 days
- [C3 — File Detail](#c3--file-detail-page) — 1.5–2 days
- [C4 — Storages](#c4--storages) — 1 day
- [C5 — Tags](#c5--tags) — 0.5 day
- [C6 — Shares](#c6--shares) — 0.5 day
- [C7 — Auth pages](#c7--auth-pages) — 0.5 day
- [C8 — Upload dialog redesign](#c8--upload-dialog) — 0.5 day (shared dep of C2 and C3)

---

## C1 — Overview (dashboard)

File: `frontend/app/dashboard/page.tsx` (replace contents)

### Layout

```
SectionHead: "Welcome, {user.display_name}"
Stat cards grid (4 columns on desktop, 2 on tablet, 1 on mobile):
  - Connected storages
  - Total used (humanized bytes + object count)
  - Files
  - Active shares
Row: [Uploads chart card, 60%] [Storage by type card, 40%]
Row: [Recent activity card] [Storage health card] [Team card]  (3-col)
Row: Starred files grid (4-col) — show empty state for now
```

### Data sources (existing endpoints)

- `storagesApi.list()` → stat cards (count, sum `used_bytes`, `object_count`)
- `filesApi.list({ limit: "500" })` → file count, uploads sparkline (group by `created_at` day), MIME breakdown → SegBar
- `sharesApi.list()` → active shares count (filter `!revoked && !consumed && (!expires || future)`)
- **Recent activity / Team / Starred**: stub with empty-state cards ("Available in the next release")

### Uploads chart

Small SVG `Sparkbars` primitive showing the last 30 days' upload count. Compute client-side from `filesApi.list` results.

### Storage breakdown

SegBar primitive with segments per MIME category (PDF, images, docs, video, other). Colors from CSS vars (`--accent`, `--info`, `--violet`, `--warn`, `--text-3`).

---

## C2 — Files page (list + grid + folder tree + bulk)

Files: `frontend/app/dashboard/files/page.tsx`, `files/[folderId]/page.tsx`, shared view component `frontend/components/files/files-view.tsx` (rewrite).

### Layout

```
[Folder sidebar 240px] [Main content]
Main:
  Breadcrumb (Root › Documents › 2024)
  Header: "{folder name}"  +  meta ({N files, {total size}})
  Toolbar row:
    left: search input, type filter dropdown, tags filter dropdown, date filter dropdown
    right: view toggle (grid/list), sort select, upload button
  Bulk selection bar (appears when ≥1 selected): {N} selected • [Download] [Share] [Tag] [Move] [Delete] [Clear]
  Folders grid (children folders)
  Files list/grid
```

### Folder sidebar

Rewrite `components/folders/folder-tree.tsx`:

- Collapsible tree (chevron icons)
- Root folders pinned at top
- Each node: icon + name + chevron (if has children) + counts (file count hover only)
- **Smart views** section at bottom (stubbed placeholders — Phase E implements): "Starred", "Recently opened", "Shared with me", "Trash"

### List view

Columns: `[checkbox] [name + star + versions badge] [size] [tags] [mime] [folder] [uploaded] [menu ⋯]`

- Row height: `var(--row-h)` (respects density)
- Hover state: `bg-surface-2`
- Star toggle (left of name): stubbed for now (store locally until Phase E)
- Menu: View / Share / Rename / Move / Delete

### Grid view

- 200px min auto-fill
- Card: preview thumbnail area (file icon for non-images, `<img>` for images via blob fetch), name, size, uploaded time

### Filters

- Type: MIME categories (All / PDF / Images / Docs / Other) — client-side filter on the fetched list
- Tags: multi-select dropdown populated from `tagsApi.list()`, pass `?tag_id=` (one at a time for now)
- Date: Today / Last 7 days / Last 30 days — client-side filter
- Search: text input, debounced, maps to `?q=`

### Bulk actions

- Selection state: Zustand store (`uiStore.selectedFileIDs: Set<string>`)
- Bar appears when size > 0 at top of list
- Download: multi-fetch via existing `/content` endpoint → zip via `jszip` (install if not present) → single download trigger
- Share: opens Share dialog with the first selected file; TODO note for batch share
- Tag: opens tag picker that attaches to all selected via `filesApi.updateTags` one-by-one
- Move: opens a folder picker dialog; calls `PATCH /files/{id}` or a future `/move` endpoint (for now, use folder_id on the existing PATCH)
- Delete: confirmation, then `filesApi.remove` in parallel

### Files inside route

- `/dashboard/files` → `FilesView folderID={null}`
- `/dashboard/files/[folderId]` → `FilesView folderID={params.folderId}` with breadcrumb from `foldersApi.get(folderId)`

---

## C3 — File Detail page

File: rewrite `frontend/app/dashboard/files/view/[fileId]/page.tsx`.

### Layout

```
[Main 70%] [AI sidebar 30% — stub]
Main: tabs bar → active tab content
Tabs: Preview | Details | Versions | Activity | Comments
```

### Preview tab

Reuses existing PDF / image / text viewers. Loading via `fetchFileBlob()` (already implemented). Toolbar is part of the viewer.

### Details tab

- Metadata grid: name, type, size, folder (with link), storage (with link), object key (mono), uploaded, created_at
- Custom fields section: list existing + Add row inline
- Tags section: `TagPicker` + chips
- "Save" button persists via `filesApi.updateTags` and a new `/files/{id}` PATCH with `name`/`custom_fields` (already supported)

### Versions tab

Stub: empty state "Version history coming soon" (Phase E).

### Activity tab

Stub: "No activity recorded yet" (Phase E).

### Comments tab

Stub: "Comments will appear here once enabled" (Phase E).

### AI sidebar (right 30%)

Stub `<AiChatPanel />` that shows placeholder messages and a disabled textarea with "AI will be available soon" tooltip. Keep the component isolated so Phase F drops in a working version.

---

## C4 — Storages

File: `frontend/app/dashboard/storages/page.tsx` (rewrite).

### Layout

```
SectionHead: "Storages" + [Add storage] button
4 stat cards: Total capacity (sum or "—" if no quota), Used, Objects, Monthly cost (stub —)
Table: Name | Kind | Endpoint | Usage (progress bar) | Objects | Status | [⋯]
Row action menu: Test / Resync / Delete
Below table: 2-col grid
  [Sync queue card: last N events from `last_sync_at` / `last_error`]
  [Storage policies card: stub toggles — "Archive after 90d", "Encrypt at rest", "Replicate to...", "Delete from trash after 30d"] — Phase E
```

### Status badge

- Green "Active" if `last_error` empty and recently synced
- Amber "Stale" if `last_sync_at` > 24h ago
- Red "Error" if `last_error` present

### Add storage dialog

Rewrite in v2 primitives; same fields as current StorageForm. Inline provider chip row (R2 / S3 / MinIO). Auto-flip `force_path_style` when MinIO selected.

---

## C5 — Tags

File: `frontend/app/dashboard/tags/page.tsx` (rewrite).

### Layout

```
SectionHead: "Tags"
Create form: [name input] [color picker: 5 preset swatches + custom] [Create] — inline horizontal
Filter input (right side of header)
Table: Tag pill | Color hex | File count* | Created | [Edit] [Delete]
* file count: for each tag, query `filesApi.list({ tag_id })` with limit=1 and read total — OR better: add `GET /tags/{id}/stats` in Phase E. For now compute lazily on hover, or show "—".
```

---

## C6 — Shares

File: `frontend/app/dashboard/shares/page.tsx` (rewrite).

### Layout

```
SectionHead: "Shares"
Filter tabs: All (N) | Active (N) | Expired (N) | Revoked (N)
Table: Target | Status | Options (password/one-time chips) | Expires | Created | [Copy] [Revoke]
Right side: [Export CSV] button
```

### CSV export

Client-side: generate CSV from current rows, trigger download via blob URL + `<a download>`.

---

## C7 — Auth pages

Files: `frontend/app/(auth)/login/page.tsx`, `(auth)/register/page.tsx` — rewrite in v2.

Same fields, same submission logic. Only visual: center card, new font, accent button, tweaks panel accessible in auth flows.

---

## C8 — Upload dialog

File: `frontend/components/upload/upload-dialog.tsx` (rewrite).

Same behaviour as current implementation (multipart, progress). Visual changes:

- New dialog size: `md`
- Drag-drop zone styled with dashed border, accent on hover
- Progress bar uses `Progress` primitive
- Custom fields editor uses compact density
- Tag picker uses `Pill` primitive
- Storage picker: visual chip row (provider icon + display name), not dropdown

Accept external props for default folder_id (used by Files page when uploading inside a subfolder).

---

## Acceptance criteria (whole Phase C)

- [ ] Every sub-phase merged individually, app working after each
- [ ] No page uses legacy `components/ui/button.tsx` / `input.tsx` / `dialog.tsx` — only v2
- [ ] `tsc --noEmit` + `next lint --quiet` clean throughout
- [ ] Manual end-to-end still works: register → login → add MinIO storage → create folder → upload PDF → view → share → open share in incognito (unchanged backend, so this must still work)
- [ ] Legacy primitives (`ui/button.tsx`, `ui/input.tsx`, `ui/dialog.tsx`) can be deleted after C is done — verify with grep

## Cleanup at end of phase

```bash
# After every page migrated, confirm no imports of old primitives
grep -r "from \"@/components/ui/button\"" frontend/app frontend/components
grep -r "from \"@/components/ui/input\"" frontend/app frontend/components
grep -r "from \"@/components/ui/dialog\"" frontend/app frontend/components
# If empty, delete the old files:
rm frontend/components/ui/button.tsx frontend/components/ui/input.tsx frontend/components/ui/dialog.tsx
# Optionally move v2/* → ui/* and update imports with a single sed pass
```

## Out of scope

- No new backend endpoints (except one optional: `PATCH /auth/me` for display_name — that's Phase D2)
- Don't implement starring/trash/versions/comments backends
- Don't add real AI to the sidebar stub
- No notification counts beyond a hardcoded "0"
