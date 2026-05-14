package repo

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

// Handlers exposes HTTP endpoints for repo-level operations.
type Handlers struct {
	Pool *pgxpool.Pool
}

// SearchInProject searches for a query string in a project's local repo.
// GET /api/projects/{id}/search?q=<term>&max=100
func (h *Handlers) SearchInProject(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "q parameter required", http.StatusBadRequest)
		return
	}
	if len(query) > 500 {
		http.Error(w, "q too long (max 500 chars)", http.StatusBadRequest)
		return
	}

	maxResults := 100
	if v := r.URL.Query().Get("max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			maxResults = n
		}
	}

	// Auth: user must be a member of the project's workspace.
	if !userOwnsProject(r, h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Get the project's local_path.
	var localPath *string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT local_path FROM projects WHERE id = $1
	`, pid).Scan(&localPath)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if localPath == nil || *localPath == "" {
		http.Error(w, "project has no local repo", http.StatusNotFound)
		return
	}

	results, err := Search(*localPath, query, maxResults)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []SearchResult{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

func userOwnsProject(r *http.Request, pool *pgxpool.Pool, uid, pid uuid.UUID) bool {
	var n int
	err := pool.QueryRow(r.Context(), `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&n)
	return err == nil
}
