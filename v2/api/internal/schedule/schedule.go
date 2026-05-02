// Package schedule manages cron-style schedules and a single in-process
// ticker that finds due schedules and enqueues runs.
package schedule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"workend/api/internal/audit"
	"workend/api/internal/auth"
	wdagger "workend/api/internal/dagger"
)

type Schedule struct {
	ID         uuid.UUID  `json:"id"`
	TaskID     uuid.UUID  `json:"task_id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	CronExpr   string     `json:"cron_expr"`
	Enabled    bool       `json:"enabled"`
	LastRunAt  *time.Time `json:"last_run_at"`
	NextRunAt  *time.Time `json:"next_run_at"`
	CreatedBy  uuid.UUID  `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	TaskName   string     `json:"task_name,omitempty"`
	TaskSource string     `json:"task_source,omitempty"`
}

// Enqueue is a callback the scheduler invokes when a schedule fires.
// Implemented by the runs package; abstracted to avoid an import cycle.
type Enqueue func(ctx context.Context, taskID uuid.UUID, ownerUserID uuid.UUID) error

type Handlers struct {
	Pool   *pgxpool.Pool
	Dagger *wdagger.Client
	Logger *slog.Logger
	Audit  *audit.Logger
	Parser cron.Parser
}

func NewHandlers(pool *pgxpool.Pool, dc *wdagger.Client, logger *slog.Logger, auditLog *audit.Logger) *Handlers {
	return &Handlers{
		Pool:   pool,
		Dagger: dc,
		Logger: logger,
		Audit:  auditLog,
		// Standard 5-field cron: minute hour day-of-month month day-of-week
		Parser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// StartTicker runs a goroutine that wakes every minute, finds due schedules,
// and calls enqueue. Returns a stop function.
func (h *Handlers) StartTicker(ctx context.Context, enqueue Enqueue) func() {
	ctx, cancel := context.WithCancel(ctx)
	go h.loop(ctx, enqueue)
	return cancel
}

func (h *Handlers) loop(ctx context.Context, enqueue Enqueue) {
	t := time.NewTicker(1 * time.Minute)
	defer t.Stop()
	// Run once immediately on startup so freshly-due schedules don't wait
	// up to a minute.
	h.tick(ctx, enqueue)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			h.tick(ctx, enqueue)
		}
	}
}

func (h *Handlers) tick(ctx context.Context, enqueue Enqueue) {
	rows, err := h.Pool.Query(ctx, `
		SELECT s.id, s.task_id, s.project_id, s.cron_expr, s.created_by, w.user_id
		FROM schedules s
		JOIN projects p   ON p.id = s.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE s.enabled = true
		  AND s.next_run_at IS NOT NULL
		  AND s.next_run_at <= now()
		ORDER BY s.next_run_at ASC
		LIMIT 50
	`)
	if err != nil {
		h.Logger.Warn("schedule tick query failed", "err", err)
		return
	}
	defer rows.Close()

	type due struct {
		id, taskID, projectID, createdBy, userID uuid.UUID
		expr                                     string
	}
	var dues []due
	for rows.Next() {
		var d due
		if err := rows.Scan(&d.id, &d.taskID, &d.projectID, &d.expr, &d.createdBy, &d.userID); err != nil {
			h.Logger.Warn("schedule scan failed", "err", err)
			continue
		}
		dues = append(dues, d)
	}
	rows.Close()

	for _, d := range dues {
		next, err := h.computeNext(d.expr, time.Now())
		if err != nil {
			h.Logger.Warn("schedule cron parse failed", "schedule", d.id, "err", err)
			continue
		}
		now := time.Now()
		if _, err := h.Pool.Exec(ctx, `
			UPDATE schedules SET last_run_at = $1, next_run_at = $2, updated_at = now()
			WHERE id = $3
		`, now, next, d.id); err != nil {
			h.Logger.Warn("schedule advance failed", "schedule", d.id, "err", err)
			continue
		}
		if err := enqueue(ctx, d.taskID, d.userID); err != nil {
			h.Logger.Warn("schedule enqueue failed", "schedule", d.id, "task", d.taskID, "err", err)
		}
	}
}

