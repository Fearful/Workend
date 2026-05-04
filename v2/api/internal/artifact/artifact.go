// Package artifact stores per-run captured files declared by the task's
// artifact_patterns. List + download endpoints; capture itself is invoked
// from the run executor after the container finishes.
package artifact

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

// Storage layout: ARTIFACTS_ROOT/<run_id>/<sanitized relative path>
// We don't try to preserve directory structure — flatten and sanitize.

type Handlers struct {
	Pool          *pgxpool.Pool
	ArtifactsRoot string
	Logger        *slog.Logger
}

type Artifact struct {
	ID           uuid.UUID `json:"id"`
	RunID        uuid.UUID `json:"run_id"`
	RelativePath string    `json:"relative_path"`
	SizeBytes    int64     `json:"size_bytes"`
	MimeType     *string   `json:"mime_type"`
	CapturedAt   time.Time `json:"captured_at"`
}

// GET /api/runs/:id/artifacts
func (h *Handlers) ListByRun(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}
	if !userOwnsRun(r.Context(), h.Pool, uid, runID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, run_id, relative_path, size_bytes, mime_type, captured_at
		FROM artifacts WHERE run_id = $1 ORDER BY relative_path
	`, runID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Artifact{}
	for rows.Next() {
		var a Artifact
		if err := rows.Scan(&a.ID, &a.RunID, &a.RelativePath, &a.SizeBytes, &a.MimeType, &a.CapturedAt); err == nil {
			out = append(out, a)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GET /api/artifacts/:id/download
func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var (
		runID    uuid.UUID
		rel      string
		storage  string
		size     int64
		mime     *string
	)
	err = h.Pool.QueryRow(r.Context(), `
		SELECT a.run_id, a.relative_path, a.storage_path, a.size_bytes, a.mime_type
		FROM artifacts a
		WHERE a.id = $1
	`, id).Scan(&runID, &rel, &storage, &size, &mime)
	if err == pgx.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !userOwnsRun(r.Context(), h.Pool, uid, runID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	f, err := os.Open(storage)
	if err != nil {
		http.Error(w, "artifact missing on disk", http.StatusNotFound)
		return
	}
	defer f.Close()
	if mime != nil && *mime != "" {
		w.Header().Set("Content-Type", *mime)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(rel)+`"`)
	_, _ = io.Copy(w, f)
}

func userOwnsRun(ctx context.Context, pool *pgxpool.Pool, uid, runID uuid.UUID) bool {
	var ok bool
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM runs r
			JOIN projects p   ON p.id = r.project_id
			JOIN workspaces w ON w.id = p.workspace_id
			JOIN workspace_members m ON m.workspace_id = w.id
			WHERE r.id = $1 AND m.user_id = $2
		)
	`, runID, uid).Scan(&ok)
	return ok
}

// SanitizeRelPath converts a glob-matched repo-relative path to a stable,
// safe storage path. Replaces unfriendly chars and prefixes the run ID.
//
// Used by the run executor when capturing artifacts.
func SanitizeRelPath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	return reUnsafe.ReplaceAllString(p, "_")
}

var reUnsafe = regexp.MustCompile(`[^A-Za-z0-9_./-]`)
