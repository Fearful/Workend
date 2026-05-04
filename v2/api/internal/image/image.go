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

// decodeSummary unmarshals the vuln_summary JSONB into a VulnSummaryView.
// Returns nil on empty / invalid data.
func decodeSummary(raw []byte) *VulnSummaryView {
	if len(raw) == 0 {
		return nil
	}
	var v VulnSummaryView
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return &v
}

type Image struct {
	ID              uuid.UUID  `json:"id"`
	ProjectID       uuid.UUID  `json:"project_id"`
	RunID           *uuid.UUID `json:"run_id"`
	DockerfilePath  string     `json:"dockerfile_path"`
	Digest          *string    `json:"digest"`
	SizeBytes       *int64     `json:"size_bytes"`
	CommitSHA       *string    `json:"commit_sha"`
	BuiltAt         time.Time  `json:"built_at"`
	ScanStatus      *string    `json:"scan_status"`
	ScanCompletedAt *time.Time `json:"scan_completed_at"`
	VulnSummary     *VulnSummaryView `json:"vuln_summary"`
	SBOMStatus      *string    `json:"sbom_status"`
	SBOMGeneratedAt *time.Time `json:"sbom_generated_at"`
	SBOMFormat      *string    `json:"sbom_format"`
}

// VulnSummaryView mirrors run.VulnSummary in JSON; we redeclare it here
// so the image package doesn't depend on run.
type VulnSummaryView struct {
	Critical int               `json:"critical"`
	High     int               `json:"high"`
	Medium   int               `json:"medium"`
	Low      int               `json:"low"`
	Unknown  int               `json:"unknown"`
	Top      []VulnSummaryRow  `json:"top"`
}

type VulnSummaryRow struct {
	ID       string `json:"id"`
	Package  string `json:"package"`
	Severity string `json:"severity"`
	FixedIn  string `json:"fixed_in,omitempty"`
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
		SELECT id, project_id, run_id, dockerfile_path, digest, size_bytes, commit_sha, built_at,
		       scan_status, scan_completed_at, vuln_summary,
		       sbom_status, sbom_generated_at, sbom_format
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
		var rawSummary []byte
		if err := rows.Scan(&img.ID, &img.ProjectID, &img.RunID, &img.DockerfilePath,
			&img.Digest, &img.SizeBytes, &img.CommitSHA, &img.BuiltAt,
			&img.ScanStatus, &img.ScanCompletedAt, &rawSummary,
			&img.SBOMStatus, &img.SBOMGeneratedAt, &img.SBOMFormat); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		img.VulnSummary = decodeSummary(rawSummary)
		out = append(out, img)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// SBOM streams the cached spdx-json document for one image.
// GET /api/images/:id/sbom
func (h *Handlers) SBOM(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var (
		pid       uuid.UUID
		raw       []byte
		format    *string
	)
	err = h.Pool.QueryRow(r.Context(), `
		SELECT project_id, sbom_data, sbom_format FROM images WHERE id = $1
	`, id).Scan(&pid, &raw, &format)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !userOwnsProject(r, h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if len(raw) == 0 {
		http.Error(w, "no SBOM yet", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="sbom-`+id.String()+`.json"`)
	_, _ = w.Write(raw)
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
