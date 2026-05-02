// Package project owns project (= a git repo within a workspace) CRUD and
// the background workers that ingest / sync repos via Dagger.
package project

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/detect"
	"workend/api/internal/oauth"
	"workend/api/internal/quota"
	"workend/api/internal/repo"
	"workend/api/internal/stats"
)

const (
	StatusPending = "pending"
	StatusCloning = "cloning"
	StatusReady   = "ready"
	StatusError   = "error"
)

type Project struct {
	ID                uuid.UUID  `json:"id"`
	WorkspaceID       uuid.UUID  `json:"workspace_id"`
	Name              string     `json:"name"`
	GitURL            string     `json:"git_url"`
	DefaultBranch     *string    `json:"default_branch"`
	LocalPath         *string    `json:"local_path"`
	Status            string     `json:"status"`
	LastCommitSHA     *string    `json:"last_commit_sha"`
	LastCommitMessage *string    `json:"last_commit_message"`
	LastCommitAuthor  *string    `json:"last_commit_author"`
	LastSyncedAt      *time.Time `json:"last_synced_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	// Stage 19/24: only populated by Get (single-project), not List, to avoid
	// surfacing webhook secrets/tokens in bulk responses.
	WebhookToken     string  `json:"webhook_token,omitempty"`
	WebhookSecretSet bool    `json:"webhook_secret_set"`
	WebhookSecret    *string `json:"webhook_secret,omitempty"` // only on the immediate response of SetWebhookSecret
}

type Handlers struct {
	Pool      *pgxpool.Pool
	Dagger    *wdagger.Client
	OAuth     *oauth.Registry // nil-safe; provider lookup returns no token when nil
	ReposRoot string
	Logger    *slog.Logger
}

// List returns projects in a given workspace (must belong to the user).
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "workspace_id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}

	if !h.userOwnsWorkspace(r.Context(), uid, wsID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, workspace_id, name, git_url, default_branch, local_path,
		       status, last_commit_sha, last_commit_message, last_commit_author,
		       last_synced_at, created_at, updated_at
		FROM projects
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.GitURL, &p.DefaultBranch, &p.LocalPath,
			&p.Status, &p.LastCommitSHA, &p.LastCommitMessage, &p.LastCommitAuthor,
			&p.LastSyncedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, p)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := h.fetchOwned(r.Context(), uid, pid)
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

type createReq struct {
	Name   string `json:"name"`
	GitURL string `json:"git_url"`
	Branch string `json:"branch"`
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "workspace_id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}

	if !h.userOwnsWorkspace(r.Context(), uid, wsID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.GitURL = strings.TrimSpace(req.GitURL)
	if req.Name == "" || req.GitURL == "" {
		http.Error(w, "name and git_url required", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(req.GitURL, "http://") && !strings.HasPrefix(req.GitURL, "https://") {
		http.Error(w, "git_url must be http(s)", http.StatusBadRequest)
		return
	}

	var p Project
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO projects (workspace_id, name, git_url, default_branch, status, webhook_token)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, encode(gen_random_bytes(24), 'base64'))
		RETURNING id, workspace_id, name, git_url, default_branch, local_path,
		          status, last_commit_sha, last_commit_message, last_commit_author,
		          last_synced_at, created_at, updated_at, webhook_token
	`, wsID, req.Name, req.GitURL, req.Branch, StatusPending).
		Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.GitURL, &p.DefaultBranch, &p.LocalPath,
			&p.Status, &p.LastCommitSHA, &p.LastCommitMessage, &p.LastCommitAuthor,
			&p.LastSyncedAt, &p.CreatedAt, &p.UpdatedAt, &p.WebhookToken)
	if err != nil {
		if strings.Contains(err.Error(), "projects_workspace_name_idx") {
			http.Error(w, "project name already used in this workspace", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	go h.cloneAsync(uid, p.ID, wsID, req.GitURL, req.Branch)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

func (h *Handlers) Sync(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	p, err := h.fetchOwned(r.Context(), uid, pid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	branch := ""
	if p.DefaultBranch != nil {
		branch = *p.DefaultBranch
	}

	if _, err := h.Pool.Exec(r.Context(),
		`UPDATE projects SET status = $1, updated_at = now() WHERE id = $2`,
		StatusCloning, pid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	go h.cloneAsync(uid, pid, p.WorkspaceID, p.GitURL, branch)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := h.fetchOwned(r.Context(), uid, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `DELETE FROM projects WHERE id = $1`, pid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if p.LocalPath != nil && *p.LocalPath != "" {
		if err := os.RemoveAll(*p.LocalPath); err != nil {
			h.Logger.Warn("failed to remove repo dir", "path", *p.LocalPath, "err", err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func (h *Handlers) userOwnsWorkspace(ctx context.Context, uid, wsID uuid.UUID) bool {
	var n int
	if err := h.Pool.QueryRow(ctx, `
		SELECT 1 FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, wsID, uid).Scan(&n); err != nil {
		return false
	}
	return true
}

func (h *Handlers) fetchOwned(ctx context.Context, uid, pid uuid.UUID) (*Project, error) {
	var p Project
	var secret *string
	err := h.Pool.QueryRow(ctx, `
		SELECT p.id, p.workspace_id, p.name, p.git_url, p.default_branch, p.local_path,
		       p.status, p.last_commit_sha, p.last_commit_message, p.last_commit_author,
		       p.last_synced_at, p.created_at, p.updated_at, p.webhook_token, p.webhook_secret
		FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.GitURL, &p.DefaultBranch, &p.LocalPath,
		&p.Status, &p.LastCommitSHA, &p.LastCommitMessage, &p.LastCommitAuthor,
		&p.LastSyncedAt, &p.CreatedAt, &p.UpdatedAt, &p.WebhookToken, &secret)
	if err != nil {
		return nil, err
	}
	p.WebhookSecretSet = secret != nil && *secret != ""
	return &p, nil
}

// SetWebhookSecret regenerates (or clears, with ?clear=true) the project's
// webhook signing secret. Returns the new plain-text secret in the
// response body (it's only shown this one time). Members can call.
//
// POST /api/projects/:id/webhook-secret
func (h *Handlers) SetWebhookSecret(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := h.fetchOwned(r.Context(), uid, pid); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if r.URL.Query().Get("clear") == "true" {
		if _, err := h.Pool.Exec(r.Context(),
			`UPDATE projects SET webhook_secret = NULL WHERE id = $1`, pid); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var newSecret string
	if err := h.Pool.QueryRow(r.Context(), `
		UPDATE projects
		SET webhook_secret = encode(gen_random_bytes(24), 'base64')
		WHERE id = $1
		RETURNING webhook_secret
	`, pid).Scan(&newSecret); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"webhook_secret": newSecret})
}

// Webhook handles a public push event from a git provider. Authn is the
// per-project token in the URL path; if the project also has a webhook_secret
// set, the request must additionally carry a valid HMAC signature in one of
// the per-provider headers (see verifySignature).
//
// POST /api/webhooks/projects/{token}
func (h *Handlers) Webhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	var (
		projectID, workspaceID uuid.UUID
		creatorID              uuid.UUID
		gitURL                 string
		defaultBranch          *string
		secret                 *string
	)
	err := h.Pool.QueryRow(r.Context(), `
		SELECT p.id, p.workspace_id, w.user_id, p.git_url, p.default_branch, p.webhook_secret
		FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE p.webhook_token = $1
	`, token).Scan(&projectID, &workspaceID, &creatorID, &gitURL, &defaultBranch, &secret)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if secret != nil && *secret != "" {
		body, err := io.ReadAll(io.LimitReader(r.Body, 5<<20)) // 5 MB cap
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		if !verifySignature(r.Header, body, *secret) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	if _, err := h.Pool.Exec(r.Context(),
		`UPDATE projects SET status = $1, updated_at = now() WHERE id = $2`,
		StatusCloning, projectID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	branch := ""
	if defaultBranch != nil {
		branch = *defaultBranch
	}
	// Use the workspace creator's identity for token lookup (the closest
	// proxy we have for "the user who set up this connection").
	go h.cloneAsync(creatorID, projectID, workspaceID, gitURL, branch)

	w.WriteHeader(http.StatusAccepted)
}

// verifySignature accepts a request as authentic if any of the supported
// per-provider headers validates against the shared secret. Constant-time
// comparison throughout.
//
//	GitHub:  X-Hub-Signature-256: sha256=<hex of HMAC-SHA256(secret, body)>
//	Gitea:   X-Gitea-Signature:   <hex of HMAC-SHA256(secret, body)>
//	GitLab:  X-Gitlab-Token:      <secret> (plain shared token, no HMAC)
func verifySignature(headers http.Header, body []byte, secret string) bool {
	if got := headers.Get("X-Hub-Signature-256"); got != "" {
		want := "sha256=" + hmacHex(body, secret)
		return hmac.Equal([]byte(got), []byte(want))
	}
	if got := headers.Get("X-Gitea-Signature"); got != "" {
		want := hmacHex(body, secret)
		return hmac.Equal([]byte(got), []byte(want))
	}
	if got := headers.Get("X-Gitlab-Token"); got != "" {
		return hmac.Equal([]byte(got), []byte(secret))
	}
	return false
}

func hmacHex(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *Handlers) cloneAsync(userID, projectID, workspaceID uuid.UUID, gitURL, branch string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := quota.CheckOrErr(ctx, h.Pool, h.ReposRoot, workspaceID); err != nil {
		h.markError(projectID, err)
		return
	}

	dest := filepath.Join(h.ReposRoot, workspaceID.String(), projectID.String())

	authToken := h.lookupAuth(ctx, userID, gitURL)

	if _, err := h.Pool.Exec(ctx,
		`UPDATE projects SET status = $1, updated_at = now() WHERE id = $2`,
		StatusCloning, projectID); err != nil {
		h.Logger.Error("update status to cloning failed", "err", err, "project", projectID)
		return
	}

	// Wipe any prior content for clean re-clone.
	if err := os.RemoveAll(dest); err != nil {
		h.markError(projectID, fmt.Errorf("remove old dest: %w", err))
		return
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		h.markError(projectID, fmt.Errorf("mkdir parent: %w", err))
		return
	}

	result, err := repo.Clone(ctx, h.Dagger, gitURL, branch, authToken, dest)
	if err != nil {
		h.markError(projectID, err)
		return
	}

	now := time.Now()
	if _, err := h.Pool.Exec(ctx, `
		UPDATE projects
		SET status = $1,
		    local_path = $2,
		    default_branch = COALESCE(NULLIF($3, ''), default_branch),
		    last_commit_sha = $4,
		    last_commit_message = NULLIF($5, ''),
		    last_commit_author = NULLIF($6, ''),
		    last_synced_at = $7,
		    updated_at = now()
		WHERE id = $8
	`, StatusReady, result.LocalPath, result.DefaultBranch,
		result.LatestCommitSHA, result.LatestCommitMsg, result.LatestCommitAuth,
		now, projectID); err != nil {
		h.Logger.Error("update post-clone failed", "err", err, "project", projectID)
		return
	}

	h.Logger.Info("clone complete", "project", projectID, "commit", result.LatestCommitSHA)

	if err := detect.Run(ctx, h.Pool, h.Logger, projectID, result.LocalPath); err != nil {
		h.Logger.Warn("task detection failed", "project", projectID, "err", err)
	}

	stats.Run(ctx, h.Dagger, h.Pool, h.Logger, projectID, result.LocalPath)
}

// lookupAuth returns a stored OAuth token if the URL points at a provider
// the user has connected. Empty string means "no auth" (public clone path).
//
// Stage 15: routes via the multi-provider registry. Host-equality match
// against each registered provider's instance host.
func (h *Handlers) lookupAuth(ctx context.Context, userID uuid.UUID, gitURL string) string {
	if h.OAuth == nil {
		return ""
	}
	return h.OAuth.AccessTokenForCloneURL(ctx, userID, gitURL)
}

func (h *Handlers) markError(projectID uuid.UUID, err error) {
	h.Logger.Error("clone failed", "project", projectID, "err", err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = h.Pool.Exec(ctx,
		`UPDATE projects SET status = $1, updated_at = now() WHERE id = $2`,
		StatusError, projectID)
}
