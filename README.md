# Workend v2

Self-hosted, browser-based developer workstation.

Pick a git repo, see what's runnable in it, run those tasks reproducibly in isolated containers, watch the output live, schedule them, build images from Dockerfiles, get notified on failures, keep a log of every run.

## Status: v0.7.0

All planned stages are complete (Stage 23 — true incremental log streaming — is deliberately deferred). The system handles multi-user shared workspaces across multiple OAuth providers, with task pipelines, outbound webhooks, Web Push, secret scanning, SBOM generation, dependency parsing, resource usage tracking, CI integration, HMAC-verified webhooks, per-user disk quotas, lazy token refresh, and a scheduler safe to run on multiple replicas.

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
| `WORKEND_VAPID_PUBLIC` + `_PRIVATE` + `_SUBJECT` | Web Push notifications |
| `WORKEND_OIDC_ISSUER` + `_CLIENT_ID` + `_CLIENT_SECRET` | OIDC SSO login |

`WORKEND_TOKEN_KEY` is a base64-encoded 32-byte key for at-rest encryption of stored OAuth tokens. Generate with:

```sh
openssl rand -base64 32
```

## What's in v0.7.0

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
- Repo browser when adding a project (pick from any connected provider, with filter)
- Diff view between two runs of the same task
- Workspace sharing: invite by email, owner / member roles, leave / remove
- Webhook-triggered syncs (per-project token URL)
- Backup / restore scripts (pg_dump + volume tars)
- Pluggable custom detectors via JSON config (Makefile / Cargo / Taskfile out of the box)
- Dagger task source: `dagger.json` projects expose their functions as runnable tasks

**Polish + hardening (Stages 24–28)**
- HMAC webhook signature verification (per-project secret; github / gitea / gitlab header schemes)
- Per-user disk quotas with pre-clone gate; `/api/me/usage` for the dashboard
- Lazy-on-401 token refresh (in addition to lazy-on-expiry from Stage 15)
- Multi-instance-safe scheduler (`SELECT ... FOR UPDATE SKIP LOCKED`)

**Pipelines, security & integrations (post-Stage 28)**
- Task pipelines: ordered task chains that short-circuit on first failure
- Resource usage capture: CPU, memory, network I/O per run via cgroup stats
- Dependency parsing: npm, Go, Cargo, PyPI lockfiles into a searchable table
- Secret scanning: gitleaks integration with fingerprint tracking + auto-resolve
- SBOM generation: syft SPDX-JSON for built images
- Web Push: VAPID-based browser push notifications (background delivery)
- Outbound webhooks: HMAC-SHA256 signed event payloads for external automation
- GitHub/GitLab/Gitea CI: trigger remote workflows + sync pipeline run status
- CI config detection: auto-detect GitHub Actions, GitLab CI, Jenkins, CircleCI configs

## What's not here

- Bitbucket OAuth (the extension point exists — implement the `oauth.Provider` interface)
- Image push to a registry (local builds only; runs record digest + size)
- VS Code / JetBrains extensions (planned as out-of-tree sub-projects)
- True line-by-line live log streaming (see "Honest limitations")

## Honest limitations in v0.7.0

- **Live log streaming is structurally complete but materially batched.** Dagger v0.13's `container.Stdout()` only returns once the container exits. SSE plumbing is correct; expect a single large delivery at completion. Stage 23 — the runner rewrite that fixes this — is deliberately deferred.
- **Self-signed TLS on self-hosted Gitea/GitLab** isn't supported out of the box — Workend's HTTP client doesn't bundle custom CAs. Mount one into the api container if you need it.
- **Repo browser search is page-local.** Filter applies only to the currently loaded page (50 repos at a time).
- **Disk-quota check walks the filesystem on every clone.** Fine for small repo counts; consider caching if you have hundreds of projects.

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
│   │   ├── artifact/    run artifact capture + download
│   │   ├── audit/       best-effort audit-log writer
│   │   ├── auth/        password, sessions, signup/login/logout
│   │   ├── board/       issue board (kanban + upstream sync)
│   │   ├── cidetect/    CI config detection (Actions, GitLab CI, etc.)
│   │   ├── comment/     run comments + @mentions
│   │   ├── compose/     Docker Compose lifecycle
│   │   ├── config/      env-driven config
│   │   ├── dagger/      lazy SDK client wrapper
│   │   ├── dashboard/   per-project health aggregation
│   │   ├── db/          pgx pool + embedded goose migrations
│   │   ├── deps/        lockfile dependency parsing
│   │   ├── detect/      task autodetection (npm, just, dockerfile, custom)
│   │   ├── diff/        run-to-run diff
│   │   ├── events/      outbound webhooks (HMAC-signed)
│   │   ├── health/      /healthz
│   │   ├── image/       built-image listing
│   │   ├── notify/      webhook/slack/email/discord/teams dispatcher
│   │   ├── oauth/       GitHub/GitLab/Gitea OAuth
│   │   ├── oidc/        OIDC SSO
│   │   ├── pipeline/    task pipelines (sequential chaining)
│   │   ├── project/     project CRUD + clone-async pipeline
│   │   ├── push/        Web Push (VAPID)
│   │   ├── quota/       per-user disk quotas
│   │   ├── remoteci/    remote CI trigger + sync
│   │   ├── repo/        Dagger-based git clone
│   │   ├── run/         task execution + SSE + image builds + scan + SBOM
│   │   ├── schedule/    cron schedules + ticker goroutine
│   │   ├── secret/      NaCl secretbox wrapper
│   │   ├── secscan/     secret scanning (gitleaks)
│   │   ├── server/      chi router wiring
│   │   ├── stats/       tokei-driven code stats
│   │   ├── task/        task listing
│   │   └── workspace/   workspace CRUD + sharing
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
