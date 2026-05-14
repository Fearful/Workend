package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/artifact"
	"workend/api/internal/audit"
	"workend/api/internal/auth"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/diff"
	"workend/api/internal/events"
	"workend/api/internal/notify"
	"workend/api/internal/oauth"
	"workend/api/internal/push"
	"workend/api/internal/repo"
)

// MaxConcurrentRunsPerUser caps active (queued|running) runs per user.
// Configurable later; hard-coded for Stage 10.
const MaxConcurrentRunsPerUser = 3

const (
	StatusQueued          = "queued"
	StatusRunning         = "running"
	StatusSucceeded       = "succeeded"
	StatusFailed          = "failed"
	StatusCancelled       = "cancelled"
	StatusPendingApproval = "pending_approval"
)

type Run struct {
	ID           uuid.UUID  `json:"id"`
	TaskID       uuid.UUID  `json:"task_id"`
	ProjectID    uuid.UUID  `json:"project_id"`
	CommitSHA    *string    `json:"commit_sha"`
	Branch       *string    `json:"branch"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	ExitCode     *int       `json:"exit_code"`
	TimedOut     bool       `json:"timed_out"`
	LogPath      string     `json:"log_path"`
	Params       RunParams  `json:"params"`
	Attempt      int        `json:"attempt"`
	ParentRunID  *uuid.UUID `json:"parent_run_id"`
	CPUMs        *int64     `json:"cpu_ms"`
	MemPeakBytes *int64     `json:"mem_peak_bytes"`
	NetRxBytes   *int64     `json:"net_rx_bytes"`
	NetTxBytes   *int64     `json:"net_tx_bytes"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Hydrated on Get/List for UI convenience.
	TaskName   string `json:"task_name,omitempty"`
	TaskSource string `json:"task_source,omitempty"`
}

// RunParams mirrors the JSONB column. Empty fields render as null/[].
type RunParams struct {
	Env  map[string]string `json:"env"`
	Args []string          `json:"args"`
}

type Handlers struct {
	Pool           *pgxpool.Pool
	Dagger         *wdagger.Client
	LogsRoot       string
	ArtifactsRoot  string
	Logger         *slog.Logger
	Audit          *audit.Logger
	Notify         *notify.Dispatcher
	Events         *events.Dispatcher
	Push           *push.Sender
	WebURL         string
	DefaultTimeout time.Duration   // applied when a task has no per-task override
	OAuth          *oauth.Registry // nil-safe; used to inject auth into pinned-commit fetches

	mu         sync.Mutex
	running    map[uuid.UUID]uuid.UUID          // task_id -> run_id
	cancellers map[uuid.UUID]context.CancelFunc // run_id -> cancel func
}

func (h *Handlers) Init() {
	h.running = map[uuid.UUID]uuid.UUID{}
	h.cancellers = map[uuid.UUID]context.CancelFunc{}
}

