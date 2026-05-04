package run

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
)

// TrendPoint is a single completed run's worth of duration data, plus the
// task it belongs to. Drives the duration-over-time chart on the project
// detail page.
type TrendPoint struct {
	RunID       uuid.UUID `json:"run_id"`
	TaskID      uuid.UUID `json:"task_id"`
	TaskName    string    `json:"task_name"`
	TaskSource  string    `json:"task_source"`
	Status      string    `json:"status"`
	TimedOut    bool      `json:"timed_out"`
	DurationSec int       `json:"duration_sec"`
	FinishedAt  time.Time `json:"finished_at"`
}

// Trends returns finished runs over a rolling window. Optional filters:
//
//	?task_id=<uuid> — only runs for that task
//	?days=N        — window size (default 30, max 365)
//
// GET /api/projects/:project_id/trends
func (h *Handlers) Trends(w http.ResponseWriter, r *http.Request) {
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

	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 365 {
			days = n
		}
	}
	since := time.Now().AddDate(0, 0, -days)

	args := []any{pid, since}
	taskFilter := ""
	if v := r.URL.Query().Get("task_id"); v != "" {
		tid, err := uuid.Parse(v)
		if err != nil {
			http.Error(w, "invalid task_id", http.StatusBadRequest)
			return
		}
		args = append(args, tid)
		taskFilter = " AND r.task_id = $3"
	}

	q := `
		SELECT r.id, r.task_id, t.name, t.source, r.status, r.timed_out,
		       EXTRACT(EPOCH FROM (r.finished_at - r.started_at))::int AS duration_sec,
		       r.finished_at
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.project_id = $1
		  AND r.started_at IS NOT NULL
		  AND r.finished_at IS NOT NULL
		  AND r.finished_at >= $2` + taskFilter + `
		ORDER BY r.finished_at ASC
		LIMIT 2000
	`
	rows, err := h.Pool.Query(r.Context(), q, args...)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []TrendPoint{}
	for rows.Next() {
		var p TrendPoint
		if err := rows.Scan(&p.RunID, &p.TaskID, &p.TaskName, &p.TaskSource,
			&p.Status, &p.TimedOut, &p.DurationSec, &p.FinishedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, p)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"days":   days,
		"points": out,
	})
}

// userOwnsProject is a thin re-implementation that mirrors the helper in
// run.go; kept private so the package is self-contained.
func init() {} // marker — no-op
