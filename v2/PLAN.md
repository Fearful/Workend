# Workend v2 — Staged Plan

This plan breaks the build into deliberate stages. Each stage is independently demoable, has an explicit "definition of done," and lists what's *out of scope* so it doesn't grow.

The intent: you can pause for days or weeks between stages without losing momentum, because each stage leaves the system in a coherent, working state.

---

## Stage 0 — Archive & Foundations

**Goal:** Cleanly separate v1 from v2 and lock in version-pinned foundations.

**Definition of done**
- v1 tagged as `archive/v0.0.1-2015` and root-level README updated to point at v2
- `v2/` directory contains: `OBJECTIVE.md` (done), `PLAN.md` (this), `ARCHITECTURE.md` (skeleton), `.gitignore`, `LICENSE`
- A `versions.md` or top-of-Compose comment pinning: Go version, Postgres major, Dagger engine tag, Node version (for SvelteKit), Docker Compose spec version
- Empty `compose.yaml` placeholder committed
- `v2/api/`, `v2/web/`, `v2/db/migrations/` directories exist with `.gitkeep`

**Out of scope**
- Any application code
- Any container images

**Key decisions to lock in here**
- Final Go version (recommend latest stable)
- Postgres major (recommend 16; 17 is fine but ecosystem support varies)
- Dagger engine tag (pin a specific version, not `latest`)
- Module path / repo path for the Go API

**Demo**
- `git tag -l` shows the archive tag
- Repo tree clearly shows v1 (read-only history) vs v2 (active dev)

---

## Stage 1 — Walking Skeleton

**Goal:** `docker compose up` brings four services online and they can see each other. Nothing useful happens yet, but the *plumbing* is real.

**Definition of done**
- `compose.yaml` defines four services: `web`, `api`, `db`, `dagger-engine`
- `web` serves a SvelteKit page that says "Workend v2 — alive"
- `api` is a Go binary serving `GET /healthz` returning `{ "status": "ok", "db": "ok", "dagger": "ok" }`
- `/healthz` actually pings Postgres and the Dagger engine; failures show in the JSON
- `db` is Postgres with an empty database, persisted via a named volume
- `dagger-engine` runs the official engine image and exposes its socket to the API via a shared volume
- Web fetches `/healthz` from the API on page load and renders the JSON
- All non-engine containers run with `USER` directive set to non-root and `cap_drop: [ALL]` where possible

**Out of scope**
- Auth, login, signup
- Any data model beyond a `healthz` ping
- Any user-facing UI beyond the status page

**Key decisions to lock in here**
- The exact Dagger engine privilege model (is `privileged: true` required for the engine container in your setup, or does a more constrained capability set work? **This is the highest-risk technical question in the whole project — answering it here, when scope is tiny, is intentional.**)
- API ↔ Dagger socket transport (Unix socket on shared volume vs TCP)
- Web ↔ API origin/CORS setup
- Whether SvelteKit runs in dev mode in Compose or as a built static export served by the Node adapter

**Demo**
- `docker compose up`, open `localhost:3000`, see four green statuses

**Risks**
- Dagger engine + rootless Compose combination may need iteration. If a fully-rootless engine proves too painful, the fallback is `privileged: true` on the engine container only, with the rest of the stack rootless. Document the choice in `ARCHITECTURE.md`.

---

## Stage 2 — Auth & Workspace Data Model

**Goal:** A real human can sign up, log in, and create empty workspaces. Persistence works end-to-end.

**Definition of done**
- Postgres migrations (using `goose` or `golang-migrate`) for: `users`, `sessions`, `workspaces`, `projects` (projects table exists but is empty / unused this stage)
- Go API endpoints: `POST /auth/signup`, `POST /auth/login`, `POST /auth/logout`, `GET /me`, `GET /workspaces`, `POST /workspaces`, `DELETE /workspaces/:id`
- Password hashing with `argon2id` or `bcrypt` (pick one, document the choice)
- Session cookies: HTTP-only, SameSite=Lax, Secure-when-HTTPS, signed
- SvelteKit pages: `/login`, `/signup`, `/` (workspace list), `/workspaces/new`
- SvelteKit hooks: load functions check session via the API, redirect to `/login` if absent
- A user can: sign up, log in, create a workspace, see it listed, delete it, log out

**Out of scope**
- GitHub OAuth (post-MVP)
- Password reset, email verification
- Multi-user permissions (one user = their own workspaces, no sharing)
- Any "project" UI