// Create starts a new run for the given task. Optional JSON body:
//
//	{ "env": { "KEY": "VALUE" }, "args": ["--flag", "value"] }
//
// POST /api/tasks/:id/runs
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	// Body is optional. Parse if present and the content-type looks JSON-ish.
	var body struct {
		Env      map[string]string `json:"env"`
		Args     []string          `json:"args"`
		AtCommit string            `json:"at_commit"` // pin to a specific SHA
		Branch   string            `json:"branch"`    // ephemeral run on this branch
	}
	if r.ContentLength > 0 && r.ContentLength < 64*1024 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	if err := validateRunInputs(body.Env, body.Args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body.AtCommit = strings.TrimSpace(body.AtCommit)
	body.Branch = strings.TrimSpace(body.Branch)
	if body.AtCommit != "" && !looksLikeSHA(body.AtCommit) {
		http.Error(w, "at_commit must be a hex commit SHA", http.StatusBadRequest)
		return
	}
	if body.Branch != "" && body.AtCommit != "" {
		http.Error(w, "specify branch OR at_commit, not both", http.StatusBadRequest)
		return
	}
	if body.Branch != "" && len(body.Branch) > 200 {
		http.Error(w, "branch too long", http.StatusBadRequest)
		return
	}

	var (
		spec             Spec
		projectID        uuid.UUID
		commitSHA        *string
		projStatus       string
		taskTOSec        *int
		gitURL           string
		requiresApproval bool
		maxConcurrency   int
		supersedePolicy  string
		needsServices    []string
		artifactPatterns []string
	)
	err = h.Pool.QueryRow(r.Context(), `
		SELECT t.source, t.name, t.raw_command, t.timeout_seconds, t.requires_approval,
		       t.max_concurrency, t.supersede_policy, t.needs_services, t.artifact_patterns,
		       p.id, p.local_path, p.last_commit_sha, p.status, p.git_url
		FROM tasks t
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE t.id = $1 AND m.user_id = $2
	`, taskID, uid).Scan(&spec.Source, &spec.Name, &spec.RawCommand, &taskTOSec, &requiresApproval,
		&maxConcurrency, &supersedePolicy, &needsServices, &artifactPatterns,
		&projectID, &spec.RepoPath, &commitSHA, &projStatus, &gitURL)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if projStatus != "ready" || spec.RepoPath == "" {
		http.Error(w, "project not ready", http.StatusConflict)
		return
	}
	// Tasks that declare `needs_services` require the project's compose
	// stack to be running. We don't auto-start it here (compose up can take
	// minutes); we prompt the user to start it on the overview page first.
	if len(needsServices) > 0 {
		var composeStatus string
		_ = h.Pool.QueryRow(r.Context(),
			`SELECT status FROM compose_instances WHERE project_id = $1`, projectID,
		).Scan(&composeStatus)
		if composeStatus != "running" {
			http.Error(w,
				"task needs services ("+strings.Join(needsServices, ", ")+"); start the compose stack on the project overview first",
				http.StatusConflict)
			return
		}
	}
	spec.TimeoutSec = h.resolveTimeout(taskTOSec)
	spec.Env = body.Env
	spec.ExtraArgs = body.Args
	spec.ArtifactPatterns = artifactPatterns

	// Pin-to-commit: source the tree from a fresh remote fetch via Dagger,
	// inject auth where we have a stored token. The actual run's commit_sha
	// column is overridden below to reflect the pinned SHA.
	branchLabel := ""
	if body.AtCommit != "" {
		spec.CommitSHA = body.AtCommit
		spec.GitURL = gitURL
		if h.OAuth != nil {
			if tok := h.OAuth.AccessTokenForCloneURL(r.Context(), uid, gitURL); tok != "" {
				if p, ok := h.OAuth.ForCloneURL(gitURL); ok {
					spec.GitURL = p.InjectCloneAuth(gitURL, tok)
				}
			}
		}
		pinned := body.AtCommit
		commitSHA = &pinned
	} else if body.Branch != "" {
		// Run on a different branch without disturbing the project's checkout.
		// We fetch the branch's tip via ls-remote and pin the run to that SHA.
		// The branch label is persisted on the run row for later filtering.
		authToken := ""
		ghURL := gitURL
		if h.OAuth != nil {
			if tok := h.OAuth.AccessTokenForCloneURL(r.Context(), uid, gitURL); tok != "" {
				authToken = tok
				if p, ok := h.OAuth.ForCloneURL(gitURL); ok {
					ghURL = p.InjectCloneAuth(gitURL, tok)
				}
			}
		}
		sha, err := repo.LsRemoteBranchSHA(r.Context(), h.Dagger, gitURL, authToken, body.Branch)
		if err != nil || sha == "" {
			http.Error(w, "could not resolve branch '"+body.Branch+"' on remote", http.StatusBadRequest)
			return
		}
		spec.CommitSHA = sha
		spec.GitURL = ghURL
		commitSHA = &sha
		branchLabel = body.Branch
	}

	// Concurrent-run limit per user (active runs across all tasks).
	var activeForUser int
	if err := h.Pool.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM runs r
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE m.user_id = $1 AND r.status IN ('queued', 'running')
	`, uid).Scan(&activeForUser); err == nil && activeForUser >= MaxConcurrentRunsPerUser {
		http.Error(w, "concurrent run limit reached", http.StatusTooManyRequests)
		return
	}

	// Per-task concurrency policy. max_concurrency=0 means "use the legacy
	// single-run guard". Anything else uses a DB count of non-terminal runs
	// for this task (canonical across multi-instance schedulers) and the
	// supersede_policy decides whether to cancel an older run, reject, or
	// queue.
	h.mu.Lock()
	if maxConcurrency <= 0 {
		if _, busy := h.running[taskID]; busy {
			h.mu.Unlock()
			http.Error(w, "task already running", http.StatusConflict)
			return
		}
	} else {
		var active int
		_ = h.Pool.QueryRow(r.Context(), `
			SELECT COUNT(*) FROM runs
			WHERE task_id = $1 AND status IN ('queued','running','pending_approval')
		`, taskID).Scan(&active)
		if active >= maxConcurrency {
			switch supersedePolicy {
			case "cancel-old":
				// Cancel the oldest active run. The runner's executeAsync
				// path checks the row before stamping success, so a row
				// flipped to 'cancelled' here will short-circuit on the
				// runner side as a normal cancel.
				_, _ = h.Pool.Exec(r.Context(), `
					UPDATE runs
					SET status = 'cancelled', finished_at = COALESCE(finished_at, now())
					WHERE id = (
						SELECT id FROM runs
						WHERE task_id = $1 AND status IN ('queued','running','pending_approval')
						ORDER BY created_at ASC LIMIT 1
					)
				`, taskID)
			case "reject":
				h.mu.Unlock()
				http.Error(w, "task at concurrency limit", http.StatusTooManyRequests)
				return
			default:
				// 'queue' — there is no real queue today; treat as reject
				// with a clearer message so the user understands.
				h.mu.Unlock()
				http.Error(w, "task at concurrency limit (queue policy not yet implemented; configure cancel-old to supersede)", http.StatusTooManyRequests)
				return
			}
		}
	}

	paramsJSON, _ := json.Marshal(map[string]any{
		"env":  body.Env,
		"args": body.Args,
	})

	initialStatus := StatusQueued
	if requiresApproval {
		initialStatus = StatusPendingApproval
	}

	var runID uuid.UUID
	var branchInsert *string
	if branchLabel != "" {
		branchInsert = &branchLabel
	}
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO runs (task_id, project_id, commit_sha, status, log_path, params, branch)
		VALUES ($1, $2, $3, $4, '', $5::jsonb, $6)
		RETURNING id
	`, taskID, projectID, commitSHA, initialStatus, string(paramsJSON), branchInsert).Scan(&runID)
	if err != nil {
		h.mu.Unlock()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	spec.LogFile = filepath.Join(h.LogsRoot, runID.String()+".log")
	if _, err := h.Pool.Exec(r.Context(),
		`UPDATE runs SET log_path = $1 WHERE id = $2`, spec.LogFile, runID); err != nil {
		h.mu.Unlock()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Legacy single-task guard. Skip when max_concurrency is set, since the
	// DB-counted policy is now authoritative for those tasks.
	if !requiresApproval && maxConcurrency <= 0 {
		h.running[taskID] = runID
	}
	h.mu.Unlock()

	if h.Audit != nil {
		h.Audit.Record(r.Context(), uid, audit.RunStart, "run", runID.String(), r.RemoteAddr,
			map[string]any{"task_id": taskID.String(), "task_name": spec.Name, "task_source": spec.Source})
	}

	h.fireOutboundEvent(runID, events.EventRunCreated)

	if !requiresApproval {
		go h.executeAsync(taskID, runID, spec)
	}

	resp := Run{
		ID:        runID,
		TaskID:    taskID,
		ProjectID: projectID,
		CommitSHA: commitSHA,
		Branch:    branchInsert,
		Status:    initialStatus,
		LogPath:   spec.LogFile,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) executeAsync(taskID, runID uuid.UUID, spec Spec) {
	finalStatus := ""
	defer func() {
		h.mu.Lock()
		delete(h.running, taskID)
		delete(h.cancellers, runID)
		h.mu.Unlock()
		// After releasing the per-task lock, decide whether to schedule a
		// retry. Only "failed" (not cancelled, not timeout) is retried — a
		// timeout means the task ran longer than allowed, retrying won't help.
		if finalStatus == StatusFailed {
			h.maybeScheduleRetry(taskID, runID, spec)
		}
	}()

	timeout := time.Duration(spec.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	h.mu.Lock()
	h.cancellers[runID] = cancel
	h.mu.Unlock()

	now := time.Now()
	if _, err := h.Pool.Exec(ctx, `
		UPDATE runs SET status = $1, started_at = $2, updated_at = now() WHERE id = $3
	`, StatusRunning, now, runID); err != nil {
		h.Logger.Error("update run to running failed", "run", runID, "err", err)
		return
	}
	h.fireOutboundEvent(runID, events.EventRunStarted)

	if err := os.MkdirAll(filepath.Dir(spec.LogFile), 0o755); err != nil {
		h.markFailed(runID, -1, "create log dir: "+err.Error())
		return
	}

	// Dockerfile builds take a different code path: produce an image and
	// record it in the images table.
	if spec.Source == "dockerfile" {
		var projectID uuid.UUID
		_ = h.Pool.QueryRow(ctx, `SELECT project_id FROM runs WHERE id = $1`, runID).Scan(&projectID)
		var commitSHA *string
		_ = h.Pool.QueryRow(ctx, `SELECT commit_sha FROM runs WHERE id = $1`, runID).Scan(&commitSHA)

		digest, size, err := ExecuteImageBuild(ctx, h.Dagger, spec)
		buildFinished := time.Now()
		if err != nil {
			h.Logger.Error("image build failed", "run", runID, "err", err)
			_ = os.WriteFile(spec.LogFile, []byte("workend: image build failed: "+err.Error()), 0o644)
			h.markFinished(runID, StatusFailed, -1, buildFinished)
			return
		}
		summary := fmt.Sprintf("workend: built image %s (%d bytes)\n", digest, size)
		_ = os.WriteFile(spec.LogFile, []byte(summary), 0o644)

		commit := ""
		if commitSHA != nil {
			commit = *commitSHA
		}
		var imageID uuid.UUID
		if err := h.Pool.QueryRow(ctx, `
			INSERT INTO images (project_id, run_id, dockerfile_path, digest, size_bytes, commit_sha)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
			RETURNING id
		`, projectID, runID, spec.RawCommand, digest, size, commit).Scan(&imageID); err != nil {
			h.Logger.Warn("image insert failed", "run", runID, "err", err)
		} else if spec.RepoPath != "" {
			// Kick off Trivy and Syft in the background so the build's
			// user-visible duration is unaffected. Each writes its own
			// status column on the image row.
			go ScanImageAsync(h.Pool, h.Dagger, imageID, spec.RepoPath)
			go GenerateSBOMAsync(h.Pool, h.Dagger, imageID, spec.RepoPath)
		}
		h.markFinished(runID, StatusSucceeded, 0, buildFinished)
		finalStatus = StatusSucceeded
		return
	}

	result, err := Execute(ctx, h.Dagger, spec)
	finished := time.Now()
	if err != nil {
		// Distinguish: deadline-exceeded (system kill) vs user cancel vs
		// infrastructure failure. Append (don't overwrite) — the streaming
		// runner has already written whatever the user command produced
		// before the failure.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			appendLog(spec.LogFile, fmt.Sprintf("\nworkend: run timed out after %s\n", timeout))
			h.markTimedOut(runID, finished)
			return
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			appendLog(spec.LogFile, "\nworkend: run cancelled\n")
			h.markFinished(runID, StatusCancelled, -1, finished)
			finalStatus = StatusCancelled
			return
		}
		h.Logger.Error("run execute failed", "run", runID, "err", err)
		appendLog(spec.LogFile, "\nworkend: run failed: "+err.Error()+"\n")
		h.markFinished(runID, StatusFailed, -1, finished)
		finalStatus = StatusFailed
		return
	}

	status := StatusSucceeded
	if result.ExitCode != 0 {
		status = StatusFailed
	}
	h.markFinishedWithResources(runID, status, result.ExitCode, finished, result.Resources)
	finalStatus = status

	// Capture artifacts after the run terminates. Runs (success or failure)
	// can both produce useful artifacts (e.g. coverage on success, crash dump
	// on failure), so this fires for both. ArtifactsRoot must be configured;
	// when empty we silently skip.
	if h.ArtifactsRoot != "" && len(spec.ArtifactPatterns) > 0 && spec.RepoPath != "" {
		captured := artifact.Capture(ctx, h.Pool, h.ArtifactsRoot, runID, spec.RepoPath, spec.ArtifactPatterns)
		if captured > 0 {
			h.Logger.Info("captured artifacts", "run", runID, "count", captured)
		}
	}
}

// maybeScheduleRetry inspects the failed run's task retry policy and, when
// applicable, queues a follow-up run after the configured backoff. The new
// run reuses the same Spec (env, args, commit pin) so retries are
// reproducible. parent_run_id chains attempts together.
func (h *Handlers) maybeScheduleRetry(taskID, runID uuid.UUID, spec Spec) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var (
		retryMax   int
		backoffSec int
		attempt    int
		ownerID    uuid.UUID
	)
	err := h.Pool.QueryRow(ctx, `
		SELECT t.retry_max, t.retry_backoff_sec, r.attempt, w.user_id
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE r.id = $1
	`, runID).Scan(&retryMax, &backoffSec, &attempt, &ownerID)
	if err != nil || retryMax < 1 || attempt >= retryMax {
		return
	}
	if backoffSec < 1 {
		backoffSec = 30
	}

	h.Logger.Info("scheduling retry",
		"task", taskID, "parent_run", runID, "attempt", attempt+1, "of", retryMax+1, "backoff", backoffSec)

	go func() {
		time.Sleep(time.Duration(backoffSec) * time.Second)
		h.startRetry(taskID, runID, ownerID, attempt+1, spec)
	}()
}

// startRetry inserts a new queued run (linked to its parent), claims the
// per-task lock, and spawns the execute goroutine. Mirrors the lifecycle
// of EnqueueForUser but with retry-specific bookkeeping.
func (h *Handlers) startRetry(taskID, parentRunID, ownerID uuid.UUID, attempt int, spec Spec) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Pull the parent's commit_sha so the retry runs against the same
	// snapshot of the project state.
	var commitSHA *string
	_ = h.Pool.QueryRow(ctx, `SELECT commit_sha FROM runs WHERE id = $1`, parentRunID).Scan(&commitSHA)
	var projectID uuid.UUID
	_ = h.Pool.QueryRow(ctx, `SELECT project_id FROM runs WHERE id = $1`, parentRunID).Scan(&projectID)

	h.mu.Lock()
	if _, busy := h.running[taskID]; busy {
		h.mu.Unlock()
		h.Logger.Warn("retry skipped: task already running", "task", taskID, "parent", parentRunID)
		return
	}

	paramsJSON, _ := json.Marshal(map[string]any{"env": spec.Env, "args": spec.ExtraArgs})

	var runID uuid.UUID
	err := h.Pool.QueryRow(ctx, `
		INSERT INTO runs (task_id, project_id, commit_sha, status, log_path, params, attempt, parent_run_id)
		VALUES ($1, $2, $3, $4, '', $5::jsonb, $6, $7)
		RETURNING id
	`, taskID, projectID, commitSHA, StatusQueued, string(paramsJSON), attempt, parentRunID).Scan(&runID)
	if err != nil {
		h.mu.Unlock()
		h.Logger.Error("retry insert failed", "task", taskID, "err", err)
		return
	}
	spec.LogFile = filepath.Join(h.LogsRoot, runID.String()+".log")
	if _, err := h.Pool.Exec(ctx,
		`UPDATE runs SET log_path = $1 WHERE id = $2`, spec.LogFile, runID); err != nil {
		h.mu.Unlock()
		h.Logger.Error("retry log_path failed", "run", runID, "err", err)
		return
	}
	h.running[taskID] = runID
	h.mu.Unlock()

	if h.Audit != nil {
		h.Audit.Record(ctx, ownerID, audit.RunStart, "run", runID.String(), "retry",
			map[string]any{"task_id": taskID.String(), "parent_run_id": parentRunID.String(), "attempt": attempt})
	}

	go h.executeAsync(taskID, runID, spec)
}

