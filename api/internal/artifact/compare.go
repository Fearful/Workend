package artifact

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
)

// CompareResult holds the full diff between artifacts of two runs.
type CompareResult struct {
	LeftRunID  uuid.UUID      `json:"left_run_id"`
	RightRunID uuid.UUID      `json:"right_run_id"`
	Artifacts  []ArtifactDiff `json:"artifacts"`
}

// ArtifactDiff describes one path's difference between two runs.
type ArtifactDiff struct {
	Path      string `json:"path"`
	LeftSize  *int64 `json:"left_size"`
	RightSize *int64 `json:"right_size"`
	SizeDelta int64  `json:"size_delta"`
	Status    string `json:"status"` // "added", "removed", "changed", "unchanged"
}

// Compare returns a path-level diff of artifacts between two runs.
// GET /api/runs/{left_id}/artifacts/compare/{right_id}
func (h *Handlers) Compare(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	leftID, err := uuid.Parse(chi.URLParam(r, "left_id"))
	if err != nil {
		http.Error(w, "invalid left run id", http.StatusBadRequest)
		return
	}
	rightID, err := uuid.Parse(chi.URLParam(r, "right_id"))
	if err != nil {
		http.Error(w, "invalid right run id", http.StatusBadRequest)
		return
	}
	if !userOwnsRun(r.Context(), h.Pool, uid, leftID) {
		http.Error(w, "left run not found", http.StatusNotFound)
		return
	}
	if !userOwnsRun(r.Context(), h.Pool, uid, rightID) {
		http.Error(w, "right run not found", http.StatusNotFound)
		return
	}

	leftArtifacts, err := h.loadArtifacts(r, leftID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rightArtifacts, err := h.loadArtifacts(r, rightID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	leftByPath := make(map[string]Artifact, len(leftArtifacts))
	for _, a := range leftArtifacts {
		leftByPath[a.RelativePath] = a
	}
	rightByPath := make(map[string]Artifact, len(rightArtifacts))
	for _, a := range rightArtifacts {
		rightByPath[a.RelativePath] = a
	}

	// Collect all unique paths.
	seen := map[string]struct{}{}
	var paths []string
	for _, a := range leftArtifacts {
		if _, ok := seen[a.RelativePath]; !ok {
			seen[a.RelativePath] = struct{}{}
			paths = append(paths, a.RelativePath)
		}
	}
	for _, a := range rightArtifacts {
		if _, ok := seen[a.RelativePath]; !ok {
			seen[a.RelativePath] = struct{}{}
			paths = append(paths, a.RelativePath)
		}
	}

	diffs := make([]ArtifactDiff, 0, len(paths))
	for _, p := range paths {
		d := ArtifactDiff{Path: p}
		la, inLeft := leftByPath[p]
		ra, inRight := rightByPath[p]
		switch {
		case inLeft && !inRight:
			d.LeftSize = &la.SizeBytes
			d.SizeDelta = -la.SizeBytes
			d.Status = "removed"
		case !inLeft && inRight:
			d.RightSize = &ra.SizeBytes
			d.SizeDelta = ra.SizeBytes
			d.Status = "added"
		default:
			d.LeftSize = &la.SizeBytes
			d.RightSize = &ra.SizeBytes
			d.SizeDelta = ra.SizeBytes - la.SizeBytes
			if la.SizeBytes == ra.SizeBytes {
				d.Status = "unchanged"
			} else {
				d.Status = "changed"
			}
		}
		diffs = append(diffs, d)
	}

	result := CompareResult{
		LeftRunID:  leftID,
		RightRunID: rightID,
		Artifacts:  diffs,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handlers) loadArtifacts(r *http.Request, runID uuid.UUID) ([]Artifact, error) {
	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, run_id, relative_path, size_bytes, mime_type, captured_at
		FROM artifacts WHERE run_id = $1 ORDER BY relative_path
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Artifact
	for rows.Next() {
		var a Artifact
		if err := rows.Scan(&a.ID, &a.RunID, &a.RelativePath, &a.SizeBytes, &a.MimeType, &a.CapturedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}
