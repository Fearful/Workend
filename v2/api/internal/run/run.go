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
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/audit"
	"workend/api/internal/auth"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/notify"
)

// MaxConcurrentRunsPerUser caps active (queued|running) runs per user.
// Configurable later; hard-coded for Stage 10.
const MaxConcurrentRunsPerUser = 3

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

type Run struct {
	ID         uuid.UUID  `json:"id"`
	TaskID     uuid.UUID  `json:"task_id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	CommitSHA  *string    `json:"commit_sha"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	ExitCode   *int       `json:"exit_code"`
	LogPath    string     `json:"log_path"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Hydrated on Get/List for UI convenience.
	TaskName   string `json:"task_name,omitempty"`
	TaskSource string `json:"task_source,omitempty"`
}

type Handlers struct {
	Pool     *pgxpool.Pool
	Dagger   *wdagger.Client
	LogsRoot string
	Logger   *slog.Logger
	Audit    *audit.Logger
	Notify   *notify.Dispatcher
	WebURL   string

	mu         sync.Mutex
	running    map[uuid.UUID]uuid.UUID            // task_id -> run_id
	cancellers map[uuid.UUID]context.CancelFunc   // run_id -> cancel func
}

func (h *Handlers) Init() {
	h.running = map[uuid.UUID]uuid.UUID{}
	h.cancellers = map[uuid.UUID]context.CancelFunc{}
}

// Create starts a new run for the given task.
// POST /api/tasks/:id/runs
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var (
		spec       Spec
		projectID  uuid.UUID
		commitSHA  *string
		projStatus string
	)
	err = h.Pool.QueryRow(r.Context(), `
		SELECT t.source, t.name, t.raw_command,
		       p.id, p.local_path, p.last_commit_sha, p.status
		FROM tasks t
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE t.id = $1 AND w.user_id = $2
	`, taskID, uid).Scan(&spec.Source, &spec.Name, &spec.RawCommand,
		&projectID, &spec.RepoPath, &commitSHA, &projStatus)
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

	// Concurrent-run limit per user (active runs across all tasks).
	var activeForUser int
	if err := h.Pool.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM runs r
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE w.user_id = $1 AND r.status IN ('queued', 'running')
	`, uid).Scan(&activeForUser); err == nil && activeForUser >= MaxConcurrentRunsPerUser {
		http.Error(w, "concurrent run limit reached", http.StatusTooManyRequests)
		return
	}

	h.mu.Lock()
	if _, busy := h.running[taskID]; busy {
		h.mu.Unlock()
		http.Error(w, "task already running", http.StatusConflict)
		return
	}

	var runID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO runs (task_id, project_id, commit_sha, status, log_path)
		VALUES ($1, $2, $3, $4, '')
		RETURNING id
	`, taskID, projectID, commitSHA, StatusQueued).Scan(&runID)
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

	h.running[taskID] = runID
	h.mu.Unlock()

	if h.Audit != nil {
		h.Audit.Record(r.Context(), uid, audit.RunStart, "run", runID.String(), r.RemoteAddr,
			map[string]any{"task_id": taskID.String(), "task_name": spec.Name, "task_source": spec.Source})
	}

	go h.executeAsync(taskID, runID, spec)

	resp := Run{
		ID:        runID,
		TaskID:    taskID,
		ProjectID: projectID,
		CommitSHA: commitSHA,
		Status:    StatusQueued,
		LogPath:   spec.LogFile,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) executeAsync(taskID, runID uuid.UUID, spec Spec) {
	defer func() {
		h.mu.Lock()
		delete(h.running, taskID)
		delete(h.cancellers, runID)
		h.mu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
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
		if _, err := h.Pool.Exec(ctx, `
			INSERT INTO images (project_id, run_id, dockerfile_path, digest, size_bytes, commit_sha)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
		`, projectID, runID, spec.RawCommand, digest, size, commit); err != nil {
			h.Logger.Warn("image insert failed", "run", runID, "err", err)
		}
		h.markFinished(runID, StatusSucceeded, 0, buildFinished)
		return
	}

	result, err := Execute(ctx, h.Dagger, spec)
	finished := time.Now()
	if err != nil {
		// Distinguish between user-initiated cancel and infrastructure failure.
		if errors.Is(ctx.Err(), context.Canceled) {
			_ = os.WriteFile(spec.LogFile, []byte("workend: run cancelled\n"), 0o644)
			h.markFinished(runID, StatusCancelled, -1, finished)
			return
		}
		h.Logger.Error("run execute failed", "run", runID, "err", err)
		_ = os.WriteFile(spec.LogFile, []byte("workend: run failed: "+err.Error()), 0o644)
		h.markFinished(runID, StatusFailed, -1, finished)
		return
	}

	status := StatusSucceeded
	if result.ExitCode != 0 {
		status = StatusFailed
	}
	h.markFinished(runID, status, result.ExitCode, finished)
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
	)
	err := h.Pool.QueryRow(ctx, `
		SELECT w.user_id, p.id, p.name, t.name, t.source, r.started_at,
		       (SELECT status FROM runs r2
		        WHERE r2.task_id = r.task_id AND r2.id <> r.id
		        ORDER BY r2.created_at DESC LIMIT 1) AS prev_status
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE r.id = $1
	`, runID).Scan(&userID, &projectID, &projectName, &taskName, &taskSource, &startedAt, &prevStatus)
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
		ProjectID:   projectID,
		ProjectName: projectName,
		TaskName:    taskName,
		TaskSource:  taskSource,
		Status:      status,
		PrevStatus:  prev,
		ExitCode:    exitCode,
		DurationSec: durSec,
		URL:         h.WebURL + "/runs/" + runID.String(),
	})
}

func (h *Handlers) markFailed(runID uuid.UUID, exitCode int, msg string) {
	h.Logger.Warn("run failed", "run", runID, "msg", msg)
	h.markFinished(runID, StatusFailed, exitCode, time.Now())
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
	)
	err := h.Pool.QueryRow(ctx, `
		SELECT t.source, t.name, t.raw_command,
		       p.id, p.local_path, p.last_commit_sha, p.status
		FROM tasks t
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE t.id = $1 AND w.user_id = $2
	`, taskID, ownerUserID).Scan(&spec.Source, &spec.Name, &spec.RawCommand,
		&projectID, &spec.RepoPath, &commitSHA, &projStatus)
	if err != nil {
		return fmt.Errorf("load task: %w", err)
	}
	if projStatus != "ready" || spec.RepoPath == "" {
		return errors.New("project not ready")
	}

	var activeForUser int
	if err := h.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM runs r
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE w.user_id = $1 AND r.status IN ('queued', 'running')
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
	_, _ = w.Write(data)
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
		       r.started_at, r.finished_at, r.exit_code, r.log_path,
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
		if err := rows.Scan(&run.ID, &run.TaskID, &run.ProjectID, &run.CommitSHA, &run.Status,
			&run.StartedAt, &run.FinishedAt, &run.ExitCode, &run.LogPath,
			&run.CreatedAt, &run.UpdatedAt,
			&run.TaskName, &run.TaskSource); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
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
		       r.started_at, r.finished_at, r.exit_code, r.log_path,
		       r.created_at, r.updated_at,
		       t.name, t.source,
		       p.name AS project_name, w.name AS workspace_name
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE w.user_id = $1
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
		if err := rows.Scan(&dr.ID, &dr.TaskID, &dr.ProjectID, &dr.CommitSHA, &dr.Status,
			&dr.StartedAt, &dr.FinishedAt, &dr.ExitCode, &dr.LogPath,
			&dr.CreatedAt, &dr.UpdatedAt,
			&dr.TaskName, &dr.TaskSource,
			&dr.ProjectName, &dr.WorkspaceName); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, dr)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func validStatus(s string) bool {
	switch s {
	case StatusQueued, StatusRunning, StatusSucceeded, StatusFailed, StatusCancelled:
		return true
	}
	return false
}

// --- helpers ---

func (h *Handlers) fetchOwned(ctx context.Context, uid, runID uuid.UUID) (*Run, error) {
	var run Run
	err := h.Pool.QueryRow(ctx, `
		SELECT r.id, r.task_id, r.project_id, r.commit_sha, r.status,
		       r.started_at, r.finished_at, r.exit_code, r.log_path,
		       r.created_at, r.updated_at,
		       t.name, t.source
		FROM runs r
		JOIN tasks t      ON t.id = r.task_id
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE r.id = $1 AND w.user_id = $2
	`, runID, uid).Scan(&run.ID, &run.TaskID, &run.ProjectID, &run.CommitSHA, &run.Status,
		&run.StartedAt, &run.FinishedAt, &run.ExitCode, &run.LogPath,
		&run.CreatedAt, &run.UpdatedAt,
		&run.TaskName, &run.TaskSource)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func userOwnsProject(ctx context.Context, pool *pgxpool.Pool, uid, pid uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE p.id = $1 AND w.user_id = $2
	`, pid, uid).Scan(&n)
	return err == nil
}