// appendLog opens spec.LogFile in append mode and writes msg. Used by the
// error paths in executeAsync so the streaming runner's partial output
// isn't clobbered.
func appendLog(path, msg string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(msg)
}

// looksLikeSHA accepts hex strings between 7 (short SHA) and 64 (sha256)
// characters. We don't require a specific length so partial commits work,
// and Dagger resolves them server-side.
func looksLikeSHA(s string) bool {
	if len(s) < 7 || len(s) > 64 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !ok {
			return false
		}
	}
	return true
}

// validateRunInputs guards the user-supplied env + args before they hit the
// container. Env keys must look like POSIX env names; values cap at 4 KiB
// each; args cap at 32 entries × 4 KiB.
func validateRunInputs(env map[string]string, args []string) error {
	const maxLen = 4 * 1024
	const maxArgs = 32
	if len(args) > maxArgs {
		return fmt.Errorf("too many args (max %d)", maxArgs)
	}
	for _, a := range args {
		if len(a) > maxLen {
			return errors.New("arg too long")
		}
	}
	for k, v := range env {
		if k == "" || len(k) > 256 {
			return errors.New("invalid env name length")
		}
		for i := 0; i < len(k); i++ {
			c := k[i]
			ok := c == '_' ||
				(c >= 'A' && c <= 'Z') ||
				(c >= 'a' && c <= 'z') ||
				(i > 0 && c >= '0' && c <= '9')
			if !ok {
				return fmt.Errorf("invalid env name: %q", k)
			}
		}
		if len(v) > maxLen {
			return fmt.Errorf("env value too long for %q", k)
		}
	}
	return nil
}

