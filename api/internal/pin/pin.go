// Package pin owns task pinning — per-user "favorite" tasks pulled to the
// top of the dashboard for one-click access.
package pin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

type Handlers struct {
	Pool *pgxpool.Pool
}

// PinnedTask is the payload returned by ListForUser. Joined fields make the
// dashboard render without a second round-trip.
type PinnedTask struct {
	TaskID        uuid.UUID `json:"task_id"`
	Position      int       `json:"position"`
	PinnedAt      time.Time `json:"pinned_at"`
	TaskName      string    `json:"task_name"`
	TaskSource    string    `json:"task_source"`
	TaskCommand   string    `json:"task_command"`
	ProjectID     uuid.UUID `json:"project_id"`
	ProjectName   string    `json:"project_name"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
}

// ListForUser returns the user's pinned tasks ordered by position.
// GET /api/me/pinned-tasks
func (h *Handlers) ListForUser(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.Pool.Query(r.Context(), `
		SELECT tp.task_id, tp.position, tp.pinned_at,
		       t.name, t.source, t.raw_command,
		       p.id, p.name,
		       w.id, w.name
		FROM task_pins tp
		JOIN tasks t      ON t.id = tp.task_id
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id AND m.user_id = tp.user_id
		WHERE tp.user_id = $1
		ORDER BY tp.position, tp.pinned_at
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []PinnedTask{}
	for rows.Next() {
		var pt PinnedTask
		if err := rows.Scan(&pt.TaskID, &pt.Position, &pt.PinnedAt,
			&pt.TaskName, &pt.TaskSource, &pt.TaskCommand,
			&pt.ProjectID, &pt.ProjectName,
			&pt.WorkspaceID, &pt.WorkspaceName); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, pt)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Pin adds a task to the user's pinned list. Idempotent.
// POST /api/tasks/:id/pin
func (h *Handlers) Pin(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	// Sanity: confirm the user actually has access to the task (membership
	// in its workspace). Cheaper to gate up front than via a foreign-key
	// stutter.
	var n int
	if err := h.Pool.QueryRow(r.Context(), `
		SELECT 1 FROM tasks t
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE t.id = $1 AND m.user_id = $2
	`, tid, uid).Scan(&n); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Append: position = max(position)+1.
	if _, err := h.Pool.Exec(r.Context(), `
		INSERT INTO task_pins (user_id, task_id, position)
		VALUES ($1, $2, COALESCE((SELECT MAX(position)+1 FROM task_pins WHERE user_id = $1), 0))
		ON CONFLICT (user_id, task_id) DO NOTHING
	`, uid, tid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unpin removes a task from the user's pinned list. Idempotent.
// DELETE /api/tasks/:id/pin
func (h *Handlers) Unpin(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}
	if _, err := h.Pool.Exec(r.Context(),
		`DELETE FROM task_pins WHERE user_id = $1 AND task_id = $2`, uid, tid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
