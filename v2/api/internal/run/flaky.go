package run

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"workend/api/internal/auth"
)

// FlakyTask is a task that flips between succeeded and failed often enough
// to be untrustworthy. Surfaced on the dashboard so users notice them.
type FlakyTask struct {
	TaskID        uuid.UUID `json:"task_id"`
	TaskName      string    `json:"task_name"`
	TaskSource    string    `json:"task_source"`
	ProjectID     uuid.UUID `json:"project_id"`
	ProjectName   string    `json:"project_name"`
	WorkspaceName string    `json:"workspace_name"`
	Total         int       `json:"total"`         // recent runs scanned
	Succeeded     int       `json:"succeeded"`
	Failed        int       `json:"failed"`
	Flips         int       `json:"flips"`         // status transitions
	FlipRate      float64   `json:"flip_rate"`     // flips / max(1, total-1)
	SuccessRate   float64   `json:"success_rate"`  // succeeded / total
}

// ListFlaky returns tasks across the user's workspaces whose recent runs
// flip-flop between succeeded and failed.
//
// A "flip" is a status transition between consecutive completed runs (so a
// task with 10 runs that go S/F/S/F/S/F/S/F/S/F has 9 flips on 10 runs and
// is maximally flaky).
//
// Filters out tasks with too few completed runs to be meaningful.
//
// GET /api/me/flaky-tasks?window=N (default 20, max 100)
func (h *Handlers) ListFlaky(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	window := 20
	if v := r.URL.Query().Get("window"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 4 && n <= 100 {
			window = n
		}
	}

	// Pull each task's most-recent N completed runs and tally flips/totals
	// in one query using window functions. Restricted to runs the user has
	// access to via workspace membership.
	rows, err := h.Pool.Query(r.Context(), `
		WITH ranked AS (
			SELECT r.task_id, r.status,
			       ROW_NUMBER() OVER (PARTITION BY r.task_id ORDER BY r.created_at DESC) AS rn,
			       LAG(r.status) OVER (PARTITION BY r.task_id ORDER BY r.created_at) AS prev_status
			FROM runs r
			JOIN projects p   ON p.id = r.project_id
			JOIN workspaces w ON w.id = p.workspace_id
			JOIN workspace_members m ON m.workspace_id = w.id
			WHERE m.user_id = $1 AND r.status IN ('succeeded', 'failed')
		)
		SELECT t.id, t.name, t.source, p.id, p.name, w.name,
		       COUNT(*) AS total,
		       COUNT(*) FILTER (WHERE ranked.status = 'succeeded') AS ok,
		       COUNT(*) FILTER (WHERE ranked.status = 'failed') AS bad,
		       COUNT(*) FILTER (WHERE prev_status IS NOT NULL AND prev_status <> status) AS flips
		FROM ranked
		JOIN tasks t      ON t.id = ranked.task_id
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE ranked.rn <= $2
		GROUP BY t.id, t.name, t.source, p.id, p.name, w.name
		HAVING COUNT(*) >= 4
		   AND COUNT(*) FILTER (WHERE prev_status IS NOT NULL AND prev_status <> status) >= 2
		ORDER BY flips DESC, total DESC
		LIMIT 50
	`, uid, window)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []FlakyTask{}
	for rows.Next() {
		var f FlakyTask
		if err := rows.Scan(&f.TaskID, &f.TaskName, &f.TaskSource,
			&f.ProjectID, &f.ProjectName, &f.WorkspaceName,
			&f.Total, &f.Succeeded, &f.Failed, &f.Flips); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		f.SuccessRate = float64(f.Succeeded) / float64(f.Total)
		denom := f.Total - 1
		if denom < 1 {
			denom = 1
		}
		f.FlipRate = float64(f.Flips) / float64(denom)
		out = append(out, f)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
