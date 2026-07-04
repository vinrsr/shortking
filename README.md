# ShortKing

A URL shortener with custom aliases, click analytics, link expiration, max-click limits, and QR codes. Shorten links with no account (free, temporary), or sign up for permanent links, custom aliases, analytics, and QR codes.

## Features

- **Anonymous shortening**: no login required, auto-generated code, fixed 2-day expiry, rate-limited (burst + daily cap) to encourage signup.
- **Account links**: custom aliases, custom or no expiry, max-click limits, click analytics (timestamp, referrer, user agent), QR codes.
- **Auth**: email/password signup and login, JWT access + refresh tokens (rotated on refresh), stored as httpOnly cookies set by the Next.js BFF layer (the browser never sees the raw tokens).
- **Email verification**: required before a logged-in user can create a link (anonymous shortening is unaffected). Verification and password-reset emails are sent via [Resend](https://resend.com); if no API key is configured, the link is logged to the console instead (useful for local dev without email configured).
- **Forgot / reset password**: token-based, single-use, 1-hour expiry. Doesn't reveal whether an email is registered.
- **QR codes**: generated on demand (PNG, not stored as a file) from the link's short URL. Once generated, a link remembers that state so the dashboard keeps showing it after a refresh instead of re-prompting.
- **Expired / deactivated / unknown short links**: redirect to a branded frontend page (`/link-expired`, `/link-not-found`) instead of raw JSON, worded for whoever clicked the link, not its owner.
- **Public stats**: aggregate counts (total/active links, total clicks, total users, total QR codes generated) shown on the landing page.
- **Rate limiting**: per-route, Redis-backed (login/signup, anonymous shorten burst + daily cap, authenticated link creation, redirects).

## Stack

- **apps/web**: Next.js (App Router, TypeScript, Tailwind v4). Route Handlers under `src/app/api/**` act as a BFF: they hold the httpOnly auth cookies and proxy to the Go API, so client components never talk to the API directly.
- **apps/api**: Go (Gin, GORM, PostgreSQL, Redis). Layered `handler -> service -> repository`.
- Monorepo managed with Turborepo + pnpm.

## Directory layout

```
apps/
  api/
    cmd/server/        entrypoint (main.go)
    internal/
      handler/          HTTP handlers (Gin)
      service/          business logic
      repository/       GORM data access
      middleware/       auth, CORS, rate limiting
      mailer/           Resend email client
      cache/            Redis (sessions, rate limits, link cache)
      config/           env var loading
      models/           GORM models
    migrations/          golang-migrate SQL migrations
    pkg/qrcode/          QR PNG generation
  web/
    src/
      app/
        (marketing)/     landing page
        (auth)/          login, signup, forgot/reset password, verify-email
        (dashboard)/     authenticated dashboard (links, QR codes, account)
        api/             Next.js Route Handlers (BFF proxy to the Go API)
        link-expired/, link-not-found/   branded redirect-error pages
      components/        shared UI
      lib/                session/cookie helpers, API client
```

## Prerequisites

- Node.js 20+, pnpm (`corepack enable`)
- Go 1.23+
- Docker (for local Postgres + Redis)
- [air](https://github.com/air-verse/air), `go install github.com/air-verse/air@latest` (Go live reload, used by `pnpm dev`)
- [golangci-lint](https://golangci-lint.run/), used by `pnpm lint`
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI, used to run DB migrations

## Getting started

```bash
pnpm install
docker compose up -d postgres redis
migrate -database "$DATABASE_URL" -path apps/api/migrations up
pnpm dev
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8080

Copy `apps/api/.env.example` to `apps/api/.env` and `apps/web/.env.local.example` to `apps/web/.env.local` before running, and fill in the values for your own setup.

## API overview

All authenticated routes require `Authorization: Bearer <accessToken>`.

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/api/v1/auth/signup` | none | Sends a verification email |
| POST | `/api/v1/auth/login` | none | |
| POST | `/api/v1/auth/refresh` | none | Rotates the refresh token |
| POST | `/api/v1/auth/logout` | none | |
| POST | `/api/v1/auth/forgot-password` | none | Always 200; doesn't reveal if the email exists |
| POST | `/api/v1/auth/reset-password` | none | |
| POST | `/api/v1/auth/verify-email` | none | |
| POST | `/api/v1/auth/resend-verification` | none | Always 200 |
| GET | `/api/v1/me` | required | |
| POST | `/api/v1/shorten` | none | Anonymous shorten; rate-limited (burst + daily cap) |
| GET | `/api/v1/stats` | none | Public landing-page stats |
| POST | `/api/v1/links` | required + verified email | Custom alias, expiry, max-clicks |
| GET | `/api/v1/links` | required | List own links |
| GET / PATCH / DELETE | `/api/v1/links/:id` | required | Must own the link |
| GET | `/api/v1/links/:id/qrcode` | required | Returns a PNG, generated on demand |
| POST | `/api/v1/links/:id/qrcode/generations` | required | Records that a QR was generated (persisted + counted once) |
| GET | `/:code` | none | Redirect; expired/missing/deactivated codes redirect to `/link-expired` or `/link-not-found` on the frontend |

## Database migrations

Located in `apps/api/migrations`, run with `golang-migrate`:

```bash
migrate -database "$DATABASE_URL" -path apps/api/migrations up      # apply
migrate -database "$DATABASE_URL" -path apps/api/migrations down 1  # roll back one
```

To add a new one: `migrate create -ext sql -dir apps/api/migrations -seq <name>`.

## Testing & linting

```bash
pnpm lint   # golangci-lint (api) + eslint (web)
pnpm test   # go test ./... (api) + vitest run (web)
pnpm build  # go build (api) + next build (web)
```