**Key decisions to lock in here**
- Session storage: Postgres table vs Redis vs JWT. Recommendation: Postgres table — simpler, no extra service, easy to invalidate.
- Argon2id vs bcrypt. Recommendation: argon2id, but bcrypt is fine.
- SvelteKit auth pattern: server-side `hooks.server.ts` reading session cookie, attaching `event.locals.user`. Standard.

**Demo**
- Sign up as `joan`, log in, create workspace "Personal", see it on the dashboard, log out, fail to access dashboard, log back in, delete workspace.

---

## Stage 3 — Repo Ingestion

**Goal:** Add a project to a workspace by git URL. The server clones it into a managed volume. The UI shows it.

**Definition of done**
- `projects` table in use: `id, workspace_id, name, git_url, default_branch, local_path, created_at, last_synced_at`
- Go API: `POST /workspaces/:id/projects`, `GET /workspaces/:id/projects`, `GET /projects/:id`, `DELETE /projects/:id`, `POST /projects/:id/sync` (re-pull)
- Cloning happens inside a Dagger pipeline, *not* via shelling out git on the API host. Clone result is exported to a managed volume mounted at a known path.
- Project detail page shows: name, git URL, default branch, last synced timestamp, latest commit SHA + message + author
- Manual "Sync" button re-pulls
- Cloning supports public HTTPS URLs only this stage

**Out of scope**
- Authenticated git (private repos, SSH) — Stage 9 with GitHub OAuth
- Branch switching
- File browser
- Any task detection or execution
- Auto-sync on push (webhooks)

**Key decisions to lock in here**
- Volume layout: e.g., `repos/<workspace_id>/<project_id>/` — keep it predictable so the Dagger pipelines can find it
- How project deletion handles the volume (soft delete + GC vs hard delete)
- Clone shallowness: full clone vs `--depth 1` (recommend full to support history features later)

**Demo**
- Add `https://github.com/sveltejs/kit.git` as a project, see it appear with last commit info, click sync, see timestamp update.

---

## Stage 4 — Task Autodetection

**Goal:** Open a project page and see what's runnable.

**Definition of done**
- Detector module in the Go API runs as a Dagger pipeline against the project's volume
- Two detectors implemented:
  - `package.json` → list each script in `scripts` as a runnable task
  - `justfile` → parse and list each recipe (use `just --list --justfile <path>` inside a container)
- `tasks` table: `id, project_id, source (npm|just), name, raw_command, detected_at`
- Tasks re-detected on every project sync
- Project page shows tasks grouped by source, with a "Run" button (button is disabled / does nothing this stage)

**Out of scope**
- Actually running tasks (Stage 5)
- Other task sources (Makefile, Taskfile, Cargo, Dagger functions) — add later, design the detector interface to accept new sources without refactor
- Task arguments / env vars

**Key decisions to lock in here**
- Detector interface in Go: `type Detector interface { Detect(ctx, repoDir) ([]Task, error) }`. Simple, easy to add new ones.
- Whether to normalize task names across sources or keep `source` as a discriminator. Recommend keeping the source explicit — running `npm run test` is a different operation from running `just test`.

**Demo**
- Add a project with both a `package.json` (with scripts) and a `justfile`. See both task lists rendered.

---

## Stage 5 — Run a Task (the core loop, batch mode)

**Goal:** Click a task's "Run" button, wait, see the result. No streaming yet — this stage proves the execution path end-to-end.

**Definition of done**
- `runs` table: `id, task_id, project_id, commit_sha, status (queued|running|succeeded|failed|cancelled), started_at, finished_at, exit_code, log_text`
- Go API: `POST /tasks/:id/runs`, `GET /runs/:id`, `GET /projects/:id/runs`
- Run execution: API constructs a Dagger pipeline that:
  1. Mounts the project volume read-write into a container with the right base image (Node for npm, a small image with `just` for justfile)
  2. Executes the raw command
  3. Captures combined stdout+stderr
  4. Returns exit code
- Captured log written to `runs.log_text` on completion (truncated to e.g. 10 MB with a flag if exceeded)
- UI: clicking Run navigates to the run detail page, which polls `/runs/:id` until status is terminal, then displays the full log + exit code
- Concurrent runs of different tasks supported
- Same task can't have two simultaneous runs (lock at the task level)

**Out of scope**
- Live log streaming (Stage 6)
- Cancel mid-run (Stage 6)
- Custom container images per task
- Run inputs (env vars, secrets)
- Parallel runs of the same task

