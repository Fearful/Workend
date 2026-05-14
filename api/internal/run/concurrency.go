package run

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"workend/api/internal/auth"
)

type concurrencyStatus struct {
	MaxConcurrency int    `json:"max_concurrency"`
	ActiveRuns     int    `json:"active_runs"`
	QueuedRuns     int    `json:"queued_runs"`
	Policy         string `json:"policy"`
}

// GetConcurrencyStatus returns the concurrency state for a task: limits,
// active count, queued count, and supersede policy.
// GET /api/tasks/{id}/concurrency
func (h *Handlers) GetConcurrencyStatus(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var maxConcurrency int
	var supersedePolicy string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT t.max_concurrency, t.supersede_policy
		FROM tasks t
		JOIN projects p          ON p.id = t.project_id
		JOIN workspaces w        ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE t.id = $1 AND m.user_id = $2
	`, taskID, uid).Scan(&maxConcurrency, &supersedePolicy)
	if err == pgx.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var activeRuns, queuedRuns int
	_ = h.Pool.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM runs WHERE task_id = $1 AND status = 'running'
	`, taskID).Scan(&activeRuns)
	_ = h.Pool.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM runs WHERE task_id = $1 AND status = 'queued'
	`, taskID).Scan(&queuedRuns)

	policy := supersedePolicy
	if policy == "" {
		policy = "queue"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(concurrencyStatus{
		MaxConcurrency: maxConcurrency,
		ActiveRuns:     activeRuns,
		QueuedRuns:     queuedRuns,
		Policy:         policy,
	})
}
