# Pinned Versions

This file is the source of truth for tool and image versions used by Workend v2.
When bumping versions, update both this file and the corresponding Dockerfile / compose.yaml / go.mod.

## Runtime images (compose.yaml)

| Service | Image | Tag | Notes |
|---|---|---|---|
| api | (built from `v2/api/Dockerfile`) | — | Base: `golang:1.25-alpine` (build), `alpine:3.20` (runtime) |
| web | (built from `v2/web/Dockerfile`) | — | Base: `node:22-alpine` |
| db | `postgres` | `16-alpine` | Postgres 16, alpine for size |
| dagger-engine | `registry.dagger.io/engine` | `v0.13.7` | Pin a specific tag, never `latest` |

## Build-time

| Tool | Version | Notes |
|---|---|---|
| Go | 1.25.x | `go.mod` declares 1.25.7 minimum |
| Node | 22.x | SvelteKit dev + build |
| Postgres client libs | 16 | For driver compatibility |
| Dagger Go SDK | v0.13.x | Match engine version |

## When updating

1. Bump the version here first
2. Update `Dockerfile`s, `compose.yaml`, `go.mod`, `web/package.json`
3. Run `docker compose up --build` and confirm `/healthz` returns green
4. Commit the bump as a single atomic commit
