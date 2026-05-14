package remoteci

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/oauth"
)

type Handlers struct {
	Pool   *pgxpool.Pool
	OAuth  *oauth.Registry
	Logger *slog.Logger
}

type RemotePipelineRun struct {
	ID             uuid.UUID  `json:"id"`
	ProjectID      uuid.UUID  `json:"project_id"`
	ProviderRunID  string     `json:"provider_run_id"`
	Status         string     `json:"status"`
	Branch         *string    `json:"branch"`
	CommitSHA      *string    `json:"commit_sha"`
	WorkflowName   *string    `json:"workflow_name"`
	HTMLURL        *string    `json:"html_url"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	FetchedAt      time.Time  `json:"fetched_at"`
	CallbackURL    string     `json:"callback_url,omitempty"`
	CallbackSentAt *time.Time `json:"callback_sent_at,omitempty"`
}

// GET /api/projects/{id}/remote-pipelines
func (h *Handlers) ListByProject(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if !userOwnsProject(r.Context(), h.Pool, uid, projectID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Refresh if stale (>5 min since last fetch)
	var lastFetch *time.Time
	_ = h.Pool.QueryRow(r.Context(), `SELECT last_pipeline_fetch_at FROM projects WHERE id = $1`, projectID).Scan(&lastFetch)

	if h.OAuth != nil && (lastFetch == nil || time.Since(*lastFetch) > 5*time.Minute) {
		go func(userID, projID uuid.UUID) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := h.syncRemoteRuns(ctx, userID, projID); err != nil {
				h.Logger.Warn("pipeline sync failed", "project", projID, "err", err)
			}
		}(uid, projectID)
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, project_id, provider_run_id, status, branch, commit_sha,
		       workflow_name, html_url, started_at, finished_at, fetched_at,
		       callback_url, callback_sent_at
		FROM remote_pipeline_runs
		WHERE project_id = $1
		ORDER BY started_at DESC NULLS LAST
		LIMIT 20
	`, projectID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []RemotePipelineRun{}
	for rows.Next() {
		var r RemotePipelineRun
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.ProviderRunID, &r.Status,
			&r.Branch, &r.CommitSHA, &r.WorkflowName, &r.HTMLURL,
			&r.StartedAt, &r.FinishedAt, &r.FetchedAt,
			&r.CallbackURL, &r.CallbackSentAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, r)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// POST /api/projects/{id}/remote-pipelines/trigger
func (h *Handlers) TriggerPipeline(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if !userOwnsProject(r.Context(), h.Pool, uid, projectID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body struct {
		Workflow string            `json:"workflow"`
		Branch   string            `json:"branch"`
		Inputs   map[string]string `json:"inputs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	if h.OAuth == nil {
		http.Error(w, "no oauth providers configured", http.StatusBadRequest)
		return
	}

	var gitURL string
	if err := h.Pool.QueryRow(r.Context(), `SELECT git_url FROM projects WHERE id = $1`, projectID).Scan(&gitURL); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	provider, ok := h.OAuth.ForCloneURL(gitURL)
	if !ok {
		http.Error(w, "no provider for this project", http.StatusBadRequest)
		return
	}
	access := h.OAuth.AccessTokenForCloneURL(r.Context(), uid, gitURL)
	if access == "" {
		http.Error(w, "no oauth connection", http.StatusForbidden)
		return
	}
	fullName := oauth.RepoFullNameFromURL(gitURL)

	result, err := provider.TriggerCIWorkflow(r.Context(), access, fullName, body.Workflow, body.Branch, body.Inputs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *Handlers) syncRemoteRuns(ctx context.Context, userID, projectID uuid.UUID) error {
	if h.OAuth == nil {
		return fmt.Errorf("no oauth providers configured")
	}

	var gitURL string
	if err := h.Pool.QueryRow(ctx, `SELECT git_url FROM projects WHERE id = $1`, projectID).Scan(&gitURL); err != nil {
		return err
	}

	provider, ok := h.OAuth.ForCloneURL(gitURL)
	if !ok {
		return fmt.Errorf("no provider for %s", gitURL)
	}
	access := h.OAuth.AccessTokenForCloneURL(ctx, userID, gitURL)
	if access == "" {
		return fmt.Errorf("no oauth connection")
	}
	fullName := oauth.RepoFullNameFromURL(gitURL)

	runs, err := provider.ListPipelineRuns(ctx, access, fullName, 20)
	if err != nil {
		return fmt.Errorf("list pipeline runs: %w", err)
	}

	for _, run := range runs {
		_, err := h.Pool.Exec(ctx, `
			INSERT INTO remote_pipeline_runs
				(project_id, provider_run_id, status, branch, commit_sha, workflow_name, html_url, started_at, finished_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (project_id, provider_run_id) DO UPDATE SET
				status = EXCLUDED.status,
				finished_at = EXCLUDED.finished_at,
				fetched_at = now()
		`, projectID, run.ProviderRunID, run.Status, run.Branch, run.CommitSHA,
			run.WorkflowName, run.HTMLURL, run.StartedAt, run.FinishedAt)
		if err != nil {
			return fmt.Errorf("upsert run %s: %w", run.ProviderRunID, err)
		}
	}

	_, err = h.Pool.Exec(ctx, `UPDATE projects SET last_pipeline_fetch_at = now() WHERE id = $1`, projectID)
	return err
}

func userOwnsProject(ctx context.Context, pool *pgxpool.Pool, userID, projectID uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE p.id = $1 AND m.user_id = $2
	`, projectID, userID).Scan(&n)
	return err == nil
}
