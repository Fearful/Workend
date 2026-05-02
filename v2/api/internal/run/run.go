package run

import (
	"context"
	"encoding/json"
	"errors"
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

	"workend/api/internal/auth"
	wdagger "workend/api/internal/dagger"
)

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

	// Per-task lock: only one run per task at a time.
	mu      sync.Mutex
	running map[uuid.UUID]uuid.UUID // task_id -> run_id
}

func (h *Handlers) Init() {
	h.running = map[uuid.UUID]uuid.UUID{}
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
		h.mu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

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

	result, err := Execute(ctx, h.Dagger, spec)
	finished := time.Now()
	if err != nil {
		h.Logger.Error("run execute failed", "run", runID, "err", err)
		// Persist the error message to the log file for the UI to surface.
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
}

func (h *Handlers) markFailed(runID uuid.UUID, exitCode int, msg string) {
	h.Logger.Warn("run failed", "run", runID, "msg", msg)
	h.markFinished(runID, StatusFailed, exitCode, time.Now())
}

// Get returns one run with its log content (Stage 5: full log inline; Stage 6
// will move to streaming and this stays as the terminal-state fallback).
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

// GetLog returns the run's log file as plain text. Caller streams it in S6.
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

	rows, err := h.Pool.Query(r.Context(), `
		SELECT r.id, r.task_id, r.project_id, r.commit_sha, r.status,
		       r.started_at, r.finished_at, r.exit_code, r.log_path,
		       r.created_at, r.updated_at,
		       t.name, t.source
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.project_id = $1
		ORDER BY r.created_at DESC
		LIMIT 200
	`, pid)
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