// resolveTimeout picks the per-task override if set; otherwise falls back to
// the dispatcher's configured default; otherwise to a hard-coded 30m.
func (h *Handlers) resolveTimeout(perTask *int) int {
	if perTask != nil && *perTask > 0 {
		return *perTask
	}
	if h.DefaultTimeout > 0 {
		return int(h.DefaultTimeout / time.Second)
	}
	return 30 * 60
}

// markTimedOut records a run as failed-due-to-timeout. We use the existing
// failed status (so notifications and filters keep working unchanged) and
// flip the dedicated timed_out flag for UI/log distinction. Exit code -2
// is the marker convention.
func (h *Handlers) markTimedOut(runID uuid.UUID, finishedAt time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := h.Pool.Exec(ctx, `
		UPDATE runs
		SET status = $1, exit_code = $2, finished_at = $3, timed_out = true, updated_at = now()
		WHERE id = $4
	`, StatusFailed, -2, finishedAt, runID); err != nil {
		h.Logger.Error("mark timed out failed", "run", runID, "err", err)
	}
	h.fireNotification(runID, StatusFailed, -2, finishedAt)
}

func (h *Handlers) markFinished(runID uuid.UUID, status string, exitCode int, finishedAt time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := h.Pool.Exec(ctx, `
		UPDATE runs
		SET status = $1, exit_code = $2, finished_at = $3, updated_at = now()
		WHERE id = $4
	`, status, exitCode, finishedAt, runID); err != nil {
		h.Logger.Error("mark finished failed", "run", runID, "err", err)
	}
	h.fireNotification(runID, status, exitCode, finishedAt)
	h.fireOutboundEvent(runID, events.EventRunFinished)
}