**Key decisions to lock in here**
- Base image per task source: e.g., `node:20-alpine` for npm, a custom small image with `just` for justfile. Pin tags.
- How the API tracks running pipelines (in-memory map keyed by run ID) vs job queue (Postgres `LISTEN`/`NOTIFY` or a real queue). Recommend in-memory for MVP — the API is a single instance.
- Log capture strategy: write incrementally to a file in a volume vs buffer in memory until completion. Recommend file-on-volume from the start; Stage 6 will tail it for SSE.

**Demo**
- Run `npm test` on a small project. Wait. See pass/fail and full log.

**This stage is the MVP's heart. Everything before it is plumbing; everything after is polish or expansion.**

---

## Stage 6 — Live Log Streaming & Cancel

**Goal:** Watch a task run in real time. Cancel it mid-run.

**Definition of done**
- Go API: `GET /runs/:id/log/stream` — SSE endpoint that:
  - Tails the run's log file from byte 0
  - Sends new lines as `data:` events
  - Sends a final `event: done` with exit code when the run terminates
- Go API: `POST /runs/:id/cancel` — sends cancel signal to the running Dagger pipeline; run status becomes `cancelled`
- SvelteKit run detail page subscribes to the SSE stream when status is `running` or `queued`; falls back to the static log when terminal
- Auto-scroll log view, with "pause auto-scroll" affordance when user scrolls up
- Cancel button visible only while running

**Out of scope**
- Resumable streams across reconnects (nice to have; defer)
- Multi-viewer streams (one user, many tabs is fine; multi-user later)
- Log search within a run

**Key decisions to lock in here**
- SSE vs WebSocket. Recommendation: SSE — simpler, one-way is all you need, plays nicely with HTTP/2 and proxies.
- Cancel semantics: graceful (SIGTERM, then wait, then SIGKILL) vs immediate (kill the container). Recommend graceful with a 10s timeout.
- Log encoding: assume UTF-8, replace invalid bytes with U+FFFD on the way to the browser.

**Demo**
- Run `npm install` on a fresh project. Watch lines appear live. Cancel halfway. See status flip to `cancelled`, see partial log preserved.

---

## Stage 7 — Run History & MVP Close

**Goal:** Find and inspect any past run. MVP feature-complete.

**Definition of done**
- Project page shows a runs list: timestamp, task name, commit SHA, duration, status
- Sort by recency (default) and by task
- Filter by status (succeeded / failed / cancelled)
- Run detail view shows full log, metadata, commit info, "re-run" button
- Re-run creates a new run of the same task at the *current* repo HEAD (not the original commit — surprising-but-correct; document this and add an issue for "re-run at original commit" later)
- Top-level dashboard shows recent runs across all projects in current workspace
- README updated with screenshots, install instructions, "this is v0.1.0"
- Tag `v0.1.0` cut

**Out of scope**
- Diff between runs (post-MVP)
- Run aggregations / charts
- Re-run at original commit (open issue, add later)

**Key decisions to lock in here**
- Pagination strategy for runs list (cursor-based recommended; offset is fine to start)
- Retention policy: keep all runs forever for now; revisit when log volume becomes a concern

**Demo**
- A user with several projects can: see what they ran today, find a failed run from last week, re-run it, see the new result.

**MVP is done.** Pause here. Use it. Decide if the design holds up before building Stage 8+.

---

## Stage 8 — Code Statistics

**Goal:** Per-project insight panel: language breakdown, line counts, file counts.

**Definition of done**
- Stats collection runs as a Dagger pipeline using `tokei` (or `scc`) inside a small container
- Triggered on project sync; results stored in a `project_stats` table (one row per project per sync)
- Project page renders a stat panel: top languages by lines, total lines, total files, last computed timestamp
- Historical stats kept (so you can chart trend later)

**Out of scope**
- Charts / trends UI (later)
- Per-directory breakdown
- Dependency analysis

**Demo**
- Open a real project, see "Go: 12k lines, JS: 8k lines, …" updated to today.

---

## Stage 9 — GitHub OAuth & Private Repos

**Goal:** Sign in with GitHub, browse your repos, add one with one click. Private repos work.

**Definition of done**
- GitHub OAuth flow added (in addition to local password)
- User can connect GitHub account post-signup; access token stored encrypted at rest in `users` table
- Repo browser page lists user's repos via GitHub API, with pagination
- "Add as project" button uses the stored token for `git clone` (HTTPS with token)
- Token refresh handled

