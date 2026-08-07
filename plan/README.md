# Document Management System — Frontend Redesign Plan

This plan migrates our existing Next.js frontend to the new design system from the mockup (`/tmp/docmgmt-redesign/src/`), keeping all current backend endpoints and features working throughout the transition.

## Reference materials

- **Mockup HTML**: `/Users/zafareshmamatov/Downloads/index (1).html`
- **Mockup components**: `/tmp/docmgmt-redesign/src/*.jsx` (15 files, ~1,800 lines)
- **Current frontend**: `/Users/zafareshmamatov/go/src/gitlab.com/docuemnt_manegment/frontend/`
- **Backend API**: `/Users/zafareshmamatov/go/src/gitlab.com/docuemnt_manegment/backend/` (Go, chi router)

## Phases

| # | Phase | Effort | Blocks | Status |
|---|---|---|---|---|
| A | [Design foundation](phase-a-design-foundation.md) — tokens + primitives | 3–4 days | B, C, D | ☐ |
| B | [Shell + navigation](phase-b-shell-navigation.md) — sidebar, topbar, Cmd+K | 2–3 days | C, D | ☐ |
| C | [Existing pages redesign](phase-c-existing-pages-redesign.md) — Overview, Files, File Detail, Storages, Tags, Shares, Auth | 7–10 days | D1 | ☐ |
| D | [New pages, no backend change](phase-d-new-pages.md) — Search, Settings scaffold, Profile | 3–4 days | — | ☐ |
| E | [Backend-backed features](phase-e-backend-features.md) — Starring, Trash, Comments, Versions, Activity, Notifications, API tokens, Team | 2–4 weeks | — | ☐ |
| F | [AI features](phase-f-ai-features.md) — Ask document, AI summary, tag suggestions | 2–3 weeks | — | ☐ |
| G | [Polish](phase-g-polish.md) — virtualization, a11y, mobile, E2E | 3–5 days | A–D done | ☐ |

## Core constraints (apply to every phase)

1. **Don't break existing features.** At the end of each phase the app must run end-to-end (register → login → add storage → upload → folder → tag → share).
2. **New primitives live alongside old ones.** `components/ui/` gets new files in Phase A; existing components stay until a page is migrated to them.
3. **Backend is untouched in Phases A–D.** Only Phase E introduces new Go code.
4. **TypeScript strict, no `any`.** ESLint + `tsc --noEmit` must pass before each PR.
5. **Docker compose works.** `make up-dev` must produce a working stack after every phase.

## Execution model

Each phase file is self-contained and actionable. A fresh Claude Code session can open a phase file and execute it without needing prior context.

Phase files follow the same structure:

- **Context** — why this phase exists and what it builds on
- **Goals** — bullet list of outcomes
- **Tasks** — ordered, actionable steps with file paths
- **Acceptance criteria** — how we know the phase is done
- **Verification commands** — what to run before marking the phase complete
- **Out of scope** — what NOT to do in this phase

## Decisions already made

- **Router**: keep Next.js 14 App Router (no migration to Router v15 or Remix)
- **State**: TanStack Query for server state, Zustand for UI state (keep as-is)
- **Styling**: Tailwind CSS with CSS custom properties for theming (no CSS modules, no styled-components)
- **Icons**: `lucide-react` (already installed, covers the mockup's 40+ icons)
- **Charts**: start with custom SVG (mockup already shows how), add Recharts only if we need anything interactive
- **File viewers**: keep current `react-pdf` + custom image viewer; expand to multi-tab in File Detail
- **Typography**: Inter (sans) + JetBrains Mono (mono) via `next/font/google`

## Open questions (answer before starting Phase E/F)

- Do we want a multi-workspace model, or stay single-workspace per user? → Defaults to single for now; multi in Phase E.
- Full-text search: Meilisearch or Typesense? → Pick during E kickoff; both are acceptable.
- AI provider: Anthropic or OpenAI? → Plan assumes Anthropic (Claude) with streaming.

## Useful commands

```bash
# Frontend dev
cd frontend && npm run dev          # http://localhost:3000 (dev) or :3001 (docker)

# Type + lint gate (run before each PR)
cd frontend && ./node_modules/.bin/tsc --noEmit && ./node_modules/.bin/next lint --quiet

# Backend dev
cd backend && go run ./cmd/api      # requires .env + mongo running

# Full stack
make up-dev                          # docker compose + minio
make down

# Tests
cd backend && go test ./...
cd backend && TEST_MONGO_URI=mongodb://docmgmt:docmgmtpass@localhost:27018/?authSource=admin go test -tags=integration ./...
```