// markFinishedWithResources is like markFinished but also persists the
// resource-usage tally captured by the runner. Used for non-Dockerfile runs
// where the wrapper script can read /proc.
func (h *Handlers) markFinishedWithResources(runID uuid.UUID, status string, exitCode int, finishedAt time.Time, res Resources) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := h.Pool.Exec(ctx, `
		UPDATE runs
		SET status = $1, exit_code = $2, finished_at = $3,
		    cpu_ms = NULLIF($4, 0)::BIGINT,
		    mem_peak_bytes = NULLIF($5, 0)::BIGINT,
		    net_rx_bytes = NULLIF($6, 0)::BIGINT,
		    net_tx_bytes = NULLIF($7, 0)::BIGINT,
		    updated_at = now()
		WHERE id = $8
	`, status, exitCode, finishedAt,
		res.CPUMs, res.MemPeakBytes, res.NetRxBytes, res.NetTxBytes,
		runID); err != nil {
		h.Logger.Error("mark finished w/ resources failed", "run", runID, "err", err)
	}
	h.fireNotification(runID, status, exitCode, finishedAt)
	h.fireOutboundEvent(runID, events.EventRunFinished)
}

// fireOutboundEvent emits a generic webhook event for the user that owns
// the run. Best-effort: failures are logged inside the dispatcher.
func (h *Handlers) fireOutboundEvent(runID uuid.UUID, eventType string) {
	if h.Events == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var (
		userID      uuid.UUID
		taskName    string
		taskSource  string
		projectName string
		status      string
		exitCode    *int
	)
	err := h.Pool.QueryRow(ctx, `
		SELECT w.user_id, t.name, t.source, p.name, r.status, r.exit_code
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE r.id = $1
	`, runID).Scan(&userID, &taskName, &taskSource, &projectName, &status, &exitCode)
	if err != nil {
		return
	}
	h.Events.Emit(userID, eventType, map[string]any{
		"run_id":       runID.String(),
		"task_name":    taskName,
		"task_source":  taskSource,
		"project_name": projectName,
		"status":       status,
		"exit_code":    exitCode,
		"url":          h.WebURL + "/runs/" + runID.String(),
	})
}

// fireNotification builds an Event from the run's metadata and invokes the
// dispatcher. Best-effort.
func (h *Handlers) fireNotification(runID uuid.UUID, status string, exitCode int, finishedAt time.Time) {
	if h.Notify == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		userID      uuid.UUID
		projectID   uuid.UUID
		projectName string
		taskName    string
		taskSource  string
		startedAt   *time.Time
		prevStatus  *string
		workspaceID uuid.UUID
		taskID      uuid.UUID
	)
	err := h.Pool.QueryRow(ctx, `
		SELECT w.user_id, w.id, p.id, t.id, p.name, t.name, t.source, r.started_at,
		       (SELECT status FROM runs r2
		        WHERE r2.task_id = r.task_id AND r2.id <> r.id
		        ORDER BY r2.created_at DESC LIMIT 1) AS prev_status
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE r.id = $1
	`, runID).Scan(&userID, &workspaceID, &projectID, &taskID, &projectName, &taskName, &taskSource, &startedAt, &prevStatus)
	if err != nil {
		h.Logger.Warn("notify: load event metadata failed", "run", runID, "err", err)
		return
	}

	durSec := 0
	if startedAt != nil {
		durSec = int(finishedAt.Sub(*startedAt).Seconds())
	}
	prev := ""
	if prevStatus != nil {
		prev = *prevStatus
	}

	h.Notify.OnRunComplete(notify.Event{
		RunID:       runID,
		UserID:      userID,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		TaskID:      taskID,
		ProjectName: projectName,
		TaskName:    taskName,
		TaskSource:  taskSource,
		Status:      status,
		PrevStatus:  prev,
		ExitCode:    exitCode,
		DurationSec: durSec,
		URL:         h.WebURL + "/runs/" + runID.String(),
	})

	// Web Push: short title/body so the browser's native notification chrome
	// stays readable. The service worker uses `tag` to coalesce repeated
	// runs of the same task.
	if h.Push != nil && h.Push.Enabled() {
		title := taskName + " " + status
		body := projectName + " · exit " + numToString(exitCode)
		h.Push.Send(userID, map[string]any{
			"title": title,
			"body":  body,
			"url":   h.WebURL + "/runs/" + runID.String(),
			"tag":   "workend-task-" + taskName,
		})
	}
}

func numToString(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [11]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func (h *Handlers) markFailed(runID uuid.UUID, exitCode int, msg string) {
	h.Logger.Warn("run failed", "run", runID, "msg", msg)
	h.markFinished(runID, StatusFailed, exitCode, time.Now())
}

// Compare returns metadata for two runs of the same task plus a line-level
// diff of their captured logs.
// GET /api/runs/:id/compare?to=<other_id>
func (h *Handlers) Compare(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	leftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}
	rightID, err := uuid.Parse(r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, "invalid 'to' run id", http.StatusBadRequest)
		return
	}

	left, err := h.fetchOwned(r.Context(), uid, leftID)
	if err != nil {
		http.Error(w, "left run not found", http.StatusNotFound)
		return
	}
	right, err := h.fetchOwned(r.Context(), uid, rightID)
	if err != nil {
		http.Error(w, "right run not found", http.StatusNotFound)
		return
	}
	if left.TaskID != right.TaskID {
		http.Error(w, "runs are of different tasks", http.StatusBadRequest)
		return
	}

	leftLog, _ := os.ReadFile(left.LogPath)
	rightLog, _ := os.ReadFile(right.LogPath)
	hunks := diff.Lines(string(leftLog), string(rightLog))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"left":  left,
		"right": right,
		"diff":  hunks,
	})
}

