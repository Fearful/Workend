package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

const (
	PreviewStatusBuilding = "building"
	PreviewStatusLive     = "live"
	PreviewStatusFailed   = "failed"
	PreviewStatusStopped  = "stopped"

	maxPreviewsPerProject = 10
)

// Preview represents a row in the previews table.
type Preview struct {
	ID             uuid.UUID  `json:"id"`
	ProjectID      uuid.UUID  `json:"project_id"`
	WorkspaceID    uuid.UUID  `json:"workspace_id"`
	Branch         string     `json:"branch"`
	Status         string     `json:"status"`
	URL            string     `json:"url"`
	DeployLog      string     `json:"deploy_log"`
	AutoDeploy     bool       `json:"auto_deploy"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	LastDeployedAt *time.Time `json:"last_deployed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PreviewHandlers holds dependencies for preview HTTP handlers.
type PreviewHandlers struct {
	Pool   *pgxpool.Pool
	WebURL string
	Logger *slog.Logger
}

type createPreviewReq struct {
	Branch     string `json:"branch"`
	AutoDeploy *bool  `json:"auto_deploy"`
}

type autoDeployResp struct {
	AutoDeploy bool `json:"auto_deploy"`
}

// Create creates a new preview for a project branch.
// POST /api/projects/{id}/previews
func (ph *PreviewHandlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	// Resolve workspace from project and check membership.
	wsID, err := projectWorkspace(r.Context(), ph.Pool, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), ph.Pool, uid, wsID); err != nil {
		http.Error(w, "not a workspace member", http.StatusForbidden)
		return
	}

	var req createPreviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Branch = strings.TrimSpace(req.Branch)
	if req.Branch == "" {
		http.Error(w, "branch required", http.StatusBadRequest)
		return
	}
	autoDeploy := true
	if req.AutoDeploy != nil {
		autoDeploy = *req.AutoDeploy
	}

	// Enforce per-project preview limit.
	var count int
	if err := ph.Pool.QueryRow(r.Context(), `
		SELECT count(*) FROM previews WHERE project_id = $1
	`, projectID).Scan(&count); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if count >= maxPreviewsPerProject {
		http.Error(w, fmt.Sprintf("max %d previews per project", maxPreviewsPerProject), http.StatusConflict)
		return
	}

	var p Preview
	err = ph.Pool.QueryRow(r.Context(), `
		INSERT INTO previews (project_id, workspace_id, branch, status, auto_deploy, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, workspace_id, branch, status, url, deploy_log,
		          auto_deploy, created_by, last_deployed_at, created_at, updated_at
	`, projectID, wsID, req.Branch, PreviewStatusBuilding, autoDeploy, uid).
		Scan(&p.ID, &p.ProjectID, &p.WorkspaceID, &p.Branch, &p.Status, &p.URL,
			&p.DeployLog, &p.AutoDeploy, &p.CreatedBy, &p.LastDeployedAt,
			&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "previews_project_id_branch_key") {
			http.Error(w, "preview already exists for this branch", http.StatusConflict)
			return
		}
		ph.Logger.Error("preview insert failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	p.URL = fmt.Sprintf("%s/preview/%s", ph.WebURL, p.ID)
	if _, err := ph.Pool.Exec(r.Context(), `
		UPDATE previews SET url = $1 WHERE id = $2
	`, p.URL, p.ID); err != nil {
		ph.Logger.Error("preview url update failed", "err", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

// List returns all previews for a project.
// GET /api/projects/{id}/previews
func (ph *PreviewHandlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	wsID, err := projectWorkspace(r.Context(), ph.Pool, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), ph.Pool, uid, wsID); err != nil {
		http.Error(w, "not a workspace member", http.StatusForbidden)
		return
	}

	rows, err := ph.Pool.Query(r.Context(), `
		SELECT id, project_id, workspace_id, branch, status, url, deploy_log,
		       auto_deploy, created_by, last_deployed_at, created_at, updated_at
		FROM previews
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Preview{}
	for rows.Next() {
		var p Preview
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.WorkspaceID, &p.Branch, &p.Status,
			&p.URL, &p.DeployLog, &p.AutoDeploy, &p.CreatedBy, &p.LastDeployedAt,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, p)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Get returns a single preview by ID.
// GET /api/previews/{id}
func (ph *PreviewHandlers) Get(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pvID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := ph.fetchWithMembership(r.Context(), uid, pvID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

// Redeploy triggers a new build for a preview.
// POST /api/previews/{id}/redeploy
func (ph *PreviewHandlers) Redeploy(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pvID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := ph.fetchWithMembership(r.Context(), uid, pvID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if p.Status != PreviewStatusLive && p.Status != PreviewStatusFailed {
		http.Error(w, "can only redeploy live or failed previews", http.StatusConflict)
		return
	}

	now := time.Now()
	if _, err := ph.Pool.Exec(r.Context(), `
		UPDATE previews SET status = $1, updated_at = $2 WHERE id = $3
	`, PreviewStatusBuilding, now, pvID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	p.Status = PreviewStatusBuilding
	p.UpdatedAt = now
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

// Stop sets a preview's status to stopped.
// POST /api/previews/{id}/stop
func (ph *PreviewHandlers) Stop(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pvID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := ph.fetchWithMembership(r.Context(), uid, pvID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	if _, err := ph.Pool.Exec(r.Context(), `
		UPDATE previews SET status = $1, updated_at = $2 WHERE id = $3
	`, PreviewStatusStopped, now, pvID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	p.Status = PreviewStatusStopped
	p.UpdatedAt = now
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

// Delete removes a preview. Only the creator or workspace owner may delete.
// DELETE /api/previews/{id}
func (ph *PreviewHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pvID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := ph.fetchWithMembership(r.Context(), uid, pvID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Only the creator or workspace owner may delete.
	if p.CreatedBy != uid {
		role, roleErr := workspace.RoleOf(r.Context(), ph.Pool, uid, p.WorkspaceID)
		if roleErr != nil || role != workspace.RoleOwner {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	if _, err := ph.Pool.Exec(r.Context(), `
		DELETE FROM previews WHERE id = $1
	`, pvID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ToggleAutoDeploy flips the auto_deploy flag on a preview.
// POST /api/previews/{id}/auto-deploy
func (ph *PreviewHandlers) ToggleAutoDeploy(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pvID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := ph.fetchWithMembership(r.Context(), uid, pvID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	newVal := !p.AutoDeploy
	now := time.Now()
	if _, err := ph.Pool.Exec(r.Context(), `
		UPDATE previews SET auto_deploy = $1, updated_at = $2 WHERE id = $3
	`, newVal, now, pvID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(autoDeployResp{AutoDeploy: newVal})
}

// --- helpers ---

// projectWorkspace returns the workspace_id for a given project.
func projectWorkspace(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) (uuid.UUID, error) {
	var wsID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT workspace_id FROM projects WHERE id = $1
	`, projectID).Scan(&wsID)
	return wsID, err
}

// fetchWithMembership fetches a preview and verifies the user is a member
// of the preview's workspace (via project). Returns pgx.ErrNoRows when not
// found or the user is not a member.
func (ph *PreviewHandlers) fetchWithMembership(ctx context.Context, uid, pvID uuid.UUID) (*Preview, error) {
	var p Preview
	err := ph.Pool.QueryRow(ctx, `
		SELECT pv.id, pv.project_id, pv.workspace_id, pv.branch, pv.status, pv.url,
		       pv.deploy_log, pv.auto_deploy, pv.created_by, pv.last_deployed_at,
		       pv.created_at, pv.updated_at
		FROM previews pv
		JOIN workspace_members m ON m.workspace_id = pv.workspace_id
		WHERE pv.id = $1 AND m.user_id = $2
	`, pvID, uid).Scan(&p.ID, &p.ProjectID, &p.WorkspaceID, &p.Branch, &p.Status,
		&p.URL, &p.DeployLog, &p.AutoDeploy, &p.CreatedBy, &p.LastDeployedAt,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
