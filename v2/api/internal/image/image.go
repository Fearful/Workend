// Package image lists the images built per project.
package image

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

type Image struct {
	ID             uuid.UUID  `json:"id"`
	ProjectID      uuid.UUID  `json:"project_id"`
	RunID          *uuid.UUID `json:"run_id"`
	DockerfilePath string     `json:"dockerfile_path"`
	Digest         *string    `json:"digest"`
	SizeBytes      *int64     `json:"size_bytes"`
	CommitSHA      *string    `json:"commit_sha"`
	BuiltAt        time.Time  `json:"built_at"`
}

type Handlers struct {
	Pool *pgxpool.Pool
}

// GET /api/projects/:project_id/images
func (h *Handlers) ListByProject(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "project_id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	if !userOwnsProject(r, h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, project_id, run_id, dockerfile_path, digest, size_bytes, commit_sha, built_at
		FROM images
		WHERE project_id = $1
		ORDER BY built_at DESC
		LIMIT 100
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Image{}
	for rows.Next() {
		var img Image
		if err := rows.Scan(&img.ID, &img.ProjectID, &img.RunID, &img.DockerfilePath,
			&img.Digest, &img.SizeBytes, &img.CommitSHA, &img.BuiltAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, img)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func userOwnsProject(r *http.Request, pool *pgxpool.Pool, uid, pid uuid.UUID) bool {
	var n int
	err := pool.QueryRow(r.Context(), `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE p.id = $1 AND w.user_id = $2
	`, pid, uid).Scan(&n)
	return err == nil
}