// EnqueuePipelineStep starts one step of a pipeline. The onDone callback
// fires after the child run reaches a terminal state (succeeded → true;
// failed/cancelled/timeout → false) and is what drives the next step.
//
// Mirrors EnqueueForUser but tags the new run row with pipeline_run_id +
// step number for the pipeline UI to render the timeline.
func (h *Handlers) EnqueuePipelineStep(ctx context.Context, taskID, ownerUserID, pipelineRunID uuid.UUID, step int, onDone func(success bool)) error {
	var (
		spec       Spec
		projectID  uuid.UUID
		commitSHA  *string
		projStatus string
		taskTOSec  *int
	)
	err := h.Pool.QueryRow(ctx, `
		SELECT t.source, t.name, t.raw_command, t.timeout_seconds,
		       p.id, p.local_path, p.last_commit_sha, p.status
		FROM tasks t
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE t.id = $1 AND m.user_id = $2
	`, taskID, ownerUserID).Scan(&spec.Source, &spec.Name, &spec.RawCommand, &taskTOSec,
		&projectID, &spec.RepoPath, &commitSHA, &projStatus)
	if err != nil {
		return fmt.Errorf("load task: %w", err)
	}
	if projStatus != "ready" || spec.RepoPath == "" {
		return errors.New("project not ready")
	}
	spec.TimeoutSec = h.resolveTimeout(taskTOSec)

	h.mu.Lock()
	if _, busy := h.running[taskID]; busy {
		h.mu.Unlock()
		return errors.New("task already running")
	}

	var runID uuid.UUID
	err = h.Pool.QueryRow(ctx, `
		INSERT INTO runs (task_id, project_id, commit_sha, status, log_path, pipeline_run_id, pipeline_step)
		VALUES ($1, $2, $3, $4, '', $5, $6)
		RETURNING id
	`, taskID, projectID, commitSHA, StatusQueued, pipelineRunID, step).Scan(&runID)
	if err != nil {
		h.mu.Unlock()
		return fmt.Errorf("insert pipeline-step run: %w", err)
	}
	spec.LogFile = filepath.Join(h.LogsRoot, runID.String()+".log")
	if _, err := h.Pool.Exec(ctx,
		`UPDATE runs SET log_path = $1 WHERE id = $2`, spec.LogFile, runID); err != nil {
		h.mu.Unlock()
		return fmt.Errorf("set log path: %w", err)
	}
	h.running[taskID] = runID
	h.mu.Unlock()

	go func() {
		h.executeAsync(taskID, runID, spec)
		// After execute, look up the final status and notify the pipeline.
		var finalStatus string
		_ = h.Pool.QueryRow(context.Background(),
			`SELECT status FROM runs WHERE id = $1`, runID).Scan(&finalStatus)
		if onDone != nil {
			onDone(finalStatus == StatusSucceeded)
		}
	}()
	return nil
}

