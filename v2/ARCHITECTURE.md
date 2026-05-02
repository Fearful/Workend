# Workend v2 — Architecture

This document describes the technical architecture. For *what* Workend does, see [OBJECTIVE.md](OBJECTIVE.md). For staged build plan, see [PLAN.md](PLAN.md).

> **Status:** Skeleton. Filled in incrementally as stages land. Each stage adds a section here.

---

## High-Level Topology

```
   Browser
      │
      │ HTTP / SSE
      ▼
┌──────────────┐         ┌──────────────┐
│     web      │  HTTP   │     api      │
│  (SvelteKit) │ ──────▶ │     (Go)     │
└──────────────┘         └──────┬───────┘
                                │
                ┌───────────────┼────────────────┐
                │               │                │
                ▼               ▼                ▼
         ┌─────────────┐  ┌──────────┐   ┌──────────────┐
         │  postgres   │  │  dagger  │   │ repos volume │
         │   (data)    │  │  engine  │   │ (git clones) │
         └─────────────┘  └──────────┘   └──────────────┘
                                │
                                ▼
                         ┌──────────────┐
                         │  ephemeral   │
                         │  task pods   │
                         │ (per run)    │
                         └──────────────┘
```

All services run in Docker Compose on a single host. Communication paths:

- **web → api**: HTTP (REST) and SSE (live logs). Same-origin via Compose network.
- **api → postgres**: TCP, Compose network.
- **api → dagger-engine**: Unix socket on a shared volume. The engine spawns ephemeral containers for every task.
- **api ↔ repos volume**: read-only metadata access (e.g., listing manifest files); never executes code on the host.
- **dagger-engine ↔ repos volume**: read-write; mounts repo dirs into task containers.

---

## Components

### web (SvelteKit, Node adapter)
- Renders pages, server-side data loading via `+page.server.ts`
- Calls the `api` service over HTTP
- Holds session cookie issued by the `api`
- Streams live job logs via SSE
- Runs as non-root user

### api (Go)
- HTTP server (router TBD: `chi` recommended)
- Owns: auth, sessions, workspace/project/task/run CRUD
- Constructs Dagger pipelines for every operation that touches user code (clone, detect, run, build)
- Streams live logs from Dagger to SSE
- Runs as non-root user; mounts the Dagger socket as the only privileged channel

### db (Postgres 16)
- Schema migrations via `goose` (chosen for simplicity; revisit if multi-direction needed)
- Tables added per stage; current state in `v2/db/migrations/`
- Persisted via named volume

### dagger-engine
- Official `registry.dagger.io/engine:vX.Y.Z` image
- Exposes Unix socket on shared volume to `api`
- Privilege model: TBD — see [Stage 1 decision in PLAN.md](PLAN.md)

---

## Data Flow Examples

### Adding a project
1. User submits git URL via web
2. web → POST `/api/projects` → api
3. api inserts row in `projects` table (status: cloning)
4. api invokes Dagger pipeline: clone repo → export to repos volume at `/repos/<workspace_id>/<project_id>`
5. api updates `projects` row (status: ready, default_branch, latest_commit)
6. web refetches and displays

### Running a task
1. User clicks "Run" on a task
2. web → POST `/api/tasks/:id/runs` → api
3. api inserts row in `runs` table (status: queued)
4. api launches goroutine that builds Dagger pipeline:
   - Container with appropriate base image (e.g., `node:22-alpine` for npm)
   - Mounts `/repos/<workspace_id>/<project_id>`
   - Sets working directory to repo root
   - Executes the task command
   - Captures combined stdout+stderr to a log file on the cache volume
5. api updates `runs` row to `running`, then `succeeded`/`failed` on completion
6. web subscribes to SSE for live log; falls back to static log when status is terminal

---

## Security Model

- API never invokes `child_process` against user code on the host
- All user code runs inside Dagger-managed containers
- Repo volumes are mounted into task containers; never executable from the API process
- Session cookies: HTTP-only, SameSite=Lax, signed
- Passwords: argon2id (Stage 2)
- No external network access for task containers by default (TBD: configurable per task — likely Stage 11 territory)

---

## Open Architectural Questions

These get answered during the named stage:

- **Stage 1**: Dagger engine privilege model (`privileged: true` vs constrained capabilities)
- **Stage 1**: Dagger socket transport (Unix domain vs TCP)
- **Stage 5**: In-memory run tracking vs persistent queue
- **Stage 6**: SSE vs WebSocket (current lean: SSE)
- **Stage 9**: Token encryption-at-rest scheme for GitHub OAuth tokens
- **Stage 11**: Distributed scheduling (probably not needed; single API instance)

---

## Per-Stage Architecture Notes

### Stage 0
Foundations only. No application architecture yet — directories scaffolded.

### Stage 1
*(filled in when stage completes)*

*(further stages added here as they land)*