**Out of scope**
- GitLab, Bitbucket, Gitea (extend the auth provider interface to allow them later)
- Webhooks for push-triggered syncs (Stage 11ish)

**Key decisions**
- Token encryption key: store in env, document the rotation procedure
- Scope set: minimum needed (`repo`, `read:user`)

---

## Stage 10 — Multi-User Polish

**Goal:** Two-plus humans can use the same instance without stepping on each other.

**Definition of done**
- Workspaces strictly per-user (already true since Stage 2); add explicit row-level checks on every API endpoint
- Per-user disk quotas on repo storage (configurable, default e.g. 5 GB)
- Per-user concurrent-run limits (configurable, default e.g. 3)
- Admin user role: can see all users, suspend accounts, see system-wide stats
- Audit log table: who did what when (login, project add, run start)

**Out of scope**
- Workspace sharing between users (deliberately deferred — adds permission model complexity)
- Teams / orgs

---

## Stage 11 — Scheduled Runs

**Goal:** Cron a task. Get notified on failure.

**Definition of done**
- `schedules` table: `id, task_id, cron_expr, enabled, last_run_at, next_run_at`
- Schedule editor UI per task
- A scheduler goroutine in the API ticks every minute, finds due schedules, enqueues runs
- Notification table + worker: on run failure, fire configured notifications (webhook URL only this stage; email later)
- UI: "Schedules" tab per project, "Notifications" settings per user

**Out of scope**
- Distributed scheduling (single API instance is fine)
- Complex schedule rules (every-other-Tuesday-unless-holiday — out)
- Email/Slack/etc. (Stage 14)

---

## Stage 12 — Cross-Project Dashboard

**Goal:** Open Workend, immediately see the health of everything you care about.

**Definition of done**
- Dashboard view shows a card per project: last sync, last run status per task type, days since last commit, language summary
- Visual indicators: green/yellow/red for last-run-status, freshness indicators
- Click-through to project detail
- Sortable / filterable

**Out of scope**
- Custom dashboard layouts
- Alerting from the dashboard view (use Stage 11 notifications)

---

## Stage 13 — Image Builds

**Goal:** Build a project's Dockerfile, see the resulting image, optionally push.

