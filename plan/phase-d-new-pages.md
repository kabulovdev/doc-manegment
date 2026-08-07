# Phase D — New pages, no backend change

**Effort**: 3–4 days
**Depends on**: Phase A (primitives), Phase B (shell)
**Blocks**: nothing

## Context

Add two new routes — **Search** and **Settings** — and a minimal `PATCH /auth/me` endpoint for profile updates. Everything else remains stub UI (marked "Coming soon"). The goal is to give the user visible feature surfaces so later phases (E/F) can plug in without restructuring.

Reference: `/tmp/docmgmt-redesign/src/page-search.jsx`, `page-settings.jsx`.

## Sub-phases

- [D1 — Search page](#d1--search-page) — 1.5 days
- [D2 — Settings scaffold + Profile tab](#d2--settings-scaffold--profile-tab) — 1.5 days
- [D3 — PATCH /auth/me backend](#d3-backend-patch-authme) — 0.5 day

---

## D1 — Search page

Route: `frontend/app/dashboard/search/page.tsx` (new).

### Layout

```
SectionHead: "Search" + Kbd shortcut hint
Large search input (auto-focus), Esc closes back to Overview
Scope tabs: Everything (default) | Files | Tags | Folders | People*
* People tab is disabled (Phase E — Team).
Results area (when query has value):
  Per-scope section with counts, grouped:
  - Files: card list with highlighted match, breadcrumb path, size, uploaded time, tags
  - Tags: pill grid
  - Folders: list with icon + path + file count
  Empty state if no results match
```

### Data

- Files scope: `filesApi.list({ q: query, limit: "50" })`
- Tags scope: `tagsApi.list()` + client-side `name.includes(query)` filter
- Folders scope: `foldersApi.list(null)` recursively (cache) + client-side name filter
- People scope: disabled until Phase E

### Highlighting

Simple function `highlight(text, query)` wraps matched substrings in `<mark class="bg-accent-soft text-accent">`.

### Keyboard

- Esc anywhere: back to previous page (`router.back()`)
- Tab navigates scope tabs
- Arrow keys navigate results (optional; keyboard-friendly is a plus)

### Shared code with palette

Extract the result rendering into `components/search/result-file.tsx`, `result-tag.tsx`, `result-folder.tsx` so Phase B's palette and this page share components.

---

## D2 — Settings scaffold + Profile tab

Route: `frontend/app/dashboard/settings/page.tsx` (new).

### Layout

```
SectionHead: "Settings"
[Left tab nav 220px] [Right content]
Tabs: Profile | Workspace | Security | API & Webhooks | Billing | Notifications
```

All tabs render a panel component. Profile is fully functional; the rest render "Coming soon" empty states with a short description of what the tab will include.

### Profile tab

File: `components/settings/profile-tab.tsx`.

- Fields: Avatar (initials for now, upload in Phase E), `display_name` (editable), `email` (read-only), Role badge ("Owner" stub), Language select (stub, defaults to English)
- Save button: `PATCH /auth/me` with `{display_name}`
- Success toast on save (use a new toast primitive — add to v2 if not present)

### Tab stubs

- **Workspace**: Card with workspace name input (read-only, single workspace), shareable URL (stub), default upload storage dropdown. "Expanded workspace management coming soon."
- **Security**: Password change form (already handled via login flow — stub for now), 2FA toggle (disabled), active sessions table (stub). "Full security controls in the next release."
- **API & Webhooks**: Empty state "API tokens and webhooks coming soon."
- **Billing**: Empty state "Billing will be available once subscriptions launch."
- **Notifications**: Empty state "Notification preferences coming soon."

---

## D3 — Backend: PATCH /auth/me

Small backend change to support the Profile tab's Save action.

### Files

- `backend/internal/http/dto/auth.go` — add `UpdateMeRequest { DisplayName string }`
- `backend/internal/http/handlers/auth.go` — add `UpdateMe` method
- `backend/internal/services/auth_service.go` — add `UpdateProfile(ctx, userID, displayName)`
- `backend/internal/ports/repository.go` — add `UpdateDisplayName` on `UserRepo`
- `backend/internal/adapters/mongo/user_repo.go` — implement method
- `backend/internal/http/router.go` — add `r.Patch("/me", authH.UpdateMe)`

### Validation

- Display name: 1–100 chars trimmed, no HTML-injectable characters (basic length check is enough; backend already hosts safe fields)

### Test

Update `auth_service_integration_test.go` to cover update + read back.

---

## Frontend API client update

Add to `frontend/lib/api/client.ts`:

```ts
export const authApi = {
  ...,
  updateMe: (input: { display_name: string }) =>
    apiFetch<User>("/auth/me", { method: "PATCH", body: JSON.stringify(input) }),
};
```

After a successful update, refresh the Zustand `authStore.user` so the sidebar shows the new name immediately.

---

## Acceptance criteria

- [ ] `/dashboard/search` renders, searching across files/tags/folders works (no full-text, substring only)
- [ ] `/dashboard/settings` renders with 6 tabs; Profile is functional; the rest show stubs
- [ ] Profile Save updates `display_name` server-side and in the UI
- [ ] `backend/go test ./...` passes (including the new test)
- [ ] `tsc --noEmit` + `next lint --quiet` clean
- [ ] Docker build succeeds for both services

## Verification commands

```bash
cd /Users/zafareshmamatov/go/src/gitlab.com/docuemnt_manegment/backend && go test ./...
cd /Users/zafareshmamatov/go/src/gitlab.com/docuemnt_manegment/frontend
./node_modules/.bin/tsc --noEmit && ./node_modules/.bin/next lint --quiet
cd .. && docker compose --profile dev up -d --build
# Manual: update display name in Settings → Profile; verify sidebar updates
# Manual: use /dashboard/search and Cmd+K palette
```

## Out of scope

- No full-text search index (Phase E if needed)
- No Team / Workspace management (Phase E)
- No API tokens (Phase E)
- No billing (out of this project scope for v1)
- No password change flow (Phase E — adds separate `PATCH /auth/password`)
