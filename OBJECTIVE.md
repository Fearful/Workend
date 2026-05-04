# Workend v2 — Objective

A self-hosted, browser-based control surface for working across many code repositories: discover what each project can do, run those tasks reproducibly in isolated containers, and watch them happen live — without ever shelling into the host.

---

## Background

Workend v1 (2015) was a Node.js + AngularJS prototype that aimed to "bridge developers and the command prompt." It worked by parsing local `gulpfile.js` and `Gruntfile.js` files via regex and shelling out to run tasks directly on the server's filesystem. The concept was sound; the execution model wasn't safe to deploy beyond a single trusted machine, and the stack (AngularJS, Bower, Jade, Mongoose 4, Gulp 3) is now end-of-life.

Workend v2 keeps the original concept and rebuilds the foundation: containerized job execution, a modern frontend, and a security model that makes it deployable as a real multi-user service.

---

## The Problem

Developers in 2026 still spend an outsized share of their time on workflows that are technically simple but operationally noisy:

- **"How do I run this?"** — every repo has its own conventions: `npm run dev`, `just test`, `make build`, `cargo run`, `dagger call`, a script in `bin/`, a paragraph in the README. Onboarding to an unfamiliar repo means reading prose to find runnable commands.
- **"Works on my machine."** — local toolchains drift from CI. A test that passes locally fails in CI (or vice versa) because Node versions, system libraries, or env vars don't match.
- **No shared view across repos.** — when you maintain five projects, there's no single place that shows which are green, which are stale, which have outstanding tasks, which are due for a dependency bump.
- **Scheduled and ad-hoc work bleed into terminals.** — "run the data export every Monday," "rebuild this image after each merge," "kick off the load test before standup" — these end up as ad-hoc cron jobs, forgotten scripts, or manual toil.
- **Job output is ephemeral.** — local `npm test` output scrolls past. CI output is locked behind a vendor's UI. There's no personal log of "what did I run, when, against what code, with what result."

Workend v2 addresses these by treating *running things in repos* as a first-class object — discoverable, reproducible, observable, and shareable.

---

## What Workend Is

Workend is a personal or small-team **developer workstation**, accessed in a browser and self-hosted via Docker Compose. It manages a set of **workspaces**, each containing one or more **projects** (git repositories). For every project it can:

- Auto-detect what's runnable (package.json scripts, justfiles, Makefiles, Taskfiles, Cargo commands, Dagger functions).
- Execute any of those tasks in an ephemeral container — never on the host — with logs streamed live to the browser.
- Persist a history of every run: which task, on which commit, with what exit status, how long it took, full log captured.
- Show per-project insights: language breakdown, line counts, recent activity, dependency health.
- Schedule recurring runs and notify on regressions.

The execution layer is built on **Dagger**, which provides a typed, language-agnostic container orchestration API. Every task — whether it's `npm test`, a custom shell script, or a Docker image build — is expressed as a Dagger pipeline. This buys reproducibility and caching for free, and means Workend never executes user code outside a container.

---

## Core Capabilities

**Workspace and project management**
- Create workspaces, add projects by git URL or local path
- Per-project metadata: description, primary language, status, owner
- Cross-project dashboard view

**Project introspection**
- Auto-detect runnable tasks from common manifests
- Code statistics (file types, line counts) via `tokei` or equivalent, run inside a container
- Detect package managers, language versions, framework hints

**Task execution**
- One-click run of any detected task
- Live stdout/stderr stream in the browser via SSE
- Cancel mid-run, view exit status, re-run against the same or different commit
- Container image used for each task is explicit and pinned

**Run history**
- Every run recorded with: task, commit SHA, started/finished timestamps, duration, exit code, full log
- Filter and search across runs
- Compare consecutive runs of the same task

**Image builds**
- Build container images from a project's Dockerfile via Dagger's BuildKit
- Tag, retain, and optionally push to a registry

**Scheduling** *(post-MVP)*
- Cron-style recurring runs of any task
- Notification hooks (webhook, email) on failure or status change