func (h *Handlers) computeNext(expr string, from time.Time) (time.Time, error) {
	sched, err := h.Parser.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(from), nil
}

// --- HTTP handlers ---

type createReq struct {
	TaskID   string `json:"task_id"`
	CronExpr string `json:"cron_expr"`
	Enabled  bool   `json:"enabled"`
}

// POST /api/projects/:project_id/schedules
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
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

	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		http.Error(w, "invalid task_id", http.StatusBadRequest)
		return
	}
	req.CronExpr = strings.TrimSpace(req.CronExpr)
	if req.CronExpr == "" {
		http.Error(w, "cron_expr required", http.StatusBadRequest)
		return
	}

	next, err := h.computeNext(req.CronExpr, time.Now())
	if err != nil {
		http.Error(w, "invalid cron expression: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Verify the task belongs to this project.
	var actualPID uuid.UUID
	if err := h.Pool.QueryRow(r.Context(),
		`SELECT project_id FROM tasks WHERE id = $1`, taskID).Scan(&actualPID); err != nil || actualPID != pid {
		http.Error(w, "task does not belong to project", http.StatusBadRequest)
		return
	}

	var s Schedule
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO schedules (task_id, project_id, cron_expr, enabled, next_run_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, task_id, project_id, cron_expr, enabled,
		          last_run_at, next_run_at, created_by, created_at, updated_at
	`, taskID, pid, req.CronExpr, req.Enabled, next, uid).
		Scan(&s.ID, &s.TaskID, &s.ProjectID, &s.CronExpr, &s.Enabled,
			&s.LastRunAt, &s.NextRunAt, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(s)
}

// GET /api/projects/:project_id/schedules
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
		SELECT s.id, s.task_id, s.project_id, s.cron_expr, s.enabled,
		       s.last_run_at, s.next_run_at, s.created_by, s.created_at, s.updated_at,
		       t.name, t.source
		FROM schedules s
		JOIN tasks t ON t.id = s.task_id
		WHERE s.project_id = $1
		ORDER BY s.created_at DESC
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Schedule{}
	for rows.Next() {
		var s Schedule
		if err := rows.Scan(&s.ID, &s.TaskID, &s.ProjectID, &s.CronExpr, &s.Enabled,
			&s.LastRunAt, &s.NextRunAt, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
			&s.TaskName, &s.TaskSource); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, s)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// DELETE /api/schedules/:id
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	tag, err := h.Pool.Exec(r.Context(), `
		DELETE FROM schedules
		WHERE id = $1
		  AND project_id IN (
		    SELECT p.id FROM projects p
		    JOIN workspaces w ON w.id = p.workspace_id
		    WHERE w.user_id = $2
		  )
	`, id, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/schedules/:id/toggle
func (h *Handlers) Toggle(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var newEnabled bool
	var expr string
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE schedules s
		SET enabled = NOT enabled, updated_at = now()
		WHERE s.id = $1
		  AND s.project_id IN (
		    SELECT p.id FROM projects p
		    JOIN workspaces w ON w.id = p.workspace_id
		    WHERE w.user_id = $2
		  )
		RETURNING enabled, cron_expr
	`, id, uid).Scan(&newEnabled, &expr)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Re-arm next_run_at if newly enabled
	if newEnabled {
		next, err := h.computeNext(expr, time.Now())
		if err == nil {
			_, _ = h.Pool.Exec(r.Context(),
				`UPDATE schedules SET next_run_at = $1 WHERE id = $2`, next, id)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"enabled": newEnabled})
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

// fmtCron is a tiny helper unused-import-killer (keep for future).
var _ = fmt.Sprintf
