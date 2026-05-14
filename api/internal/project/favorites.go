package project

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

// FavoriteProject is the JSON shape returned by the favorites and recent-
// projects endpoints. It carries the denormalised project name + workspace so
// the UI can render a list without a follow-up call.
type FavoriteProject struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	FavoritedAt time.Time `json:"favorited_at,omitempty"`
	ViewedAt    time.Time `json:"viewed_at,omitempty"`
}

// FavoriteHandlers groups the HTTP handlers for project favorites and recent
// project views. Kept separate from the main project.Handlers because it has
// no dependency on Dagger, OAuth, etc.
type FavoriteHandlers struct {
	Pool *pgxpool.Pool
}

// ListFavorites returns the authenticated user's favorited projects.
//
// GET /api/me/favorites
func (h *FavoriteHandlers) ListFavorites(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT f.project_id, p.name, p.workspace_id, f.created_at
		FROM project_favorites f
		JOIN projects p ON p.id = f.project_id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []FavoriteProject{}
	for rows.Next() {
		var fp FavoriteProject
		if err := rows.Scan(&fp.ProjectID, &fp.ProjectName, &fp.WorkspaceID, &fp.FavoritedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, fp)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// AddFavorite marks a project as a favorite for the authenticated user. The
// user must be a member of the project's workspace.
//
// POST /api/projects/{id}/favorite
func (h *FavoriteHandlers) AddFavorite(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if !userIsMember(r.Context(), h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `
		INSERT INTO project_favorites (user_id, project_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, uid, pid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveFavorite removes a project from the user's favorites.
//
// DELETE /api/projects/{id}/favorite
func (h *FavoriteHandlers) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `
		DELETE FROM project_favorites WHERE user_id = $1 AND project_id = $2
	`, uid, pid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListRecent returns the 10 most recently viewed projects for the user.
//
// GET /api/me/recent-projects
func (h *FavoriteHandlers) ListRecent(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT v.project_id, p.name, p.workspace_id, v.viewed_at
		FROM project_views v
		JOIN projects p ON p.id = v.project_id
		WHERE v.user_id = $1
		ORDER BY v.viewed_at DESC
		LIMIT 10
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []FavoriteProject{}
	for rows.Next() {
		var fp FavoriteProject
		if err := rows.Scan(&fp.ProjectID, &fp.ProjectName, &fp.WorkspaceID, &fp.ViewedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, fp)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// RecordView upserts a project view timestamp. The user must be a member of
// the project's workspace.
//
// POST /api/projects/{id}/view
func (h *FavoriteHandlers) RecordView(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if !userIsMember(r.Context(), h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `
		INSERT INTO project_views (user_id, project_id, viewed_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id, project_id) DO UPDATE SET viewed_at = now()
	`, uid, pid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// userIsMember checks whether the user belongs to the workspace that owns the
// given project. Used by favorites and views to enforce access control.
func userIsMember(ctx context.Context, pool *pgxpool.Pool, uid, projectID uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE p.id = $1 AND m.user_id = $2
	`, projectID, uid).Scan(&n)
	return err == nil
}
