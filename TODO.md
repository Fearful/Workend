# Workend v2 — Feature TODO

Tracking the brainstorm of 30 candidate features plus follow-ups. Twenty-eight feature tracks shipped (see [Done](#done) at the bottom), two were dropped, **0 remain** (VS Code / JetBrains extension moved to a separate repo).

Effort key: **S** (≤ 1 day), **M** (2–5 days), **L** (≥ 1 week).

Each entry sketches scope, key data/code touchpoints, and any prerequisites. None of these are committed scope — pull from the list as priorities firm up.

---

## Future (out-of-tree)

### VS Code / JetBrains extension — L
Sidebar extension that shows workspace status and triggers runs without leaving the IDE.
- Separate sub-project (`extensions/vscode/`, `extensions/jetbrains/`).
- Wraps the existing HTTP API; reuse the CLI's session-cookie storage.
- VS Code: TypeScript with the standard `vscode` API. JetBrains: Kotlin plugin.
- Out-of-tree dev cycle and publishing each push the effort up.

### Incremental live log streaming (Stage 23) — L *(deferred)*
True line-by-line streaming. Today the SSE plumbing is correct but Dagger's `container.Stdout()` resolves at exit, so output arrives as one blob. Needs a Dagger SDK refactor or `dagger session` wrapper. Current 1-second polling works fine for most runs.

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
| Task pipelines / chaining | [pipeline/pipeline.go](api/internal/pipeline/pipeline.go); `pipelines`, `pipeline_steps`, `pipeline_runs` tables; callback-driven sequential execution with first-failure short-circuit; pipeline editor UI with drag-to-reorder at `/projects/[id]/pipelines`; pipeline run detail at `/pipeline-runs/[id]` |
| Resource usage per run | `runs.cpu_ms`/`mem_peak_bytes`/`net_rx_bytes`/`net_tx_bytes` columns; runner captures cgroup stats from `/proc` inside Dagger container; "Resources" panel on run detail |
| Dependency parsing | [deps/deps.go](api/internal/deps/deps.go); parsers for npm (`package-lock.json`), Go (`go.sum`), Cargo (`Cargo.lock`), PyPI (`requirements.txt`); `project_dependencies` table with ecosystem/name/version/depth; synced on every project sync |
| Secret scanning | [secscan/secscan.go](api/internal/secscan/secscan.go); gitleaks 8.18.4 in a Dagger container; `secret_findings` table with fingerprint-based upsert + auto-resolve; per-project findings endpoint |
| SBOM generation | [run/sbom.go](api/internal/run/sbom.go); syft v1.16.0 outputs SPDX-JSON; stored in `images.sbom_data` (JSONB) with `sbom_format`/`sbom_generated_at`; runs async after image build |
| Web Push notifications | [push/push.go](api/internal/push/push.go); VAPID-based Web Push via `webpush-go`; `push_subscriptions` table; subscribe/unsubscribe/list endpoints; auto-deletes 410 Gone subscriptions; fires from `notify.Dispatcher.OnRunComplete` |
| Outbound webhook events | [events/events.go](api/internal/events/events.go); `outbound_webhook_targets` table; HMAC-SHA256 signed payloads (`X-Workend-Signature-256`); events: `run.created`, `run.started`, `run.finished`, `project.synced`, `image.built`; per-target event filter, test endpoint |
| GitHub/GitLab/Gitea CI trigger | [remoteci/remoteci.go](api/internal/remoteci/remoteci.go); `remote_pipeline_runs` table; auto-sync (5-min staleness check); trigger endpoint dispatches workflows via OAuth provider tokens; CI config detection in [cidetect/cidetect.go](api/internal/cidetect/cidetect.go) |

## Dropped

- **Log search** — full-text search across historical run logs.
- **Bitbucket OAuth provider** — fourth provider implementation.
