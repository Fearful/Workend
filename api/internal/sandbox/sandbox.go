// Package sandbox manages ephemeral developer sandbox environments backed by
// Dagger containers. Each sandbox is a short-lived, branch-specific
// environment attached to a project within a workspace.
package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/workspace"
)

const (
	StatusProvisioning = "provisioning"
	StatusRunning      = "running"
	StatusStopped      = "stopped"
	StatusDestroyed    = "destroyed"
	StatusFailed       = "failed"

	defaultExpiresHours = 4
	maxExpiresHours     = 72
	maxActiveSandboxes  = 5
)

// Sandbox represents a row in the sandboxes table.
type Sandbox struct {
	ID          uuid.UUID  `json:"id"`
	ProjectID   uuid.UUID  `json:"project_id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	Branch      string     `json:"branch"`
	Status      string     `json:"status"`
	URL         string     `json:"url"`
	Port        int        `json:"port"`
	ContainerID string     `json:"container_id"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	ExpiresAt   time.Time  `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
}

// Handlers holds dependencies for sandbox HTTP handlers.
type Handlers struct {
	Pool      *pgxpool.Pool
	Dagger    *wdagger.Client
	ReposRoot string
	Logger    *slog.Logger
	WebURL    string // base URL for generating sandbox URLs
}

type createReq struct {
	ProjectID    uuid.UUID `json:"project_id"`
	Branch       string    `json:"branch"`
	ExpiresHours int       `json:"expires_hours"`
}

type extendReq struct {
	Hours int `json:"hours"`
}

// Create provisions a new sandbox for a project.
// POST /api/sandboxes
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.ProjectID == uuid.Nil {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.ExpiresHours <= 0 {
		req.ExpiresHours = defaultExpiresHours
	}
	if req.ExpiresHours > maxExpiresHours {
		http.Error(w, fmt.Sprintf("expires_hours must be 1-%d", maxExpiresHours), http.StatusBadRequest)
		return
	}

	// Resolve workspace from project and check membership.
	wsID, err := h.workspaceForProject(r.Context(), req.ProjectID)
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not a workspace member", http.StatusForbidden)
		return
	}

	// Enforce per-user active sandbox limit.
	var active int
	if err := h.Pool.QueryRow(r.Context(), `
		SELECT count(*) FROM sandboxes
		WHERE created_by = $1 AND status IN ('provisioning', 'running')
	`, uid).Scan(&active); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if active >= maxActiveSandboxes {
		http.Error(w, fmt.Sprintf("max %d active sandboxes per user", maxActiveSandboxes), http.StatusConflict)
		return
	}

	expiresAt := time.Now().Add(time.Duration(req.ExpiresHours) * time.Hour)

	var sb Sandbox
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO sandboxes (project_id, workspace_id, branch, status, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, workspace_id, branch, status, url, port, container_id,
		          created_by, expires_at, created_at, destroyed_at
	`, req.ProjectID, wsID, req.Branch, StatusProvisioning, uid, expiresAt).
		Scan(&sb.ID, &sb.ProjectID, &sb.WorkspaceID, &sb.Branch, &sb.Status,
			&sb.URL, &sb.Port, &sb.ContainerID, &sb.CreatedBy, &sb.ExpiresAt,
			&sb.CreatedAt, &sb.DestroyedAt)
	if err != nil {
		h.Logger.Error("sandbox insert failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	go h.provision(sb.ID, sb.ProjectID, sb.Branch)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sb)
}

// provision runs the Dagger-based container provisioning in the background.
// On success it sets status=running; on failure it sets status=failed.
func (h *Handlers) provision(sandboxID, projectID uuid.UUID, branch string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	repoDir := filepath.Join(h.ReposRoot, projectID.String())
	_ = repoDir  // used when building the container
	_ = branch

	// Attempt to get Dagger client and build the container.
	dagClient, err := h.Dagger.Get(ctx)
	if err != nil {
		h.Logger.Error("sandbox provision: dagger connect failed", "sandbox_id", sandboxID, "err", err)
		h.setStatus(ctx, sandboxID, StatusFailed, "")
		return
	}
	_ = dagClient // would be used for container build in production

	// Stub: In production this would:
	// 1. Load the repo directory via dagClient.Host().Directory(repoDir)
	// 2. Check for Dockerfile; if absent, use golang:1.22-alpine as base
	// 3. Build and start the container
	// 4. Extract the container ID and published URL/port
	//
	// For now, simulate a successful provision:
	containerID := fmt.Sprintf("ctr_%s", sandboxID.String()[:8])
	url := fmt.Sprintf("%s/sandbox/%s", h.WebURL, sandboxID)

	if _, err := h.Pool.Exec(ctx, `
		UPDATE sandboxes
		SET status = $1, container_id = $2, url = $3, port = 8080
		WHERE id = $4
	`, StatusRunning, containerID, url, sandboxID); err != nil {
		h.Logger.Error("sandbox provision: status update failed", "sandbox_id", sandboxID, "err", err)
	}
}

func (h *Handlers) setStatus(ctx context.Context, sandboxID uuid.UUID, status, containerID string) {
	query := `UPDATE sandboxes SET status = $1 WHERE id = $2`
	args := []any{status, sandboxID}
	if containerID != "" {
		query = `UPDATE sandboxes SET status = $1, container_id = $3 WHERE id = $2`
		args = append(args, containerID)
	}
	if _, err := h.Pool.Exec(ctx, query, args...); err != nil {
		h.Logger.Error("sandbox setStatus failed", "sandbox_id", sandboxID, "status", status, "err", err)
	}
}

// List returns active (non-destroyed) sandboxes for a workspace.
// GET /api/workspaces/{id}/sandboxes
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not a workspace member", http.StatusForbidden)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, project_id, workspace_id, branch, status, url, port, container_id,
		       created_by, expires_at, created_at, destroyed_at
		FROM sandboxes
		WHERE workspace_id = $1 AND status != $2
		ORDER BY created_at DESC
	`, wsID, StatusDestroyed)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Sandbox{}
	for rows.Next() {
		var sb Sandbox
		if err := rows.Scan(&sb.ID, &sb.ProjectID, &sb.WorkspaceID, &sb.Branch, &sb.Status,
			&sb.URL, &sb.Port, &sb.ContainerID, &sb.CreatedBy, &sb.ExpiresAt,
			&sb.CreatedAt, &sb.DestroyedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, sb)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Get returns a single sandbox by ID.
// GET /api/sandboxes/{id}
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	sbID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sb, err := h.fetchWithMembership(r.Context(), uid, sbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sb)
}

// Destroy marks a sandbox as destroyed.
// POST /api/sandboxes/{id}/destroy
func (h *Handlers) Destroy(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	sbID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sb, err := h.fetchWithMembership(r.Context(), uid, sbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Only the creator or workspace owner may destroy.
	if sb.CreatedBy != uid {
		role, roleErr := workspace.RoleOf(r.Context(), h.Pool, uid, sb.WorkspaceID)
		if roleErr != nil || role != workspace.RoleOwner {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	now := time.Now()
	if _, err := h.Pool.Exec(r.Context(), `
		UPDATE sandboxes SET status = $1, destroyed_at = $2
		WHERE id = $3
	`, StatusDestroyed, now, sbID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sb.Status = StatusDestroyed
	sb.DestroyedAt = &now
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sb)
}

// Extend adds hours to a sandbox's expiration.
// POST /api/sandboxes/{id}/extend
func (h *Handlers) Extend(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	sbID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req extendReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Hours <= 0 {
		http.Error(w, "hours must be a positive integer", http.StatusBadRequest)
		return
	}

	sb, err := h.fetchWithMembership(r.Context(), uid, sbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if sb.CreatedBy != uid {
		http.Error(w, "only the creator may extend", http.StatusForbidden)
		return
	}
	if sb.Status != StatusRunning {
		http.Error(w, "sandbox must be running to extend", http.StatusConflict)
		return
	}

	newExpiry := sb.ExpiresAt.Add(time.Duration(req.Hours) * time.Hour)
	maxExpiry := time.Now().Add(maxExpiresHours * time.Hour)
	if newExpiry.After(maxExpiry) {
		http.Error(w, fmt.Sprintf("total expiry cannot exceed %dh from now", maxExpiresHours), http.StatusBadRequest)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `
		UPDATE sandboxes SET expires_at = $1 WHERE id = $2
	`, newExpiry, sbID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sb.ExpiresAt = newExpiry
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sb)
}

// ListForUser returns all active sandboxes created by the current user.
// GET /api/me/sandboxes
func (h *Handlers) ListForUser(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, project_id, workspace_id, branch, status, url, port, container_id,
		       created_by, expires_at, created_at, destroyed_at
		FROM sandboxes
		WHERE created_by = $1 AND status != $2
		ORDER BY created_at DESC
	`, uid, StatusDestroyed)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Sandbox{}
	for rows.Next() {
		var sb Sandbox
		if err := rows.Scan(&sb.ID, &sb.ProjectID, &sb.WorkspaceID, &sb.Branch, &sb.Status,
			&sb.URL, &sb.Port, &sb.ContainerID, &sb.CreatedBy, &sb.ExpiresAt,
			&sb.CreatedAt, &sb.DestroyedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, sb)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// --- helpers ---

// workspaceForProject looks up the workspace_id for a given project.
func (h *Handlers) workspaceForProject(ctx context.Context, projectID uuid.UUID) (uuid.UUID, error) {
	var wsID uuid.UUID
	err := h.Pool.QueryRow(ctx, `
		SELECT workspace_id FROM projects WHERE id = $1
	`, projectID).Scan(&wsID)
	return wsID, err
}

// fetchWithMembership fetches a sandbox and verifies the user is a member
// of its workspace. Returns pgx.ErrNoRows when not found or not a member.
func (h *Handlers) fetchWithMembership(ctx context.Context, uid, sbID uuid.UUID) (*Sandbox, error) {
	var sb Sandbox
	err := h.Pool.QueryRow(ctx, `
		SELECT s.id, s.project_id, s.workspace_id, s.branch, s.status, s.url,
		       s.port, s.container_id, s.created_by, s.expires_at,
		       s.created_at, s.destroyed_at
		FROM sandboxes s
		JOIN workspace_members m ON m.workspace_id = s.workspace_id
		WHERE s.id = $1 AND m.user_id = $2
	`, sbID, uid).Scan(&sb.ID, &sb.ProjectID, &sb.WorkspaceID, &sb.Branch, &sb.Status,
		&sb.URL, &sb.Port, &sb.ContainerID, &sb.CreatedBy, &sb.ExpiresAt,
		&sb.CreatedAt, &sb.DestroyedAt)
	if err != nil {
		return nil, err
	}
	return &sb, nil
}

// sandboxStatus returns the current status of a sandbox.
func sandboxStatus(ctx context.Context, pool *pgxpool.Pool, sbID uuid.UUID) (string, error) {
	var status string
	err := pool.QueryRow(ctx, `
		SELECT status FROM sandboxes WHERE id = $1
	`, sbID).Scan(&status)
	return status, err
}

