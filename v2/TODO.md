# Workend v2 — Feature TODO

Tracking the brainstorm of 30 candidate features plus follow-ups. Nineteen feature tracks shipped (see [Done](#done) at the bottom), two were dropped, **9 remain**.

Effort key: **S** (≤ 1 day), **M** (2–5 days), **L** (≥ 1 week).

Each entry sketches scope, key data/code touchpoints, and any prerequisites. None of these are committed scope — pull from the list as priorities firm up.

---

## Execution & Runtime

### Task pipelines / chaining — L
Define ordered task sequences (e.g., `lint → test → build`) that short-circuit on failure.
- New tables: `pipelines (id, project_id, name)` and `pipeline_steps (pipeline_id, position, task_id)`.
- New runtime concept: a `pipeline_run` row that holds child run IDs, status (pending/running/succeeded/failed), and stops on first failure.
- Endpoints: `POST /api/projects/:id/pipelines`, `POST /api/pipelines/:id/runs`, `GET /api/pipeline-runs/:id` (with SSE for live status).
- UI: pipeline editor (drag-to-reorder), pipeline run page with stage-by-stage timeline.

### Incremental live log streaming (Stage 23) — L
The deferred plan-stage. Today the SSE stream emits all output in one event at run completion because Dagger's `container.Stdout()` only resolves at exit.
- Investigate `dagger session` wrapper that exposes per-step stdout, or use the engine's GraphQL subscription API directly.
- Touch points: `run/runner.go` `Execute()`, `run/run.go` `Stream()`. Existing SSE plumbing is already correct — only the data source changes.
- Risk: real coupling to Dagger SDK internals; revisit each engine bump.

---

## Observability & Insights

### Resource usage per run — M
Capture CPU time, peak memory, network I/O from the Dagger container and store on the run record.
- Schema: `runs.cpu_ms BIGINT`, `runs.mem_peak_bytes BIGINT`, `runs.net_rx_bytes BIGINT`, `runs.net_tx_bytes BIGINT`.
- Runner: read cgroup stats inside the Dagger container before exit, or query the engine's stats API. Likely needs a small wrapper script in each base image.
- UI: add "Resources" panel to run detail; chart on trends page (toggle between duration / CPU / memory).

### Dependency graph visualization — L
Parse common lockfiles on sync (`package-lock.json`, `go.sum`, `Cargo.lock`, `requirements.txt`) and render a tree.
- New per-source detectors: produce `(name, version, depth)` rows into a new `project_dependencies` table.
- Optional: cross-reference with OSV / GHSA for vulnerability flags (overlaps with image vuln scanning).
- UI: collapsible tree with filter/search, "outdated" badges using best-effort semver compare.

---

## Security & Supply Chain

### Secret scanning — M
Scan cloned repos for leaked credentials on every sync.
- Tool: gitleaks or trufflehog in a Dagger container; runs as part of `project.cloneAsync`.
- Schema: `secret_findings (id, project_id, commit_sha, file, rule, severity, found_at)`.
- Notifications: dispatch via the existing `notify` package on new findings.
- UI: per-project findings panel; admin overview.

### SBOM generation — M
Generate Software Bill of Materials (SPDX/CycloneDX) for built images using syft.
- Pipeline: `syft <image> -o spdx-json` via Dagger; store output as a file alongside the image.
- Endpoint: `GET /api/images/:id/sbom` returns the SBOM.
- UI: download button on image detail.

---

## Developer Experience

### Web Push (background notifications) — S/M
Today's browser notifications only fire while the run page is open. Web Push via a service worker would deliver them when the tab is backgrounded or closed.
- Service worker registers a push subscription; backend stores endpoint + keys per device.
- Use the existing `notify` dispatcher's `OnRunComplete` to send web push payloads via VAPID.

---

## Integration & Ecosystem

### Webhook outbound events — S/M
Emit structured webhook events for state changes (run started/finished, project synced) so users can build their own automations.
- Schema: `outbound_webhook_targets (user_id, url, secret, event_filter)`.
- Dispatcher: extend `notify` package with a generic event emitter; HMAC-sign payloads.
- UI: settings panel to add/remove targets; "test event" button.

### GitHub Actions / GitLab CI trigger — M
Two-way bridge with hosted CI:
1. After a successful local run, optionally call `POST /repos/:owner/:repo/actions/workflows/:id/dispatches` (GitHub) or the equivalent GitLab pipeline trigger.
2. Pull CI status back into the dashboard health card.
- Reuses existing OAuth tokens.
- Per-task config (`tasks.ci_dispatch_workflow`).

### VS Code / JetBrains extension — L
Sidebar extension that shows workspace status and triggers runs without leaving the IDE.
- Separate sub-project (`v2/extensions/vscode/`, `v2/extensions/jetbrains/`).
- Wraps the existing HTTP API; reuse the CLI's session-cookie storage.
- VS Code: TypeScript with the standard `vscode` API. JetBrains: Kotlin plugin.
- Out-of-tree dev cycle and publishing each push the effort up.

---

## Done

| Feature | Where it landed |
|---|---|
| Task timeout enforcement | migration 018, [run.go](v2/api/internal/run/run.go), task UI pill |
| Task input parameters | migration 019, run params modal, run detail "Inputs" row |
| Run duration trends | [trends.go](v2/api/internal/run/trends.go), [/projects/[id]/trends](v2/web/src/routes/projects/[id]/trends/+page.svelte) |
| Dark/Light mode | CSS vars in layout, FOUC-free init in [app.html](v2/web/src/app.html) |
| CLI companion | [v2/cli/](v2/cli/cmd/workend/main.go) |
| Branch listing + switching | [project.go ListBranches/SwitchBranch](v2/api/internal/project/project.go), [/projects/[id]/branches](v2/web/src/routes/projects/[id]/branches/+page.svelte), `LsRemoteBranches` fallback for unauthenticated repos |
| Drag-to-PR (cross-provider) | `Provider.CreatePullRequest` in oauth/{[github](v2/api/internal/oauth/github.go),[gitlab](v2/api/internal/oauth/gitlab.go),[gitea](v2/api/internal/oauth/gitea.go)}.go, drag-and-drop on branches page with PR-detail modal |
| Task pinning / favorites | migration 020, [pin/pin.go](v2/api/internal/pin/pin.go), star button on task rows, "Pinned tasks" section on dashboard |
| Keyboard shortcuts | global handler in [+layout.svelte](v2/web/src/routes/+layout.svelte): `?` help, `t` theme, `g h/d/s` navigation |
| Browser notifications | per-tab Notification API toggle on the run detail page |
| Discord / Teams notifications | new dispatcher kinds in [notify.go](v2/api/internal/notify/notify.go) (Discord embeds, Teams MessageCard); selector entries in [settings/notifications](v2/web/src/routes/settings/notifications/+page.svelte) |
| Workspace activity feed | [workspace.Activity](v2/api/internal/workspace/workspace.go) UNION over audit_log; surfaced on the workspace detail page |
| Project activity heatmap | [run/activity.go](v2/api/internal/run/activity.go) per-day aggregation; SVG calendar grid on the trends page |
| Flaky task detection | [run/flaky.go](v2/api/internal/run/flaky.go) flip-detection via window functions; "Flaky tasks" section on the dashboard |
| Re-run at original commit | `Spec.CommitSHA`/`GitURL` honored by the runner via `client.Git(...).Commit(sha).Tree()`; `Re-run @ <sha>` button on the run detail page |
| Run comments + @mentions | migration 021, [comment/comment.go](v2/api/internal/comment/comment.go); thread under the log on run detail page; mention regex highlights `@names`; per-user inbox at [/mentions](v2/web/src/routes/mentions/+page.svelte) with header badge |
| SSH-key git clone | migration 022, [sshkey/sshkey.go](v2/api/internal/sshkey/sshkey.go) (PEM parsed via `golang.org/x/crypto/ssh`, encrypted at rest); [repo.CloneSSH](v2/api/internal/repo/clone.go) sidecar with `GIT_SSH_COMMAND` for `git@…` URLs; settings UI to add/list/delete keys |
| Run retry with backoff | migration 023, `tasks.retry_max`/`retry_backoff_sec`, `runs.attempt`/`parent_run_id`; runner schedules a queued retry on failure; "↻ N×/Ns" pill on each task; "retry · attempt N" badge on retry runs |
| Approval gates | migration 024, `tasks.requires_approval` + `run_approvals` table + new `pending_approval` status; toggle pill on each task; Approve/Reject buttons on the run detail page; gated runs don't enter the queue until approved |
| Image vulnerability scanning | migration 025, [run/scan.go](v2/api/internal/run/scan.go) launches Trivy in a sidecar container after each successful image build; `scan_status`+`vuln_summary` rendered as colored severity pills with a "top findings" expander on the images page |
| Local issue board *(new)* | migration 026, [board/board.go](v2/api/internal/board/board.go); pulls upstream labels at first setup, lets the user pick which become columns; kanban view with HTML5 drag/drop on `/projects/[id]/board`; drop-to-move PATCHes the issue's labels both locally and upstream; `/issues/[id]` view shows body + threaded comments with refresh-on-load + Close button. Provider methods `ListLabels`/`ListIssues`/`ListIssueComments`/`CreateIssueComment`/`SetIssueLabels`/`CloseIssue` implemented for GitHub, GitLab, and Gitea |

## Dropped

- **Log search** — full-text search across historical run logs.
- **Bitbucket OAuth provider** — fourth provider implementation.
