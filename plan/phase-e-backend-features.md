# Phase E — Backend-backed features

**Effort**: 2–4 weeks (parallelisable)
**Depends on**: Phases A, B, C (new design)
**Blocks**: nothing (but Phase F's AI features benefit from Activity log + Versioning)

## Context

Each feature below is an independent task that pairs a backend change (new collection, new endpoint) with a UI change (a Phase C stub becomes real). Features can be tackled in any order; priority should match product needs.

Reference: mockup shows each feature's UI in the corresponding page.

## Feature map

- [E1 — Starring](#e1--starring) — 0.5 day
- [E2 — Trash (restore, auto-purge)](#e2--trash) — 1–1.5 days
- [E3 — Comments](#e3--comments) — 2–3 days
- [E4 — File versions](#e4--file-versions) — 3–4 days
- [E5 — Activity log](#e5--activity-log) — 1.5–2 days
- [E6 — Notifications](#e6--notifications) — 2–3 days
- [E7 — API tokens](#e7--api-tokens) — 1.5–2 days
- [E8 — Sessions + password change + 2FA](#e8--sessions--password-change--2fa) — 2–3 days
- [E9 — Team / workspaces](#e9--team--workspaces) — 4–5 days
- [E10 — Full-text search index](#e10--full-text-search-index) — 2–3 days
- [E11 — Storage policies](#e11--storage-policies) — 2–3 days

---

## E1 — Starring

### Backend

- `backend/internal/domain/file.go` — add `Starred bool \`bson:"starred"\``
- `backend/internal/adapters/mongo/file_repo.go` — add `SetStarred(userID, id, starred)`
- `backend/internal/services/file_service.go` — expose `ToggleStarred`
- `backend/internal/http/handlers/files.go` — add `POST /files/{id}/star` body `{starred:bool}`
- Mongo index: `{user_id: 1, starred: 1, updated_at: -1}` (sparse)

### Frontend

- `filesApi.toggleStar(id, starred)`
- Star icon in file list row + Files view "Starred" smart view in sidebar
- Starred files grid on Overview page

---

## E2 — Trash

Existing soft-delete already marks `status: "deleted"`. This adds Restore + auto-purge.

### Backend

- `POST /files/{id}/restore` — flips status back to `ready` if the user owns it and it's in deleted state
- `GET /files?status=deleted` — already supported via `FileListFilter.Status`
- Scheduled purge: standalone `backend/cmd/purge/main.go` cron (daily) that deletes objects + documents where `status=deleted && updated_at < now-30d`
  - Uses storage provider `Delete` to remove the object from the bucket
  - Logs per-file outcome

### Frontend

- Files sidebar: "Trash" smart view → routes to `/dashboard/files?trash=1` → passes `status: "deleted"`
- Row context menu: Restore + Delete forever
- Banner: "Items in Trash are deleted after 30 days."

### Tests

- Integration: create → soft-delete → restore → list → verify status

---

## E3 — Comments

### Backend

- New collection `file_comments`:
  ```
  { _id, user_id, file_id, body (markdown), parent_id (threading), created_at, edited_at, deleted_at }
  ```
- Indexes: `{file_id: 1, parent_id: 1, created_at: 1}`, `{user_id: 1}`
- Endpoints:
  - `POST /files/{id}/comments` body `{body, parent_id?}`
  - `GET /files/{id}/comments?cursor=`
  - `PATCH /comments/{id}` body `{body}`
  - `DELETE /comments/{id}` (soft delete by setting `deleted_at`)
- Auth: only author can edit/delete; any file owner collaborator can view/post (for now, only the file owner)

### Frontend

- `commentsApi` in `lib/api/comments.ts`
- File Detail → Comments tab: thread viewer, reply form, edit/delete menu
- Markdown rendering (use `marked` or `react-markdown`)

---

## E4 — File versions

Strategy: on re-upload with the same file ID (or new endpoint `/files/{id}/new-version`), keep the previous object_key and link it.

### Backend

- New collection `file_versions`:
  ```
  { _id, file_id, version_number, object_key, size_bytes, checksum, created_by, created_at }
  ```
- On upload-complete, if `?replaces={fileID}` is present: save the previous `files.object_key` + metadata into `file_versions`, update `files` with the new data
- Endpoints:
  - `GET /files/{id}/versions`
  - `POST /files/{id}/versions/{versionID}/restore` — swap current with historical
- Storage: keep all versions in the bucket (don't delete on new upload)

### Frontend

- File Detail → Versions tab: list with version number, uploaded_at, size, Restore button
- Upload dialog on a file detail page offers "Replace as new version" mode

### Open decisions

- Bucket object tagging vs. our own `file_versions` collection: go with our collection for portability (no reliance on S3 versioning support, which R2 doesn't have in the same API)

---

## E5 — Activity log

### Backend

- New collection `activity_logs`:
  ```
  { _id, user_id, subject_type: "file"|"folder"|"tag"|"storage"|"share", subject_id, action, metadata, created_at, ip, user_agent }
  ```
- Middleware in handlers layer records: create/delete/rename/move/share/restore/upload-complete
- TTL: 180 days
- Endpoints:
  - `GET /activity?subject_id=...&limit=...`
  - `GET /activity/recent?limit=20` — user-wide feed

### Frontend

- File Detail → Activity tab: per-file feed
- Overview → Recent activity card: user-wide feed (replaces the Phase C1 stub)

---

## E6 — Notifications

### Backend

- Collection `notification_prefs` per user: `{user_id, event: "share_opened"|"comment_reply"|..., channels: {email, push, inapp}}`
- Collection `notifications`: `{user_id, type, subject, body, read_at, created_at}`
- Trigger points: on share access, on comment reply (when Phase E3 merged)
- Endpoints:
  - `GET /notifications?unread_only=` — inbox
  - `POST /notifications/mark-read` — bulk
  - `GET /notifications/prefs`, `PATCH /notifications/prefs`
- Delivery: in-app only for now. Email can be added later via SMTP/SendGrid.

### Frontend

- Topbar bell: real unread count (query every 60s)
- Bell click → dropdown panel with recent notifications + "Mark all read"
- Settings → Notifications tab: prefs table

---

## E7 — API tokens

### Backend

- Collection `api_tokens`: `{user_id, name, hash (sha256), prefix (first 8 chars for display), scopes: [], created_at, last_used_at, revoked_at}`
- Endpoints:
  - `POST /auth/tokens` body `{name, scopes?}` → returns the plaintext ONCE
  - `GET /auth/tokens`
  - `DELETE /auth/tokens/{id}`
- Middleware: accept `Authorization: Bearer doc_XXXXX` (prefix `doc_` distinguishes from JWT). Look up by hashed token. On match, set `userID` in context with the API token's scopes.
- Scopes: `files:read`, `files:write`, `storages:read`, `shares:write`, etc.

### Frontend

- Settings → API & Webhooks tab: table with create dialog, copy-on-create modal, revoke
- Shows masked token (`doc_**********${lastFour}`) in list

### Docs

- Add a small API reference section to README

---

## E8 — Sessions + password change + 2FA

### Sessions

- Existing `refresh_tokens` collection already tracks per-device; expose to user:
  - `GET /auth/sessions` — list current family members (user-agent, IP, created_at, last-used hint)
  - `DELETE /auth/sessions/{id}` — revoke a single session (not family)

### Password change

- `PATCH /auth/password` body `{current_password, new_password}` — argon2id verify current then hash new
- Force logout of all sessions except current

### 2FA

- TOTP (RFC 6238):
  - `POST /auth/2fa/setup` → returns otpauth URL + secret (QR code rendered client-side with `qrcode`)
  - `POST /auth/2fa/enable` body `{code}` — verifies, sets `users.totp_secret`
  - `POST /auth/2fa/disable` body `{code}`
  - Login flow: if 2FA enabled, after password verify return `{requires_totp: true}`; client submits second POST `/auth/login/totp`

### Frontend

- Settings → Security tab: password change form, sessions table with Revoke, 2FA toggle + QR code + verify step

---

## E9 — Team / workspaces

Major feature — single-user app becomes multi-tenant inside a workspace.

### Data model

- Collection `workspaces`: `{_id, name, slug, owner_id, created_at, plan}`
- Collection `workspace_members`: `{workspace_id, user_id, role: "owner"|"admin"|"member"|"viewer", created_at}`
- Existing user-scoped collections gain `workspace_id`: `storage_configs`, `folders`, `files`, `tags`, `share_links`
- Migration: for each existing user, create a personal workspace, stamp `workspace_id` on their data

### Endpoints

- `POST /workspaces`, `GET /workspaces` (user's workspaces), `POST /workspaces/{id}/invites`, `DELETE /workspaces/{id}/members/{userID}`
- JWT payload gains active `workspace_id` claim
- Switching workspace = reissue JWT

### Frontend

- Workspace switcher in sidebar is now functional
- `/dashboard/team` page: members table, invite form, role management
- All API calls gain `X-Workspace-Id` header (or scope derived from JWT)

### Out of scope for E9

- Billing per workspace
- Row-level ACLs finer than workspace — deferred

---

## E10 — Full-text search index

### Stack

- Meilisearch (simpler than Typesense for our scale)
- Docker compose service added

### Sync

- On `files` create/update/delete: publish to Meilisearch index `files` with `{id, user_id, workspace_id, name, mime, tags, custom_fields_text, content_text}`
- Optional text extraction: for PDFs, extract first 10 MB worth of text with `pdftotext` sidecar; for office docs, `tika`. Async job queue via a simple in-Go goroutine pool backed by Mongo queue collection.

### Endpoint

- `GET /search?q=&scope=files|tags|folders&limit=` — proxies to Meilisearch; returns ranked results with highlight spans

### Frontend

- Search page upgrades from substring-match to this endpoint with snippet highlighting

---

## E11 — Storage policies

### Backend

- Collection `storage_policies`: `{storage_id, archive_after_days, encrypt_at_rest, replicate_to_storage_id, trash_purge_days}`
- Scheduled worker applies rules (re-run at midnight)
- Endpoints: `GET/PATCH /storages/{id}/policies`

### Frontend

- Storages page → Policies card is now live with toggles + input for days

---

## Acceptance criteria (per feature)

Each feature above has its own acceptance criteria:

1. New backend endpoints have an integration test under `backend/internal/services/` with `//go:build integration`
2. Frontend tab/UI is wired to real API (no stub), and existing pages still work
3. `go test ./...` and `tsc --noEmit && next lint --quiet` clean
4. End-to-end smoke test in a browser

## Priority recommendation

If you can only pick three for the next iteration:

1. **E1 Starring** + **E2 Trash** (high user value, low effort)
2. **E5 Activity log** (feeds Overview + File Detail's stub tab)
3. **E7 API tokens** (enables programmatic access — unblocks integrations)

If you have more capacity:
- **E3 Comments**, **E4 Versions** together turn File Detail into a real collaboration surface
- **E9 Team** is the biggest single feature and opens the door to multi-user

## Out of scope (never)

- Real-time collaborative editing (Docs-style)
- On-device encryption (E2E)
- Anti-virus scanning (needs third-party)
