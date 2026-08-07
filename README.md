# Document Management System

Self-hosted document management where each user connects their own S3-compatible object storage (Cloudflare R2, AWS S3, MinIO) and uses the platform to upload, organise (folders, tags, custom fields), and share files via expiring view-only links.

The platform holds metadata and orchestrates; your bucket holds the bytes.

## Stack

- **Backend**: Go 1.24+, chi router, `mongo-go-driver`, `aws-sdk-go-v2` (S3), argon2id, slog
- **Database**: MongoDB 7
- **Frontend**: Next.js 14 App Router + TypeScript + Tailwind + shadcn-style primitives, TanStack Query, Zustand
- **Object storage**: S3-compatible adapter (R2 / AWS S3 / MinIO via `BaseEndpoint` + `UsePathStyle`)
- **Deployment**: Docker Compose + Caddy reverse proxy

## Features

1. Register / login with argon2id + JWT access tokens and rotating refresh tokens in HttpOnly cookies
2. Connect multiple object storages per user; credentials encrypted at rest (AES-GCM, per-field nonce)
3. SSRF-hardened endpoint validation with DNS rebinding mitigation (re-verifies IP at connect time)
4. Upload files with storage picker, 8 MB multipart parts, 4 parallel workers, progress bar
5. Hierarchical folders (ID-based materialised path: move operation rewrites the subtree)
6. Per-user tags attachable to files or folders
7. Custom fields per file (key/value/type), indexed for search
8. View-only share links with optional password (argon2id), expiration, and one-time-use
9. Public viewer uses PDF.js and a hardened image viewer (right-click blocked, `Ctrl+S`/`Ctrl+P` blocked, `Content-Disposition: inline`)
10. Dashboard with storage usage cards, resync button, access logs

## Quick start

```bash
cp .env.example .env
# Generate proper secrets:
# MASTER_KEY   — openssl rand -base64 32
# JWT_SECRET   — openssl rand -base64 48

make up-dev      # starts backend + frontend + mongodb + minio (dev profile)
```

Services after `up-dev`:

- Frontend: http://localhost:3000
- Backend: http://localhost:8080/api/v1
- MongoDB: localhost:27018 (host port, to avoid colliding with another Mongo on :27017)
- MinIO: http://localhost:9000 (API), http://localhost:9001 (console)

For local testing with MinIO:

```bash
docker compose exec -T minio mc alias set local http://localhost:9000 minio minio12345
docker compose exec -T minio mc mb --ignore-existing local/test-bucket
```

Then in the frontend, add a storage config with:

- Provider: MinIO
- Endpoint: `http://localhost:9000` (allowed in dev via `ALLOW_HTTP_ENDPOINTS=true`)
- Bucket: `test-bucket`
- Access / Secret: `minio` / `minio12345`
- Force path-style: on

## Repository layout

```
backend/              Go service (hexagonal: domain, ports, adapters, services, http)
frontend/             Next.js 14 app
docker-compose.yml    mongo, backend, frontend, minio (dev profile), optional caddy
Makefile              common commands
```

Key files inside backend:

- `internal/ports/` — all interfaces (repos, storage provider, encryptor)
- `internal/adapters/mongo/` — repo implementations
- `internal/adapters/storage/s3compat/` — single S3-compatible implementation for R2/S3/MinIO
- `internal/adapters/storage/ssrf/` — endpoint validation + DNS-rebinding-safe `http.Transport`
- `internal/adapters/crypto/aesgcm.go` — per-field AES-GCM encryption for storage credentials
- `internal/services/` — application use cases (auth, storage, file, folder, tag, share)

## Security posture

- **No presigned URLs are ever issued to clients.** All uploads and downloads flow through the backend. This keeps credentials server-side and makes quota bookkeeping + SSRF guard reachable.
- **Credential encryption at rest.** `MASTER_KEY` is a 32-byte base64 value. Each ciphertext uses a fresh nonce. `key_version` is stored for future key rotation.
- **SSRF guard.** User-supplied storage endpoints are resolved and rejected if they hit RFC1918, loopback, link-local, CGNAT, or IPv6 ULA ranges. A custom `http.Transport.DialContext` re-verifies the resolved IP at connect time, closing the DNS-rebinding window.
- **Passwords** are argon2id (`time=3, memory=64MB, parallelism=2`).
- **Refresh tokens** are opaque 32-byte values stored hashed (SHA-256), with family + rotation + reuse detection; logout revokes the family.
- **Share tokens** are 32 random bytes, base64url. One-time use is an atomic Mongo CAS. Password gating uses argon2id and sets a short-lived HttpOnly session cookie scoped to the share path.
- **Share responses** set `Content-Disposition: inline`, `Cache-Control: private, no-store`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`.
- **Screenshots cannot be prevented.** The public viewer UI says so explicitly. Treat share links as a deterrent, not a guarantee.

## Testing

```bash
make backend-test
```

Runs unit tests for the SSRF guard and folder path parsing.

End-to-end smoke tests exist as shell commands documented in the internal Phase 1–4 logs; a Playwright suite is planned for the next milestone.

## Environment variables

See `.env.example`. Required: `MASTER_KEY` (32-byte base64), `JWT_SECRET` (≥32 chars). Everything else has sensible defaults. Set `ALLOW_HTTP_ENDPOINTS=true` in development so you can point to a local MinIO over HTTP; production should leave this `false` so user-supplied endpoints must be HTTPS.
