# Workend v2

Self-hosted, browser-based developer workstation.

Pick a git repo, see what's runnable in it, run those tasks reproducibly in isolated containers, watch the output live, schedule them, build images from Dockerfiles, get notified on failures, keep a log of every run.

## Status: v0.5.0

Stages 0–22 of [PLAN.md](PLAN.md) are complete. The system covers the full lifecycle of "manage shared repos across providers, run things in them, see and compare results, schedule and webhook-trigger them, and back up the whole thing."

See [OBJECTIVE.md](OBJECTIVE.md) for vision, [PLAN.md](PLAN.md) for the staged build plan, [ARCHITECTURE.md](ARCHITECTURE.md) for technical design.

## Running it

Requirements: Docker (or Rancher Desktop), 4 GB free disk for images.

```sh
cd v2
docker compose up --build
```

First boot pulls the postgres, dagger-engine, node, and golang base images and builds the api + web — expect 5–10 minutes.

When the stack is up:

- Open <http://localhost:3000>
- Sign up with any email + password
- Create a workspace, then add a project by public git URL (e.g., `https://github.com/sveltejs/kit.git`)
- Once the clone finishes, runnable tasks appear under the project (npm scripts, justfile recipes, Dockerfile builds)
- Click **Run** to execute one in a Dagger-managed container; watch live logs, cancel mid-run
- Browse run history, re-run, view code stats, build images, schedule recurring runs
- Configure notification targets in **Settings → Notifications** to get pinged on failures

To stop: `docker compose down`. Add `--volumes` to wipe the database, repo cache, and logs.

## Optional configuration

All optional — Workend works fine without them.

| Env var | What it enables |
|---|---|
| `WORKEND_GITHUB_CLIENT_ID` + `_SECRET` + `WORKEND_TOKEN_KEY` | GitHub OAuth → private repo cloning |
| `WORKEND_WEB_PUBLIC_URL` | Run-detail link in notifications (defaults to localhost:3000) |
| `WORKEND_SMTP_HOST/_FROM/_USERNAME/_PASSWORD` | Email notifications |

`WORKEND_TOKEN_KEY` is a base64-encoded 32-byte key for at-rest encryption of stored OAuth tokens. Generate with:

```sh
openssl rand -base64 32
```

## What's in v0.5.0

**Core (Stages 0–7, MVP)**
- Local password auth (argon2id), session cookies
- Workspaces and projects
- Repo ingestion via Dagger pipeline (public HTTPS git)
- Task autodetection: `package.json` scripts, `justfile` recipes, `Dockerfile`
- Task execution in pinned container images
- Server-Sent Events for live log tail; cancel mid-run
- Run history with status filter; re-run from any past run

**Post-MVP (Stages 8–14)**
- Code statistics via `tokei` in a Dagger container
- GitHub OAuth + private repo cloning (token encrypted at rest)
- Audit log, concurrent-run limits, admin role + page
- Scheduled runs via cron (5-field), in-process scheduler
- Cross-project dashboard with per-project health cards
- Image builds via Dagger BuildKit
- Notifications: webhook, Slack webhook, email (SMTP)

**Multi-provider + advanced (Stages 15–22)**
- GitLab + Gitea OAuth alongside GitHub; multiple instances per provider
- Repo browser when adding a project (pick from any connected provider)
- Diff view between two runs of the same task
- Workspace sharing: invite by email, owner / member roles, leave / remove
- Webhook-triggered syncs (per-project token URL)
- Backup / restore scripts (pg_dump + volume tars)
- Pluggable custom detectors via JSON config (Makefile / Cargo / Taskfile out of the box)
- Dagger task source: `dagger.json` projects expose their functions as runnable tasks

## What's not here

- Bitbucket OAuth (the extension point exists — implement the `oauth.Provider` interface)
- Image push to a registry (local builds only; runs record digest + size)
- Discord / Teams notification channels
- True line-by-line live log streaming (see "Honest limitations")
- Provider webhook signature verification (token-in-URL is the auth)
- Per-user disk quotas (concurrent-run limit is enforced; disk is not)
- Repo browser search / filter (paging only)

## Honest limitations in v0.5.0

- **Live log streaming is structurally complete but materially batched.** Dagger v0.13's `container.Stdout()` only returns once the container exits. SSE plumbing is correct; expect a single large delivery at completion. Stage 23 territory.
- **Re-run uses current repo HEAD**, not the original run's commit.
- **Single-instance scheduler.** The cron ticker runs in the api process. Don't horizontally scale the api yet — schedules will fire multiple times.
- **No per-user disk quotas.** Concurrent-run limit is enforced (3 active per user); disk is not.
- **Image build size is not reported.** Without a registry to inspect against, the size we'd report would be the exported tarball size.
- **Webhook auth is the URL token.** No per-provider HMAC signature verification yet (Stage 24).
- **Self-signed TLS on self-hosted Gitea/GitLab** isn't supported out of the box — Workend's HTTP client doesn't bundle custom CAs. Mount one into the api container if you need it.
- **Token refresh is lazy-on-expiry**, not lazy-on-401. If a provider revokes a token mid-validity-window you'll see a clone failure rather than an automatic refresh + retry.

## Stack

- **Frontend:** SvelteKit 2 + Svelte 5, Node adapter
- **API:** Go 1.24 + chi router + pgx/v5 + goose migrations + robfig/cron
- **Database:** PostgreSQL 16
- **Job runner:** Dagger engine v0.13.7 + Dagger Go SDK
- **Image builds:** Dagger / BuildKit
- **Realtime:** Server-Sent Events
- **Encryption-at-rest:** NaCl secretbox (golang.org/x/crypto)
- **Orchestration:** Docker Compose

All non-engine containers run as non-root (uid 10001) with all capabilities dropped. The Dagger engine is the only privileged container — required by BuildKit, isolated to its own service.

## Layout

```
v2/
├── api/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── admin/       admin endpoints + RequireAdmin middleware
│   │   ├── audit/       best-effort audit-log writer
│   │   ├── auth/        password, sessions, signup/login/logout
│   │   ├── config/      env-driven config
│   │   ├── dagger/      lazy SDK client wrapper
│   │   ├── dashboard/   per-project health aggregation
│   │   ├── db/          pgx pool + embedded goose migrations
│   │   ├── detect/      task autodetection (npm, just, dockerfile)
│   │   ├── health/      /healthz
│   │   ├── image/       built-image listing
│   │   ├── notify/      webhook/slack/email dispatcher
│   │   ├── oauth/       GitHub OAuth flow
│   │   ├── project/     project CRUD + clone-async pipeline
│   │   ├── repo/        Dagger-based git clone
│   │   ├── run/         task execution + SSE log streaming + image builds
│   │   ├── schedule/    cron schedules + ticker goroutine
│   │   ├── secret/      NaCl secretbox wrapper
│   │   ├── server/      chi router wiring
│   │   ├── stats/       tokei-driven code stats
│   │   ├── task/        task listing
│   │   └── workspace/   workspace CRUD
│   └── Dockerfile       multi-stage; installs dagger CLI
├── web/                 SvelteKit, Node adapter
├── compose.yaml         four services: web, api, db, dagger-engine
├── OBJECTIVE.md
├── PLAN.md              Stages 0-7 = MVP, 8-14 = post-MVP, 15+ = future
├── ARCHITECTURE.md
└── versions.md          pinned tool / image versions
```

## License

See repo root LICENSE.
