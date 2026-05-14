package workspace

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

// RetentionPolicy describes how long runs are kept in a workspace.
type RetentionPolicy struct {
	ID               uuid.UUID `json:"id"`
	WorkspaceID      uuid.UUID `json:"workspace_id"`
	MaxAgeDays       int       `json:"max_age_days"`
	MaxRuns          int       `json:"max_runs"`
	ArchiveAfterDays int       `json:"archive_after_days"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// defaultRetention returns a sensible default when no policy exists yet.
func defaultRetention(wsID uuid.UUID) RetentionPolicy {
	now := time.Now()
	return RetentionPolicy{
		WorkspaceID:      wsID,
		MaxAgeDays:       90,
		MaxRuns:          1000,
		ArchiveAfterDays: 30,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// GetRetention returns the retention policy for a workspace.
// If none is configured, returns defaults.
// GET /api/workspaces/{id}/retention
func (h *Handlers) GetRetention(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Membership check.
	if _, err := RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var p RetentionPolicy
	err = h.Pool.QueryRow(r.Context(), `
		SELECT id, workspace_id, max_age_days, max_runs, archive_after_days, created_at, updated_at
		FROM retention_policies
		WHERE workspace_id = $1
	`, wsID).Scan(&p.ID, &p.WorkspaceID, &p.MaxAgeDays, &p.MaxRuns, &p.ArchiveAfterDays, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		p = defaultRetention(wsID)
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

type setRetentionReq struct {
	MaxAgeDays       *int `json:"max_age_days"`
	MaxRuns          *int `json:"max_runs"`
	ArchiveAfterDays *int `json:"archive_after_days"`
}

// SetRetention upserts the retention policy for a workspace. Owner only.
// PUT /api/workspaces/{id}/retention
func (h *Handlers) SetRetention(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	role, err := RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if role != RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req setRetentionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Apply defaults for omitted fields, then validate.
	maxAge := 90
	if req.MaxAgeDays != nil {
		maxAge = *req.MaxAgeDays
	}
	if maxAge < 1 || maxAge > 3650 {
		http.Error(w, "max_age_days must be between 1 and 3650", http.StatusBadRequest)
		return
	}

	maxRuns := 1000
	if req.MaxRuns != nil {
		maxRuns = *req.MaxRuns
	}
	if maxRuns < 10 || maxRuns > 100000 {
		http.Error(w, "max_runs must be between 10 and 100000", http.StatusBadRequest)
		return
	}

	archiveAfter := 30
	if req.ArchiveAfterDays != nil {
		archiveAfter = *req.ArchiveAfterDays
	}
	if archiveAfter < 1 || archiveAfter > 3650 {
		http.Error(w, "archive_after_days must be between 1 and 3650", http.StatusBadRequest)
		return
	}

	var p RetentionPolicy
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO retention_policies (workspace_id, max_age_days, max_runs, archive_after_days)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace_id) DO UPDATE
		SET max_age_days       = EXCLUDED.max_age_days,
		    max_runs           = EXCLUDED.max_runs,
		    archive_after_days = EXCLUDED.archive_after_days,
		    updated_at         = now()
		RETURNING id, workspace_id, max_age_days, max_runs, archive_after_days, created_at, updated_at
	`, wsID, maxAge, maxRuns, archiveAfter).Scan(
		&p.ID, &p.WorkspaceID, &p.MaxAgeDays, &p.MaxRuns, &p.ArchiveAfterDays, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}