**Authentication**
- Local username/password, with GitHub OAuth as a second option
- Per-user workspaces; multi-user from day one (even if the day-one user count is one)

---

## Operating Principles

**Isolation by default.** The API process never executes user code on its own filesystem. Every task — including code stat collection — runs in a container managed by the Dagger engine. Compromising a project's `package.json` script cannot escalate to the API host.

**Rootless where it counts.** API, frontend, and Postgres run as non-root users via the `USER` directive in their containers, with capabilities dropped. The Dagger engine container is the one component that requires elevated privileges (BuildKit's nature); it is isolated and accessed only via its socket. No user repository is ever bind-mounted into the API container.

**Reproducible by construction.** Tasks run in declared container images, not against whatever happens to be installed on the host. A run from six months ago can be re-executed against the same image and the same commit and produce identical output.

**Observable by default.** Every run is recorded. Logs are first-class, persistent objects, not ephemeral terminal output. The user is never in a position of "I ran something, what did it say again?"

**Local-first, deploy-anywhere.** Designed to run on a developer's laptop via `docker compose up`, but the same Compose file (with adjusted volumes/secrets) can run on a small VPS for a multi-user instance. No cloud dependency.

**Honest about state.** No optimistic UI. If a job is queued, the UI says queued. If the engine is unreachable, the UI says so. If a repo can't be fetched, the error is shown verbatim.

---

## Non-Goals

Workend deliberately does not try to be:

- **A code editor.** VS Code Web, Gitpod, and Coder serve that need. Workend assumes the user edits code elsewhere (locally, in a cloud IDE, etc.) and pushes/pulls via git.
- **A CI/CD platform.** GitHub Actions, GitLab CI, Buildkite, etc. exist. Workend can complement them — running the same Dagger pipelines locally that CI runs remotely — but it is not trying to be the system of record for production deployments.
- **A project management tool.** v1 had Scrum roles, sprints, and ticket-like fields. v2 drops all of that. If users want issue tracking, they use GitHub/GitLab/Linear.
- **A general container orchestrator.** No Kubernetes ambitions. Workend orchestrates *jobs*, not long-running services.
- **A SaaS.** Workend is self-hosted. There is no hosted version, no telemetry, no account on someone else's server.

---

## Stack at a Glance

| Layer | Choice |
|---|---|
| Frontend | SvelteKit (Node adapter) |
| API | Go (chi or echo router, sqlc for DB) |
| Database | Postgres 16 |
| Job execution | Dagger engine + Dagger Go SDK |
| Image builds | Dagger / BuildKit (no Kaniko) |
| Realtime | Server-Sent Events |
| Orchestration | Docker Compose, containers run rootless via `USER` directive |

Architectural detail lives in a separate `ARCHITECTURE.md` (forthcoming).

---

## MVP Scope (v0.1.0)

The smallest version that demonstrates the core loop:

1. Single-user auth (local password)
2. Create a workspace, add a project by git URL
3. Server clones the repo into a managed volume
4. Auto-detect tasks from `package.json` scripts and `justfile`
5. Run any detected task in a Dagger pipeline; stream logs to the browser
6. Persist run history (task, commit, started/finished, exit code, log)
7. Browse run history per project

Out of MVP: code stats, GitHub OAuth, scheduling, image builds, notifications, multi-user, dashboards.

---

## Beyond MVP

Roughly in priority order:

- Code statistics (containerized `tokei`)
- GitHub OAuth + repo browser for adding projects
- Multi-user with per-user workspaces
- Scheduled runs (cron)
- Cross-project dashboard (status, last run, freshness)
- Image build tasks (Dockerfile detection + Dagger BuildKit)
- Notifications (webhook, email) on run failure
- Task templates / shareable Dagger modules
- Diff view between consecutive runs of the same task

---

## Success Criteria

Workend v2 is successful when its primary user (the author) replaces at least three habitual terminal workflows with Workend equivalents and prefers the result. Concretely: routine test runs, ad-hoc data scripts, and "build the image and check the size" loops should feel faster and more legible in Workend than in a terminal — or the design has missed something and needs to be revisited.
