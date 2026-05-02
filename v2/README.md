# Workend v2

Self-hosted, browser-based developer workstation.

Pick a git repo, see what's runnable in it, run those tasks in isolated containers, watch the output live, keep a log of every run.

## Status: v0.1.0 MVP

The core loop works: sign up → workspace → add a project by git URL → see auto-detected tasks → click run → watch logs → re-run.

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
- Once the clone finishes, runnable tasks appear under the project (npm scripts, justfile recipes)
- Click **Run** to execute one in a Dagger-managed container
- Watch live logs, cancel mid-run, see history, re-run

To stop: `docker compose down`. Add `--volumes` to wipe the database, repo cache, and logs.

## What's in v0.1.0

- Local password auth (argon2id), session cookies
- Workspaces and projects (1 user → many workspaces → many projects)
- Repo ingestion via Dagger pipeline (public HTTPS git only)
- Task autodetection: `package.json` scripts + `justfile` recipes
- Task execution in pinned container images (`node:22-alpine` for npm, `alpine:3.20` + just for justfile recipes)
- Server-Sent Events for live log tail; cancel button while running
- Run history with status filter; re-run from any past run

## What's not in v0.1.0

Deliberately deferred to later stages — see PLAN.md.

- GitHub OAuth + private repos (Stage 9)
- Code statistics (Stage 8)
- Scheduled runs (Stage 11)
- Cross-project dashboard with task health indicators (Stage 12)
- Image builds via Dagger BuildKit (Stage 13)
- Notifications (Stage 14)
- Sharing workspaces between users (Stage 17)

## Honest limitations in v0.1.0

- **Live log streaming is structurally complete but materially batched.** The SSE endpoint + browser tail-follow work correctly, but Dagger v0.13's `container.Stdout()` only returns once the container exits, so the log file is written once at the end. Browser sees the full log dump on completion rather than line-by-line during execution. Reworking the runner for true incremental output is a planned post-MVP refactor.
- **Re-run uses current repo HEAD**, not the original run's commit. Surprising-but-defensible default; "re-run at original commit" is an open issue.
- **Single user per Workend instance is fine; multi-user has no real isolation polish.** Per-user disk quotas, concurrent-run limits, and admin tooling come in Stage 10.
- **No scheduling, no notifications, no dashboards.** This MVP is a manual-only tool.

## Stack

- **Frontend:** SvelteKit 2 + Svelte 5, Node adapter
- **API:** Go 1.24 + chi router + pgx/v5 + goose migrations
- **Database:** PostgreSQL 16
- **Job runner:** Dagger engine v0.13.7 + Dagger Go SDK
- **Image builds:** Dagger / BuildKit (no Kaniko — see project decisions)
- **Realtime:** Server-Sent Events
- **Orchestration:** Docker Compose

All non-engine containers run as non-root (uid 10001) with all capabilities dropped. The Dagger engine is the only privileged container — required by BuildKit, isolated to its own service.

## Layout

```
v2/
├── api/          Go HTTP server, Dagger pipelines, migrations (embedded)
├── web/          SvelteKit app, Node adapter
├── compose.yaml  Four services: web, api, db, dagger-engine
├── OBJECTIVE.md
├── PLAN.md       Staged plan, Stages 0-7 = MVP, Stages 8+ = post-MVP
├── ARCHITECTURE.md
└── versions.md   Pinned tool / image versions
```

## License

See repo root LICENSE.