// EnqueueForUser starts a run on behalf of a user (no HTTP context). Used
// by the scheduler. Concurrent-run limit and per-task lock are still
// enforced. Errors are returned for the caller to log.
func (h *Handlers) EnqueueForUser(ctx context.Context, taskID uuid.UUID, ownerUserID uuid.UUID) error {
	var (
		spec       Spec
		projectID  uuid.UUID
		commitSHA  *string
		projStatus string
		taskTOSec  *int
	)
	err := h.Pool.QueryRow(ctx, `
		SELECT t.source, t.name, t.raw_command, t.timeout_seconds,
		       p.id, p.local_path, p.last_commit_sha, p.status
		FROM tasks t
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE t.id = $1 AND m.user_id = $2
	`, taskID, ownerUserID).Scan(&spec.Source, &spec.Name, &spec.RawCommand, &taskTOSec,
		&projectID, &spec.RepoPath, &commitSHA, &projStatus)
	if err != nil {
		return fmt.Errorf("load task: %w", err)
	}
	if projStatus != "ready" || spec.RepoPath == "" {
		return errors.New("project not ready")
	}
	spec.TimeoutSec = h.resolveTimeout(taskTOSec)

	var activeForUser int
	if err := h.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM runs r
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE m.user_id = $1 AND r.status IN ('queued', 'running')
	`, ownerUserID).Scan(&activeForUser); err == nil && activeForUser >= MaxConcurrentRunsPerUser {
		return fmt.Errorf("concurrent run limit reached for user")
	}

	h.mu.Lock()
	if _, busy := h.running[taskID]; busy {
		h.mu.Unlock()
		return errors.New("task already running")
	}

	var runID uuid.UUID
	err = h.Pool.QueryRow(ctx, `
		INSERT INTO runs (task_id, project_id, commit_sha, status, log_path)
		VALUES ($1, $2, $3, $4, '')
		RETURNING id
	`, taskID, projectID, commitSHA, StatusQueued).Scan(&runID)
	if err != nil {
		h.mu.Unlock()
		return fmt.Errorf("insert run: %w", err)
	}
	spec.LogFile = filepath.Join(h.LogsRoot, runID.String()+".log")
	if _, err := h.Pool.Exec(ctx,
		`UPDATE runs SET log_path = $1 WHERE id = $2`, spec.LogFile, runID); err != nil {
		h.mu.Unlock()
		return fmt.Errorf("set log path: %w", err)
	}
	h.running[taskID] = runID
	h.mu.Unlock()

	if h.Audit != nil {
		h.Audit.Record(ctx, ownerUserID, audit.RunStart, "run", runID.String(), "scheduler",
			map[string]any{"task_id": taskID.String(), "task_name": spec.Name, "task_source": spec.Source, "via": "schedule"})
	}

	go h.executeAsync(taskID, runID, spec)
	return nil
}

// Approve transitions a pending_approval run into the queued state and
// kicks off execution. Decline marks the run as cancelled with an audit
// note. Either decision is recorded in run_approvals so the run page can
// show "approved/rejected by X at Y".
//
// POST /api/runs/:id/approve   body: {"approved": true|false, "note": "..."}
func (h *Handlers) Approve(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}
	var body struct {
		Approved bool   `json:"approved"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	current, err := h.fetchOwned(r.Context(), uid, runID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if current.Status != StatusPendingApproval {
		http.Error(w, "run is not awaiting approval", http.StatusConflict)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `
		INSERT INTO run_approvals (run_id, user_id, approved, note)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		ON CONFLICT (run_id) DO UPDATE
		SET user_id = EXCLUDED.user_id, approved = EXCLUDED.approved, note = EXCLUDED.note, decided_at = now()
	`, runID, uid, body.Approved, strings.TrimSpace(body.Note)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !body.Approved {
		h.markFinished(runID, StatusCancelled, -1, time.Now())
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": StatusCancelled})
		return
	}

	// Reload the run + task to rebuild a Spec, then queue execution.
	var (
		spec       Spec
		projectID  uuid.UUID
		commitSHA  *string
		projStatus string
		taskTOSec  *int
		gitURL     string
	)
	if err := h.Pool.QueryRow(r.Context(), `
		SELECT t.id, t.source, t.name, t.raw_command, t.timeout_seconds,
		       p.id, p.local_path, p.last_commit_sha, p.status, p.git_url
		FROM runs r
		JOIN tasks t    ON t.id = r.task_id
		JOIN projects p ON p.id = r.project_id
		WHERE r.id = $1
	`, runID).Scan(new(uuid.UUID), &spec.Source, &spec.Name, &spec.RawCommand, &taskTOSec,
		&projectID, &spec.RepoPath, &commitSHA, &projStatus, &gitURL); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if projStatus != "ready" || spec.RepoPath == "" {
		http.Error(w, "project not ready", http.StatusConflict)
		return
	}
	spec.TimeoutSec = h.resolveTimeout(taskTOSec)
	spec.LogFile = filepath.Join(h.LogsRoot, runID.String()+".log")

	// Re-hydrate env/args/commit pin from the stored params.
	rp := decodeParams(rawParamsForRun(r.Context(), h.Pool, runID))
	spec.Env = rp.Env
	spec.ExtraArgs = rp.Args
	if current.CommitSHA != nil && looksLikeSHA(*current.CommitSHA) {
		spec.CommitSHA = *current.CommitSHA
		spec.GitURL = gitURL
	}

	if _, err := h.Pool.Exec(r.Context(),
		`UPDATE runs SET status = $1, updated_at = now() WHERE id = $2`, StatusQueued, runID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.running[current.TaskID] = runID
	h.mu.Unlock()
	go h.executeAsync(current.TaskID, runID, spec)

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": StatusQueued})
}

// rawParamsForRun is a tiny helper to read the JSONB column as bytes.
func rawParamsForRun(ctx context.Context, pool *pgxpool.Pool, runID uuid.UUID) []byte {
	var raw []byte
	_ = pool.QueryRow(ctx, `SELECT params FROM runs WHERE id = $1`, runID).Scan(&raw)
	return raw
}

// Cancel signals the executing goroutine to abort.
// POST /api/runs/:id/cancel
func (h *Handlers) Cancel(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	if _, err := h.fetchOwned(r.Context(), uid, runID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.mu.Lock()
	cancelFn, found := h.cancellers[runID]
	h.mu.Unlock()

	if !found {
		// If it's not in memory, check if it's an orphaned job stuck in the database
		// due to a server restart. If so, mark it cancelled directly.
		var status string
		err := h.Pool.QueryRow(r.Context(), `SELECT status FROM runs WHERE id = $1`, runID).Scan(&status)
		if err == nil && (status == StatusRunning || status == StatusQueued) {
			h.markFinished(runID, StatusCancelled, -1, time.Now())
			if h.Audit != nil {
				h.Audit.Record(r.Context(), uid, audit.RunCancel, "run", runID.String(), r.RemoteAddr, map[string]any{"orphan": true})
			}
			w.WriteHeader(http.StatusAccepted)
			return
		}

		http.Error(w, "run not active", http.StatusConflict)
		return
	}

	cancelFn()
	if h.Audit != nil {
		h.Audit.Record(r.Context(), uid, audit.RunCancel, "run", runID.String(), r.RemoteAddr, nil)
	}
	w.WriteHeader(http.StatusAccepted)
}

// Stream emits the run's log as Server-Sent Events. While the run is active,
// the handler tail-follows the log file (1s polling) and emits new bytes as
// `log` events, then a final `done` event with status + exit code when the
// run reaches a terminal state.
//
// IMPORTANT (Stage 6 limitation): the underlying Dagger SDK's container.Stdout()
// only returns once the container exits, so today the log file is written
// once at the end of the run rather than incrementally. The SSE plumbing is
// correct — clients see the same "tail-follow → done" sequence — but expect
// a single large delivery at completion until the runner is reworked to
// produce incremental output (planned post-MVP).
//
// GET /api/runs/:id/log/stream
func (h *Handlers) Stream(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	run, err := h.fetchOwned(r.Context(), uid, runID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	logPath := run.LogPath
	var offset int64

	// Send any existing content
	if size, ok := readAndSend(w, logPath, 0); ok {
		offset = size
		flusher.Flush()
	}

	if isTerminal(run.Status) {
		sendDone(w, run.Status, run.ExitCode)
		flusher.Flush()
		return
	}

	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepAlive.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		case <-tick.C:
			if size, ok := readAndSend(w, logPath, offset); ok {
				offset = size
				flusher.Flush()
			}
			current, err := h.fetchOwned(r.Context(), uid, runID)
			if err != nil {
				return
			}
			if isTerminal(current.Status) {
				// Final flush before done event in case more bytes arrived
				if size, ok := readAndSend(w, logPath, offset); ok {
					offset = size
				}
				sendDone(w, current.Status, current.ExitCode)
				flusher.Flush()
				return
			}
		}
	}
}

func readAndSend(w io.Writer, path string, fromOffset int64) (newSize int64, sent bool) {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= fromOffset {
		return fromOffset, false
	}
	f, err := os.Open(path)
	if err != nil {
		return fromOffset, false
	}
	defer f.Close()
	if _, err := f.Seek(fromOffset, io.SeekStart); err != nil {
		return fromOffset, false
	}
	buf := make([]byte, info.Size()-fromOffset)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return fromOffset, false
	}
	if n == 0 {
		return fromOffset, false
	}
	sendEvent(w, "log", string(buf[:n]))
	return fromOffset + int64(n), true
}

func sendEvent(w io.Writer, event, data string) {
	// Each SSE message: "event: <name>\ndata: <line>\n... \n\n"
	// Multi-line data must be split into multiple data: lines.
	fmt.Fprintf(w, "event: %s\n", event)
	for i := 0; i < len(data); {
		j := i
		for j < len(data) && data[j] != '\n' {
			j++
		}
		fmt.Fprintf(w, "data: %s\n", data[i:j])
		i = j + 1
	}
	fmt.Fprint(w, "\n")
}

func sendDone(w io.Writer, status string, exitCode *int) {
	code := -1
	if exitCode != nil {
		code = *exitCode
	}
	payload, _ := json.Marshal(map[string]any{"status": status, "exit_code": code})
	sendEvent(w, "done", string(payload))
}

func isTerminal(status string) bool {
	return status == StatusSucceeded || status == StatusFailed || status == StatusCancelled
}

// Get returns one run with its log content (terminal state) or just metadata
// (live state).
// GET /api/runs/:id
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	run, err := h.fetchOwned(r.Context(), uid, runID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(run)
}

// GetLog returns the run's log file as plain text. Used by the SSE proxy
// fallback and as the post-completion view.
// GET /api/runs/:id/log
func (h *Handlers) GetLog(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	run, err := h.fetchOwned(r.Context(), uid, runID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	data, err := os.ReadFile(run.LogPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(MaskSecrets(string(data))))
}

// ListByProject returns runs for a project, most recent first.
// Optional ?status=<succeeded|failed|cancelled|running|queued> filters.
// GET /api/projects/:project_id/runs
func (h *Handlers) ListByProject(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "project_id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	if !userOwnsProject(r.Context(), h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	statusFilter := r.URL.Query().Get("status")
	args := []any{pid}
	whereStatus := ""
	if statusFilter != "" && validStatus(statusFilter) {
		args = append(args, statusFilter)
		whereStatus = " AND r.status = $2"
	}

	q := `
		SELECT r.id, r.task_id, r.project_id, r.commit_sha, r.status,
		       r.started_at, r.finished_at, r.exit_code, r.timed_out, r.log_path, r.params, r.attempt, r.parent_run_id, r.cpu_ms, r.mem_peak_bytes, r.net_rx_bytes, r.net_tx_bytes,
		       r.created_at, r.updated_at,
		       t.name, t.source
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.project_id = $1` + whereStatus + `
		ORDER BY r.created_at DESC
		LIMIT 200
	`
	rows, err := h.Pool.Query(r.Context(), q, args...)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Run{}
	for rows.Next() {
		var run Run
		var rawParams []byte
		if err := rows.Scan(&run.ID, &run.TaskID, &run.ProjectID, &run.CommitSHA, &run.Status,
			&run.StartedAt, &run.FinishedAt, &run.ExitCode, &run.TimedOut, &run.LogPath, &rawParams, &run.Attempt, &run.ParentRunID, &run.CPUMs, &run.MemPeakBytes, &run.NetRxBytes, &run.NetTxBytes,
			&run.CreatedAt, &run.UpdatedAt,
			&run.TaskName, &run.TaskSource); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		run.Params = decodeParams(rawParams)
		out = append(out, run)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ListForUser returns the most recent runs across every project the user
// owns. Used by the dashboard.
// GET /api/me/runs?limit=N
func (h *Handlers) ListForUser(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	limit := 25
	rows, err := h.Pool.Query(r.Context(), `
		SELECT r.id, r.task_id, r.project_id, r.commit_sha, r.status,
		       r.started_at, r.finished_at, r.exit_code, r.timed_out, r.log_path, r.params, r.attempt, r.parent_run_id, r.cpu_ms, r.mem_peak_bytes, r.net_rx_bytes, r.net_tx_bytes,
		       r.created_at, r.updated_at,
		       t.name, t.source,
		       p.name AS project_name, w.name AS workspace_name
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE m.user_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2
	`, uid, limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type dashRun struct {
		Run
		ProjectName   string `json:"project_name"`
		WorkspaceName string `json:"workspace_name"`
	}

	out := []dashRun{}
	for rows.Next() {
		var dr dashRun
		var rawParams []byte
		if err := rows.Scan(&dr.ID, &dr.TaskID, &dr.ProjectID, &dr.CommitSHA, &dr.Status,
			&dr.StartedAt, &dr.FinishedAt, &dr.ExitCode, &dr.TimedOut, &dr.LogPath, &rawParams, &dr.Attempt, &dr.ParentRunID, &dr.CPUMs, &dr.MemPeakBytes, &dr.NetRxBytes, &dr.NetTxBytes,
			&dr.CreatedAt, &dr.UpdatedAt,
			&dr.TaskName, &dr.TaskSource,
			&dr.ProjectName, &dr.WorkspaceName); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		dr.Params = decodeParams(rawParams)
		out = append(out, dr)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func validStatus(s string) bool {
	switch s {
	case StatusQueued, StatusRunning, StatusSucceeded, StatusFailed, StatusCancelled, StatusPendingApproval:
		return true
	}
	return false
}

// --- helpers ---

func (h *Handlers) fetchOwned(ctx context.Context, uid, runID uuid.UUID) (*Run, error) {
	var run Run
	var rawParams []byte
	err := h.Pool.QueryRow(ctx, `
		SELECT r.id, r.task_id, r.project_id, r.commit_sha, r.status,
		       r.started_at, r.finished_at, r.exit_code, r.timed_out, r.log_path, r.params, r.attempt, r.parent_run_id, r.cpu_ms, r.mem_peak_bytes, r.net_rx_bytes, r.net_tx_bytes,
		       r.created_at, r.updated_at,
		       t.name, t.source
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE r.id = $1 AND m.user_id = $2
	`, runID, uid).Scan(&run.ID, &run.TaskID, &run.ProjectID, &run.CommitSHA, &run.Status,
		&run.StartedAt, &run.FinishedAt, &run.ExitCode, &run.TimedOut, &run.LogPath, &rawParams, &run.Attempt, &run.ParentRunID, &run.CPUMs, &run.MemPeakBytes, &run.NetRxBytes, &run.NetTxBytes,
		&run.CreatedAt, &run.UpdatedAt,
		&run.TaskName, &run.TaskSource)
	if err != nil {
		return nil, err
	}
	run.Params = decodeParams(rawParams)
	return &run, nil
}

// decodeParams unmarshals the JSONB column into a RunParams. Returns an
// empty RunParams on any decode error or empty input — callers shouldn't
// fail just because old rows have legacy shapes.
func decodeParams(raw []byte) RunParams {
	var p RunParams
	if len(raw) == 0 {
		return p
	}
	_ = json.Unmarshal(raw, &p)
	return p
}

func userOwnsProject(ctx context.Context, pool *pgxpool.Pool, uid, pid uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&n)
	return err == nil
}