**Definition of done**
- Detector for `Dockerfile` adds a `build-image` task type
- Run executes `dag.Container().Build(repoDir)` (Dagger's BuildKit) producing an image
- Image can be tagged and stored in a local registry container (add to Compose) or pushed to a configured external registry
- Run history shows image digest and size
- Image list per project: digest, size, built-from commit, timestamp

**Out of scope**
- Vulnerability scanning (later, with `trivy` in a container)
- Multi-arch builds (BuildKit supports it; expose later)

---

## Stage 14 — Notifications: Beyond Webhooks

**Goal:** Email and Slack on run failure.

**Definition of done**
- SMTP-based email notifications (configured per-user)
- Slack webhook integration
- Notification rules: per project, per task, on failure / on status change

**Out of scope**
- Discord, Teams, etc. (extension point; add as needed)
- Notification batching / digesting

---

## Stage 15 — GitLab + Gitea OAuth (multi-provider)

**Goal:** Sign in with GitLab or Gitea (in addition to GitHub), connect any combination of the three, and clone private repos from any connected provider — including self-hosted instances.

**Why now:** Stage 9 baked GitHub-specific assumptions into 6+ places: the `oauth.GitHub` struct's hardcoded URLs, `project.lookupAuth`'s `strings.Contains(url, "github.com")` check, the web's `/auth/github/*` proxy routes, the env-var naming, the Settings UI, the user_tokens schema's missing refresh-token columns, and the lack of an `instance_url` for self-hosted servers. Adding even one more provider without lifting these forces copy-paste. With three providers in scope, the abstraction pays for itself immediately.

### Current GitHub-specific code that needs lifting

| Location | What's hardcoded |
|---|---|
| `internal/oauth/github.go` | URL constants, `Bearer` auth header, `/user` → `login` field shape |
| `internal/oauth/handlers.go` | Handler methods on `*GitHub` directly, not on a Provider interface |
| `internal/project/project.go:lookupAuth` | String-matches `"github.com"` to decide token use |
| `internal/repo/clone.go:injectAuth` | Single `x-access-token:` injection scheme |
| `internal/server/server.go` | `s.github` field, `if s.github != nil` block per route |
| `internal/config/config.go` | GitHub-only env vars |
| `internal/db/migrations/008_user_tokens.sql` | No `refresh_token`, no `expires_at`, no `instance_url` |
| `web/src/routes/settings/+page.svelte` | One hardcoded "GitHub" panel |
| `web/src/routes/auth/github/start/+server.ts` | One hardcoded provider in the URL |
| `web/src/routes/auth/github/callback/+server.ts` | Same |

### Provider differences that drive the abstraction

| | **GitHub** | **GitLab** | **Gitea** |
|---|---|---|---|
| Instance URL | fixed `github.com` | configurable (defaults `gitlab.com`) | configurable (no default) |
| Auth header | `Authorization: Bearer <token>` | `Authorization: Bearer <token>` | `Authorization: token <token>` |
| User endpoint | `/user` → `.login` | `/api/v4/user` → `.username` | `/api/v1/user` → `.login` |
| OAuth scopes (read user + repos) | `repo read:user` | `read_user read_repository write_repository api` | `repo read:user` |
| Refresh tokens | not by default | yes | yes |
| Clone URL token injection | `https://x-access-token:<tok>@host/...` | `https://oauth2:<tok>@host/...` | `https://<tok>:@host/...` |
| Repo-list endpoint | `/user/repos?per_page=100` | `/api/v4/projects?membership=true&simple=true` | `/api/v1/user/repos` |

### Definition of done

- New `internal/oauth.Provider` interface:
  ```go
  type Provider interface {
      Kind() string                     // "github" | "gitlab" | "gitea"
      InstanceURL() string              // e.g. "https://gitlab.example.com"
      AuthorizeURL(state string) string
      ExchangeCode(ctx, code) (Token, error)
      RefreshToken(ctx, refresh string) (Token, error)
      FetchHandle(ctx, accessToken string) (string, error)
      InjectCloneAuth(rawURL, accessToken string) string
      OwnsHost(host string) bool        // for clone routing
  }
  type Token struct { Access, Refresh string; ExpiresAt *time.Time; Scopes string }
  ```
- New `internal/oauth.Registry` keyed by provider Kind, populated from config at startup. `s.providers` replaces `s.github` on the `Server`.
- Migration **013_user_tokens_v2**: add `refresh_token BYTEA`, `expires_at TIMESTAMPTZ`, `instance_url TEXT NOT NULL DEFAULT ''`. Existing GitHub rows stay valid (instance_url defaults to `https://github.com`).
- Three concrete provider implementations (`github.go`, `gitlab.go`, `gitea.go`) — each ~150 lines, mostly URL/header/field-name differences.
- Per-provider OAuth handlers registered as `/api/auth/{provider}/start`, `/api/auth/{provider}/callback`, plus `/api/me/connections` (list all), `/api/me/connections/{provider}` (status/disconnect).
- `project.lookupAuth` is rewritten: parse the git URL's host, ask each connected provider whether it owns that host (`OwnsHost`), use the first match's token. Multi-provider users with self-hosted instances now route correctly.
- `repo.Clone` calls `provider.InjectCloneAuth(url, token)` — the URL injection scheme stops being a single hardcoded function.
- A token-refresh middleware: when a provider's API call returns 401 *and* a refresh_token exists *and* `time.Now() > expires_at`, the refresh path runs once before retry. Failed refresh marks the row disconnected.
- Web Settings page becomes a "Connections" panel that lists all configured providers; each shows connect/disconnect/handle. Driven by `GET /api/me/connections`.
- Web `/auth/{provider}/start` and `/auth/{provider}/callback` proxy routes are parametrized — single dynamic route per direction, not three copies.
- Compose `compose.yaml` adds env passthroughs for GitLab and Gitea: `WORKEND_GITLAB_CLIENT_ID/_SECRET/_URL`, `WORKEND_GITEA_CLIENT_ID/_SECRET/_URL`. All optional.
- README and ARCHITECTURE updated.

### Configuration story

| Provider | Env vars | Notes |
|---|---|---|
| GitHub | `WORKEND_GITHUB_CLIENT_ID`, `WORKEND_GITHUB_CLIENT_SECRET` | Instance fixed to github.com |
| GitLab | `WORKEND_GITLAB_CLIENT_ID`, `WORKEND_GITLAB_CLIENT_SECRET`, `WORKEND_GITLAB_URL` (default `https://gitlab.com`) | Self-hosted by setting URL |
| Gitea | `WORKEND_GITEA_CLIENT_ID`, `WORKEND_GITEA_CLIENT_SECRET`, `WORKEND_GITEA_URL` (no default; required if client_id set) | Always self-hosted |
| Shared | `WORKEND_TOKEN_KEY` (existing) | Same encryption key used for all providers' tokens |

A provider is enabled only if its `_CLIENT_ID` and `_CLIENT_SECRET` are set. Workend boots fine with zero, one, two, or three.

### Out of scope

- Multiple instances of the same provider type (e.g., two different self-hosted GitLab servers). One instance per provider in v0.3 — revisit if real use demands it.
- Repo browser UI (pick from a list when adding a project). Stage 16.
- Webhook-driven sync from any provider. Stage 18 territory.
- BitBucket. Easy to add later by implementing the `Provider` interface; not requested.
- SSH-key-based clone. HTTPS+token only.
- Automatic permission discovery (knowing whether the user can write to a given repo).

### Key decisions (locked in)

1. **Multiple instances per provider kind.** `user_tokens` renamed to `provider_connections` with surrogate `id` PK and `UNIQUE (user_id, provider, instance_url)`.
2. **Lazy token refresh.** Refresh happens at use-time when `expires_at < now()` and a refresh_token is stored, inside `Registry.accessTokenForProvider`. No background ticker.
3. **Host-equality matching** for clone-URL routing. `OwnsURL` strips ports and lowercases.
4. **Existing rows preserved** in migration via `ALTER TABLE ... ADD COLUMN instance_url TEXT NOT NULL DEFAULT 'https://github.com'`.

### Demo

- Configure all three providers in compose env (real credentials for github.com, gitlab.com, your-gitea-instance.local)
- Sign in to Workend, go to Settings → Connections
- Connect all three
- Add a private project from each provider — clone succeeds, latest commit shows
- Disconnect one — that provider's repos can no longer be cloned (error on next sync)
- Reconnect — works again

### Risks

- **GitLab/Gitea response shapes drift between versions.** Pin tested versions in `versions.md`. The provider methods isolate parsing so a single-file change covers any future API shift.
- **Self-hosted Gitea behind self-signed TLS.** Workend's HTTP client doesn't add custom CA bundles. Document that Gitea instances need a publicly-trusted cert OR users mount a custom CA bundle into the api container.
- **Refresh-token race on concurrent API calls.** A naive lazy-refresh has a TOCTOU window. Mitigation: per-(user,provider) mutex around refresh, plus check `expires_at > now()` after acquiring the lock.

---

## Stage 16 — Repo browser & token refresh (DONE)

**Shipped:**
- `Provider.ListRepos(ctx, accessToken, page, perPage) ([]Repo, error)` added to the interface; per-provider implementations map github/gitlab/gitea response shapes onto a normalized `Repo` struct (name, full_name, description, private, html_url, clone_url, default_branch, updated_at)
- `Registry.ListReposForUser` reuses `accessTokenForProvider` so lazy-refresh applies here too
- `GET /api/me/connections/{provider}/repos?page=N` endpoint
- Add-project page split into a two-column layout: form on the left, picker on the right with provider tabs and a clickable repo list. Click fills name + clone_url + default_branch in the form.

**Still deferred:**
- Filtering/searching across repos (visible-pages-only for now)
- Lazy-on-401-response refresh (current path is lazy-on-expiry)

---

## Stages Beyond

Captured here as headers only — fill in when relevant:

- **Stage 17** — Diff view between runs
- **Stage 18** — Workspace sharing between users
- **Stage 19** — Webhook-triggered syncs (push to repo → auto-sync → optional auto-run)
- **Stage 20** — Backup/restore of the Workend instance
- **Stage 21** — Plugin system for custom detectors / runners
- **Stage 22** — Task templates / shared Dagger modules from Daggerverse

---

## How to Use This Plan

- Each stage is its own checkpoint. Tag a release at the end of each stage that lands. (`v0.1.0` after Stage 7, `v0.2.0` after Stage 8, etc.)
- Don't pull future-stage work into a current stage. If something looks tempting, add it to the next stage's section instead.
- Out-of-scope lists are load-bearing. They prevent the stage from sprawling.
- Stages 0–7 are MVP. Stages 8+ are independent and can be reordered based on what's actually painful in real use.
- Revisit this plan after each stage; update freely.
