package run

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"workend/api/internal/auth"
)

// archiveResp is returned by the archive / unarchive endpoints.
type archiveResp struct {
	ID         uuid.UUID  `json:"id"`
	ArchivedAt *time.Time `json:"archived_at"`
}

// ArchiveRun sets archived_at = now() for the specified run.
// POST /api/runs/{id}/archive
func (h *Handlers) ArchiveRun(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var resp archiveResp
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE runs SET archived_at = now()
		WHERE id = $1
		  AND id IN (
		    SELECT r.id FROM runs r
		    JOIN projects p ON p.id = r.project_id
		    JOIN workspaces ws ON ws.id = p.workspace_id
		    JOIN workspace_members m ON m.workspace_id = ws.id
		    WHERE r.id = $1 AND m.user_id = $2
		  )
		RETURNING id, archived_at
	`, runID, uid).Scan(&resp.ID, &resp.ArchivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// UnarchiveRun sets archived_at = NULL for the specified run.
// POST /api/runs/{id}/unarchive
func (h *Handlers) UnarchiveRun(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var resp archiveResp
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE runs SET archived_at = NULL
		WHERE id = $1
		  AND id IN (
		    SELECT r.id FROM runs r
		    JOIN projects p ON p.id = r.project_id
		    JOIN workspaces ws ON ws.id = p.workspace_id
		    JOIN workspace_members m ON m.workspace_id = ws.id
		    WHERE r.id = $1 AND m.user_id = $2
		  )
		RETURNING id, archived_at
	`, runID, uid).Scan(&resp.ID, &resp.ArchivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
