// Package pipeline owns task chains: an ordered sequence of tasks that runs
// step-by-step and stops on the first failure. Each step is a real run with
// its own log + status; the pipeline_run row aggregates them.
package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"

	"workend/api/internal/auth"
)

// Runner is anything that can start a child run for a step. Implemented by
// run.Handlers.EnqueuePipelineStep so this package doesn't import run.
type Runner interface {
	EnqueuePipelineStep(ctx context.Context, taskID, ownerID, pipelineRunID uuid.UUID, step int, onDone func(success bool)) error
}

type Handlers struct {
	Pool   *pgxpool.Pool
	Runner Runner
	Logger *slog.Logger
}

type Pipeline struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Steps     []Step    `json:"steps,omitempty"`
}

type Step struct {
	Position      int       `json:"position"`
	TaskID        uuid.UUID `json:"task_id"`
	TaskName      string    `json:"task_name,omitempty"`
	TaskSource    string    `json:"task_source,omitempty"`
	ConditionExpr string    `json:"condition_expr,omitempty"`
	OnFailure     string    `json:"on_failure,omitempty"`
}

type PipelineRun struct {
	ID          uuid.UUID  `json:"id"`
	PipelineID  uuid.UUID  `json:"pipeline_id"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Children    []Child    `json:"children,omitempty"`
}

type Child struct {
	RunID    uuid.UUID `json:"run_id"`
	Position int       `json:"step"`
	Status   string    `json:"status"`
	TaskName string    `json:"task_name"`
}

// ListByProject returns the project's pipelines + their step lists.
// GET /api/projects/:project_id/pipelines
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
		SELECT id, project_id, name, created_at FROM pipelines
		WHERE project_id = $1 ORDER BY created_at DESC
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var pipelines []Pipeline
	for rows.Next() {
		var p Pipeline
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Name, &p.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		pipelines = append(pipelines, p)
	}
	for i := range pipelines {
		steps, _ := h.loadSteps(r.Context(), pipelines[i].ID)
		pipelines[i].Steps = steps
	}
	if pipelines == nil {
		pipelines = []Pipeline{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pipelines)
}

// Create defines a new pipeline.
// POST /api/projects/:project_id/pipelines
//   body: {"name":"ci","task_ids":["…","…","…"]}
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
	var body struct {
		Name    string      `json:"name"`
		TaskIDs []uuid.UUID `json:"task_ids"`
		Steps   []struct {
			TaskID        uuid.UUID `json:"task_id"`
			ConditionExpr string    `json:"condition_expr"`
			OnFailure     string    `json:"on_failure"`
		} `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.Name = strings.TrimSpace(body.Name)

	// Support both legacy task_ids array and the richer steps array.
	if len(body.Steps) == 0 && len(body.TaskIDs) > 0 {
		for _, tid := range body.TaskIDs {
			body.Steps = append(body.Steps, struct {
				TaskID        uuid.UUID `json:"task_id"`
				ConditionExpr string    `json:"condition_expr"`
				OnFailure     string    `json:"on_failure"`
			}{TaskID: tid})
		}
	}
	if body.Name == "" || len(body.Steps) == 0 {
		http.Error(w, "name and task_ids (or steps) required", http.StatusBadRequest)
		return
	}
	for _, s := range body.Steps {
		if len(s.ConditionExpr) > 200 {
			http.Error(w, "condition_expr too long (max 200)", http.StatusBadRequest)
			return
		}
		if s.OnFailure != "" && s.OnFailure != "stop" && s.OnFailure != "continue" && s.OnFailure != "skip_remaining" {
			http.Error(w, "on_failure must be stop, continue, or skip_remaining", http.StatusBadRequest)
			return
		}
	}

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var p Pipeline
	err = tx.QueryRow(r.Context(), `
		INSERT INTO pipelines (project_id, name) VALUES ($1, $2)
		RETURNING id, project_id, name, created_at
	`, pid, body.Name).Scan(&p.ID, &p.ProjectID, &p.Name, &p.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "pipelines_project_name_idx") {
			http.Error(w, "pipeline name already used in this project", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	for i, s := range body.Steps {
		onFailure := s.OnFailure
		if onFailure == "" {
			onFailure = "stop"
		}
		if _, err := tx.Exec(r.Context(), `
			INSERT INTO pipeline_steps (pipeline_id, position, task_id, condition_expr, on_failure)
			VALUES ($1, $2, $3, $4, $5)
		`, p.ID, i, s.TaskID, s.ConditionExpr, onFailure); err != nil {
			http.Error(w, "step insert failed: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	p.Steps, _ = h.loadSteps(r.Context(), p.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

// Delete removes a pipeline (children cascade via FK).
// DELETE /api/pipelines/:id
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if !h.userOwnsPipeline(r.Context(), uid, id) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if _, err := h.Pool.Exec(r.Context(), `DELETE FROM pipelines WHERE id = $1`, id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Run starts a pipeline_run + executes the first step. Subsequent steps
// chain via the runner's onDone callback.
// POST /api/pipelines/:id/runs
func (h *Handlers) Run(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pipID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if !h.userOwnsPipeline(r.Context(), uid, pipID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	steps, err := h.loadSteps(r.Context(), pipID)
	if err != nil || len(steps) == 0 {
		http.Error(w, "pipeline has no steps", http.StatusBadRequest)
		return
	}

	var prID uuid.UUID
	if err := h.Pool.QueryRow(r.Context(), `
		INSERT INTO pipeline_runs (pipeline_id, status, started_at)
		VALUES ($1, 'running', now()) RETURNING id
	`, pipID).Scan(&prID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	go h.driveSteps(uid, prID, steps, 0)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"pipeline_run_id": prID,
		"status":          "running",
	})
}

// driveSteps fires step `i`; on completion either advances to step `i+1`
// (success) or marks the pipeline_run failed (failure). Runs in the
// background; callers don't wait. prevSuccess tracks whether the previous
// step succeeded (true for the very first step).
func (h *Handlers) driveSteps(ownerID uuid.UUID, prID uuid.UUID, steps []Step, i int) {
	h.driveStepsWithStatus(ownerID, prID, steps, i, true)
}

func (h *Handlers) driveStepsWithStatus(ownerID uuid.UUID, prID uuid.UUID, steps []Step, i int, prevSuccess bool) {
	if i >= len(steps) {
		h.markPipelineRun(prID, "succeeded")
		return
	}
	step := steps[i]
	if h.Runner == nil {
		h.markPipelineRun(prID, "failed")
		return
	}

	// Evaluate the step's condition expression.
	if !shouldRunStep(step.ConditionExpr, prevSuccess) {
		// Skip this step and move to the next, preserving prevSuccess.
		h.driveStepsWithStatus(ownerID, prID, steps, i+1, prevSuccess)
		return
	}

	bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	onFailure := step.OnFailure
	if onFailure == "" {
		onFailure = "stop"
	}
	err := h.Runner.EnqueuePipelineStep(bgCtx, step.TaskID, ownerID, prID, step.Position, func(success bool) {
		if !success {
			switch onFailure {
			case "continue":
				h.driveStepsWithStatus(ownerID, prID, steps, i+1, false)
			case "skip_remaining":
				h.markPipelineRun(prID, "failed")
			default: // "stop"
				h.markPipelineRun(prID, "failed")
			}
			return
		}
		h.driveStepsWithStatus(ownerID, prID, steps, i+1, true)
	})
	if err != nil {
		if h.Logger != nil {
			h.Logger.Warn("pipeline step enqueue failed", "step", i, "err", err)
		}
		h.markPipelineRun(prID, "failed")
	}
}

// shouldRunStep evaluates a step's condition expression against the outcome
// of the previous step. An empty expression or "always" means unconditional.
func shouldRunStep(expr string, prevSuccess bool) bool {
	switch strings.TrimSpace(strings.ToLower(expr)) {
	case "", "always":
		return true
	case "on_success":
		return prevSuccess
	case "on_failure":
		return !prevSuccess
	default:
		// Unknown expressions default to always-run for forward compatibility.
		return true
	}
}

func (h *Handlers) markPipelineRun(prID uuid.UUID, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = h.Pool.Exec(ctx, `
		UPDATE pipeline_runs SET status = $1, finished_at = now() WHERE id = $2
	`, status, prID)
}

// GetRun returns one pipeline_run with its child runs hydrated.
// GET /api/pipeline-runs/:id
func (h *Handlers) GetRun(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	prID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var pr PipelineRun
	err = h.Pool.QueryRow(r.Context(), `
		SELECT pr.id, pr.pipeline_id, pr.status, pr.started_at, pr.finished_at, pr.created_at
		FROM pipeline_runs pr
		JOIN pipelines pp        ON pp.id = pr.pipeline_id
		JOIN projects p          ON p.id = pp.project_id
		JOIN workspaces w        ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE pr.id = $1 AND m.user_id = $2
	`, prID, uid).Scan(&pr.ID, &pr.PipelineID, &pr.Status, &pr.StartedAt, &pr.FinishedAt, &pr.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rows, err := h.Pool.Query(r.Context(), `
		SELECT r.id, r.pipeline_step, r.status, t.name
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.pipeline_run_id = $1
		ORDER BY r.pipeline_step
	`, prID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c Child
			var step *int
			if rows.Scan(&c.RunID, &step, &c.Status, &c.TaskName) == nil {
				if step != nil {
					c.Position = *step
				}
				pr.Children = append(pr.Children, c)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pr)
}

// --- internals ---

func (h *Handlers) loadSteps(ctx context.Context, pipelineID uuid.UUID) ([]Step, error) {
	rows, err := h.Pool.Query(ctx, `
		SELECT s.position, s.task_id, t.name, t.source, s.condition_expr, s.on_failure
		FROM pipeline_steps s
		JOIN tasks t ON t.id = s.task_id
		WHERE s.pipeline_id = $1
		ORDER BY s.position
	`, pipelineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []Step
	for rows.Next() {
		var s Step
		if err := rows.Scan(&s.Position, &s.TaskID, &s.TaskName, &s.TaskSource, &s.ConditionExpr, &s.OnFailure); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	return steps, nil
}

func (h *Handlers) userOwnsPipeline(ctx context.Context, userID, pipelineID uuid.UUID) bool {
	var n int
	err := h.Pool.QueryRow(ctx, `
		SELECT 1 FROM pipelines pp
		JOIN projects p          ON p.id = pp.project_id
		JOIN workspaces w        ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE pp.id = $1 AND m.user_id = $2
	`, pipelineID, userID).Scan(&n)
	return err == nil
}

func userOwnsProject(ctx context.Context, pool *pgxpool.Pool, userID, projectID uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE p.id = $1 AND m.user_id = $2
	`, projectID, userID).Scan(&n)
	return err == nil
}
