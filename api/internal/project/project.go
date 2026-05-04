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
	"workend/api/internal/cidetect"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/deps"
	"workend/api/internal/detect"
	"workend/api/internal/oauth"
	"workend/api/internal/patcred"
	"workend/api/internal/quota"
	"workend/api/internal/repo"
	"workend/api/internal/secret"
	"workend/api/internal/secscan"
	"workend/api/internal/sshkey"
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
	Secret    *secret.Box     // nil-safe; required to decrypt user_ssh_keys rows
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
	if !strings.HasPrefix(req.GitURL, "http://") && !strings.HasPrefix(req.GitURL, "https://") && !sshkey.IsSSHURL(req.GitURL) {
		http.Error(w, "git_url must be http(s) or ssh", http.StatusBadRequest)
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

// ListBranches enumerates branches available for the project's repo.
// Prefers OAuth-provider API (richer metadata: protected, default flag); falls
// back to `git ls-remote --heads` via Dagger when no OAuth connection covers
// the URL.
//
// GET /api/projects/:id/branches
func (h *Handlers) ListBranches(w http.ResponseWriter, r *http.Request) {
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

	type out struct {
		Source   string          `json:"source"` // "provider" | "ls-remote"
		Branches []oauth.Branch  `json:"branches"`
	}
	resp := out{Source: "ls-remote", Branches: []oauth.Branch{}}

	if h.OAuth != nil {
		if branches, err := h.OAuth.ListBranchesForCloneURL(r.Context(), uid, p.GitURL); err == nil && branches != nil {
			resp.Source = "provider"
			resp.Branches = branches
		}
	}

	if resp.Source == "ls-remote" {
		token := h.lookupAuth(r.Context(), uid, p.GitURL)
		raw, err := repo.LsRemoteBranches(r.Context(), h.Dagger, p.GitURL, token)
		if err != nil {
			http.Error(w, "list branches: "+err.Error(), http.StatusBadGateway)
			return
		}
		current := ""
		if p.DefaultBranch != nil {
			current = *p.DefaultBranch
		}
		for _, b := range raw {
			resp.Branches = append(resp.Branches, oauth.Branch{
				Name:      b.Name,
				CommitSHA: b.CommitSHA,
				Default:   b.Name == current,
			})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// SwitchBranch updates the project's default_branch and re-syncs against
// the new branch. Triggers a fresh clone in the background.
//
// POST /api/projects/:id/branch  body: {"name": "feature/x"}
func (h *Handlers) SwitchBranch(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `
		UPDATE projects SET default_branch = $1, status = $2, updated_at = now()
		WHERE id = $3
	`, body.Name, StatusCloning, pid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	go h.cloneAsync(uid, pid, p.WorkspaceID, p.GitURL, body.Name)

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"branch": body.Name})
}

// TriggerCI dispatches an upstream CI run for the project.
//
// POST /api/projects/:id/ci-trigger
//   body: {"workflow": "ci.yml", "branch": "main", "inputs": {"k": "v"}}
//
// `workflow` is mandatory for GitHub/Gitea; ignored by GitLab (whose pipeline
// is monolithic). `branch` defaults to the project's stored default branch.
func (h *Handlers) TriggerCI(w http.ResponseWriter, r *http.Request) {
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
	if h.OAuth == nil {
		http.Error(w, "no oauth providers configured", http.StatusBadRequest)
		return
	}

	var body struct {
		Workflow string            `json:"workflow"`
		Branch   string            `json:"branch"`
		Inputs   map[string]string `json:"inputs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.Workflow = strings.TrimSpace(body.Workflow)
	body.Branch = strings.TrimSpace(body.Branch)
	if body.Branch == "" && p.DefaultBranch != nil {
		body.Branch = *p.DefaultBranch
	}
	if body.Branch == "" {
		http.Error(w, "branch required", http.StatusBadRequest)
		return
	}

	provider, ok := h.OAuth.ForCloneURL(p.GitURL)
	if !ok {
		http.Error(w, "no oauth provider for this URL", http.StatusBadRequest)
		return
	}
	access := h.OAuth.AccessTokenForCloneURL(r.Context(), uid, p.GitURL)
	if access == "" {
		http.Error(w, "no oauth connection", http.StatusBadRequest)
		return
	}
	fullName := oauth.RepoFullNameFromURL(p.GitURL)

	res, err := provider.TriggerCIWorkflow(r.Context(), access, fullName, body.Workflow, body.Branch, body.Inputs)
	if err != nil {
		http.Error(w, "upstream dispatch failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(res)
}

// CreatePullRequest opens a PR/MR on the upstream provider. Requires an
// OAuth connection covering the project's git host.
//
// POST /api/projects/:id/pull-requests
//   body: {"source": "...", "target": "...", "title": "...", "body": "..."}
func (h *Handlers) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
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
	if h.OAuth == nil {
		http.Error(w, "no oauth providers configured", http.StatusBadRequest)
		return
	}

	var body struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.Source = strings.TrimSpace(body.Source)
	body.Target = strings.TrimSpace(body.Target)
	body.Title = strings.TrimSpace(body.Title)
	if body.Source == "" || body.Target == "" {
		http.Error(w, "source and target required", http.StatusBadRequest)
		return
	}
	if body.Source == body.Target {
		http.Error(w, "source and target must differ", http.StatusBadRequest)
		return
	}
	if body.Title == "" {
		body.Title = fmt.Sprintf("Merge %s into %s", body.Source, body.Target)
	}

	res, err := h.OAuth.CreatePullRequestForCloneURL(r.Context(), uid, p.GitURL, oauth.PullRequestInput{
		Source: body.Source, Target: body.Target,
		Title: body.Title, Body: body.Body,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
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
	sshKey := h.lookupSSHKey(ctx, userID, gitURL)

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

	result, err := repo.Clone(ctx, h.Dagger, gitURL, branch, authToken, sshKey, dest)
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

	cidetect.PersistCIConfigs(ctx, h.Pool, h.Logger, projectID, result.LocalPath)

	stats.Run(ctx, h.Dagger, h.Pool, h.Logger, projectID, result.LocalPath)

	// Best-effort secret scan; runs in background so the user-visible sync
	// duration isn't dominated by it on large repos.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if err := secscan.Scan(bgCtx, h.Dagger, h.Pool, projectID, result.LocalPath, result.LatestCommitSHA); err != nil {
			h.Logger.Warn("secret scan failed", "project", projectID, "err", err)
		}
	}()

	// Lockfile parse: cheap, runs synchronously so the dependencies tab
	// is fresh when the user clicks through after a sync completes.
	if err := deps.Parse(ctx, h.Pool, projectID, result.LocalPath); err != nil {
		h.Logger.Warn("dep parse failed", "project", projectID, "err", err)
	}
}

// lookupAuth returns a stored token to inject into the HTTPS clone URL.
// Tries OAuth first (admin-configured providers, host-equality match against
// each registered provider's instance host); falls back to a user-managed
// PAT credential keyed on the URL's host. Empty string means "no auth"
// (public clone path).
func (h *Handlers) lookupAuth(ctx context.Context, userID uuid.UUID, gitURL string) string {
	if h.OAuth != nil {
		if tok := h.OAuth.AccessTokenForCloneURL(ctx, userID, gitURL); tok != "" {
			return tok
		}
	}
	return patcred.LookupForCloneURL(ctx, h.Pool, h.Secret, userID, gitURL)
}

// lookupSSHKey returns the user's most-recent SSH key when gitURL looks
// like a git+SSH URL. Returns nil otherwise (HTTPS path).
func (h *Handlers) lookupSSHKey(ctx context.Context, userID uuid.UUID, gitURL string) *repo.SSHKey {
	if !sshkey.IsSSHURL(gitURL) || h.Secret == nil {
		return nil
	}
	priv, _, err := sshkey.LookupAnyKeyForUser(ctx, h.Pool, h.Secret, userID)
	if err != nil || priv == "" {
		return nil
	}
	return &repo.SSHKey{PrivateKey: priv}
}

func (h *Handlers) markError(projectID uuid.UUID, err error) {
	h.Logger.Error("clone failed", "project", projectID, "err", err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = h.Pool.Exec(ctx,
		`UPDATE projects SET status = $1, updated_at = now() WHERE id = $2`,
		StatusError, projectID)
}
