// Package board owns the per-project issue board: a local cache of upstream
// issues organized into columns by label. Reads serve from the cache;
// writes (drag/drop, comment, close) hit both the cache and the upstream
// provider so the change persists across syncs.
package board

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/oauth"
)

type Handlers struct {
	Pool  *pgxpool.Pool
	OAuth *oauth.Registry
}

// Column is a board column, identified by an upstream label name.
type Column struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type Board struct {
	ID           uuid.UUID  `json:"id"`
	ProjectID    uuid.UUID  `json:"project_id"`
	Columns      []Column   `json:"columns"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
	CreatedAt    time.Time  `json:"created_at"`

	// Hydrated by the GET endpoint:
	Issues          []Issue       `json:"issues"`
	AvailableLabels []oauth.Label `json:"available_labels,omitempty"`
}

type Issue struct {
	ID                uuid.UUID  `json:"id"`
	ProviderNumber    int        `json:"provider_number"`
	Title             string     `json:"title"`
	State             string     `json:"state"`
	Labels            []string   `json:"labels"`
	AuthorHandle      string     `json:"author_handle"`
	HTMLURL           string     `json:"html_url"`
	UpstreamUpdatedAt *time.Time `json:"upstream_updated_at"`
}

type IssueDetail struct {
	Issue
	Body     string         `json:"body"`
	Comments []IssueComment `json:"comments"`
}

type IssueComment struct {
	ID           uuid.UUID `json:"id"`
	ProviderID   int64     `json:"provider_id"`
	Body         string    `json:"body"`
	AuthorHandle string    `json:"author_handle"`
	HTMLURL      string    `json:"html_url"`
	CreatedAt    time.Time `json:"created_at"`
}

// Get returns the board with issues. If no board exists yet, returns the
// project's available labels so the UI can show a setup screen.
//
// GET /api/projects/:id/board
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}
	if !userOwnsProject(r.Context(), h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	board, err := h.fetchBoard(r.Context(), pid)
	if errors.Is(err, pgx.ErrNoRows) {
		// No board configured yet: surface the available label set so the
		// UI's setup screen can show checkboxes.
		labels, _ := h.fetchUpstreamLabels(r.Context(), uid, pid)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"configured":       false,
			"available_labels": labels,
		})
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := h.hydrateIssues(r.Context(), board); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"configured": true,
		"board":      board,
	})
}

// Setup creates (or replaces) the board configuration with the chosen
// columns, then triggers an initial sync from the upstream provider.
//
// POST /api/projects/:id/board  body: {"columns": ["todo", "in progress", "done"]}
func (h *Handlers) Setup(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}
	if !userOwnsProject(r.Context(), h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var req struct {
		Columns []string `json:"columns"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	cols := make([]Column, 0, len(req.Columns))
	seen := map[string]struct{}{}
	for i, name := range req.Columns {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		cols = append(cols, Column{Name: name, Position: i})
	}
	if len(cols) == 0 {
		http.Error(w, "at least one column required", http.StatusBadRequest)
		return
	}

	colsJSON, _ := json.Marshal(cols)
	if _, err := h.Pool.Exec(r.Context(), `
		INSERT INTO issue_boards (project_id, columns) VALUES ($1, $2::jsonb)
		ON CONFLICT (project_id) DO UPDATE SET columns = EXCLUDED.columns
	`, pid, string(colsJSON)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Best-effort initial sync. The user can re-trigger if it fails.
	if err := h.syncIssuesNow(r.Context(), uid, pid); err != nil {
		// Don't fail Setup if sync errors out — the empty board is still
		// usable; UI shows a "sync failed" hint and a retry button.
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "sync_error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// Sync re-fetches issues from the upstream provider into the local cache.
// POST /api/projects/:id/board/sync
func (h *Handlers) Sync(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}
	if !userOwnsProject(r.Context(), h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := h.syncIssuesNow(r.Context(), uid, pid); err != nil {
		http.Error(w, "sync failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MoveIssue replaces the issue's labels so it lands in the target column.
// Effect: removes any *configured-column* labels from the issue's existing
// set, adds the target label. Non-column labels (kinds like "bug",
// "enhancement") are preserved.
//
// PATCH /api/issues/:id/move  body: {"to_column": "in progress"}
func (h *Handlers) MoveIssue(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	iid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		ToColumn string `json:"to_column"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.ToColumn = strings.TrimSpace(body.ToColumn)
	if body.ToColumn == "" {
		http.Error(w, "to_column required", http.StatusBadRequest)
		return
	}

	issue, projectID, gitURL, columns, err := h.loadIssueContext(r.Context(), uid, iid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Validate target column.
	colNames := map[string]struct{}{}
	for _, c := range columns {
		colNames[c.Name] = struct{}{}
	}
	if _, ok := colNames[body.ToColumn]; !ok {
		http.Error(w, "to_column is not a configured column", http.StatusBadRequest)
		return
	}

	// Compute the new label set.
	next := make([]string, 0, len(issue.Labels)+1)
	for _, l := range issue.Labels {
		if _, isCol := colNames[l]; isCol {
			continue
		}
		next = append(next, l)
	}
	next = append(next, body.ToColumn)

	provider, _ := h.OAuth.ForCloneURL(gitURL)
	if provider == nil {
		http.Error(w, "no oauth provider for this project", http.StatusBadRequest)
		return
	}
	access := h.OAuth.AccessTokenForCloneURL(r.Context(), uid, gitURL)
	if access == "" {
		http.Error(w, "no oauth connection for this project", http.StatusBadRequest)
		return
	}
	fullName := oauth.RepoFullNameFromURL(gitURL)

	if err := provider.SetIssueLabels(r.Context(), access, fullName, issue.ProviderNumber, next); err != nil {
		http.Error(w, "upstream label update failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	encoded, _ := json.Marshal(next)
	if _, err := h.Pool.Exec(r.Context(), `
		UPDATE issues SET labels = $1::jsonb, fetched_at = now() WHERE id = $2
	`, string(encoded), iid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = projectID // reserved for future audit logging
	w.WriteHeader(http.StatusNoContent)
}

// GetIssue returns one issue with its comments. Comments are pulled fresh
// on each call (cheap; one HTTP round-trip) so users always see the
// latest thread.
//
// GET /api/issues/:id
func (h *Handlers) GetIssue(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	iid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	issue, _, gitURL, _, err := h.loadIssueContext(r.Context(), uid, iid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body string
	_ = h.Pool.QueryRow(r.Context(), `SELECT COALESCE(body, '') FROM issues WHERE id = $1`, iid).Scan(&body)

	detail := IssueDetail{Issue: issue, Body: body}

	// Refresh comments from upstream into the cache, then read them back.
	if h.OAuth != nil {
		if provider, ok := h.OAuth.ForCloneURL(gitURL); ok {
			access := h.OAuth.AccessTokenForCloneURL(r.Context(), uid, gitURL)
			if access != "" {
				fullName := oauth.RepoFullNameFromURL(gitURL)
				if cs, err := provider.ListIssueComments(r.Context(), access, fullName, issue.ProviderNumber); err == nil {
					_ = h.cacheComments(r.Context(), iid, cs)
				}
			}
		}
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, provider_id, body, COALESCE(author_handle, ''), COALESCE(html_url, ''), created_at
		FROM issue_comments WHERE issue_id = $1 ORDER BY created_at
	`, iid)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c IssueComment
			if err := rows.Scan(&c.ID, &c.ProviderID, &c.Body, &c.AuthorHandle, &c.HTMLURL, &c.CreatedAt); err == nil {
				detail.Comments = append(detail.Comments, c)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}

// CreateComment posts a comment upstream and caches it locally.
// POST /api/issues/:id/comments  body: {"body": "..."}
func (h *Handlers) CreateComment(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	iid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.Body = strings.TrimSpace(body.Body)
	if body.Body == "" {
		http.Error(w, "body required", http.StatusBadRequest)
		return
	}

	issue, _, gitURL, _, err := h.loadIssueContext(r.Context(), uid, iid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	provider, _ := h.OAuth.ForCloneURL(gitURL)
	if provider == nil {
		http.Error(w, "no oauth provider", http.StatusBadRequest)
		return
	}
	access := h.OAuth.AccessTokenForCloneURL(r.Context(), uid, gitURL)
	if access == "" {
		http.Error(w, "no oauth connection", http.StatusBadRequest)
		return
	}
	fullName := oauth.RepoFullNameFromURL(gitURL)

	c, err := provider.CreateIssueComment(r.Context(), access, fullName, issue.ProviderNumber, body.Body)
	if err != nil {
		http.Error(w, "upstream comment failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	_ = h.cacheComments(r.Context(), iid, []oauth.IssueComment{*c})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

// CloseIssue closes upstream and updates the cache.
// POST /api/issues/:id/close
func (h *Handlers) CloseIssue(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	iid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	issue, _, gitURL, _, err := h.loadIssueContext(r.Context(), uid, iid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	provider, _ := h.OAuth.ForCloneURL(gitURL)
	if provider == nil {
		http.Error(w, "no oauth provider", http.StatusBadRequest)
		return
	}
	access := h.OAuth.AccessTokenForCloneURL(r.Context(), uid, gitURL)
	if access == "" {
		http.Error(w, "no oauth connection", http.StatusBadRequest)
		return
	}
	fullName := oauth.RepoFullNameFromURL(gitURL)

	if err := provider.CloseIssue(r.Context(), access, fullName, issue.ProviderNumber); err != nil {
		http.Error(w, "upstream close failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	_, _ = h.Pool.Exec(r.Context(), `UPDATE issues SET state = 'closed' WHERE id = $1`, iid)
	w.WriteHeader(http.StatusNoContent)
}

// --- internals ---

func (h *Handlers) fetchBoard(ctx context.Context, projectID uuid.UUID) (*Board, error) {
	var b Board
	var rawCols []byte
	err := h.Pool.QueryRow(ctx, `
		SELECT id, project_id, columns, last_synced_at, created_at
		FROM issue_boards WHERE project_id = $1
	`, projectID).Scan(&b.ID, &b.ProjectID, &rawCols, &b.LastSyncedAt, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	if len(rawCols) > 0 {
		_ = json.Unmarshal(rawCols, &b.Columns)
	}
	sort.Slice(b.Columns, func(i, j int) bool { return b.Columns[i].Position < b.Columns[j].Position })
	return &b, nil
}

func (h *Handlers) hydrateIssues(ctx context.Context, b *Board) error {
	rows, err := h.Pool.Query(ctx, `
		SELECT id, provider_number, title, state, labels,
		       COALESCE(author_handle, ''), html_url, upstream_updated_at
		FROM issues WHERE project_id = $1
		ORDER BY upstream_updated_at DESC NULLS LAST, provider_number DESC
	`, b.ProjectID)
	if err != nil {
		return err
	}
	defer rows.Close()
	out := []Issue{}
	for rows.Next() {
		var i Issue
		var rawLabels []byte
		if err := rows.Scan(&i.ID, &i.ProviderNumber, &i.Title, &i.State, &rawLabels,
			&i.AuthorHandle, &i.HTMLURL, &i.UpstreamUpdatedAt); err != nil {
			return err
		}
		_ = json.Unmarshal(rawLabels, &i.Labels)
		out = append(out, i)
	}
	b.Issues = out
	return nil
}

func (h *Handlers) fetchUpstreamLabels(ctx context.Context, userID, projectID uuid.UUID) ([]oauth.Label, error) {
	if h.OAuth == nil {
		return nil, errors.New("oauth not configured")
	}
	var gitURL string
	if err := h.Pool.QueryRow(ctx, `SELECT git_url FROM projects WHERE id = $1`, projectID).Scan(&gitURL); err != nil {
		return nil, err
	}
	p, ok := h.OAuth.ForCloneURL(gitURL)
	if !ok {
		return nil, errors.New("no provider for project URL")
	}
	access := h.OAuth.AccessTokenForCloneURL(ctx, userID, gitURL)
	if access == "" {
		return nil, errors.New("no oauth connection")
	}
	return p.ListLabels(ctx, access, oauth.RepoFullNameFromURL(gitURL))
}

// syncIssuesNow pulls upstream issues + labels and upserts the local cache.
// Stale local rows for closed-or-deleted issues are NOT removed; instead
// the state column reflects upstream status.
func (h *Handlers) syncIssuesNow(ctx context.Context, userID, projectID uuid.UUID) error {
	if h.OAuth == nil {
		return errors.New("oauth not configured")
	}
	var gitURL string
	if err := h.Pool.QueryRow(ctx, `SELECT git_url FROM projects WHERE id = $1`, projectID).Scan(&gitURL); err != nil {
		return err
	}
	provider, ok := h.OAuth.ForCloneURL(gitURL)
	if !ok {
		return errors.New("no provider for project URL")
	}
	access := h.OAuth.AccessTokenForCloneURL(ctx, userID, gitURL)
	if access == "" {
		return errors.New("no oauth connection")
	}
	fullName := oauth.RepoFullNameFromURL(gitURL)

	issues, err := provider.ListIssues(ctx, access, fullName)
	if err != nil {
		return fmt.Errorf("list issues: %w", err)
	}

	for _, i := range issues {
		labelsJSON, _ := json.Marshal(i.Labels)
		_, err := h.Pool.Exec(ctx, `
			INSERT INTO issues
			    (project_id, provider_number, title, body, state, labels,
			     author_handle, author_url, html_url, upstream_updated_at)
			VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10)
			ON CONFLICT (project_id, provider_number) DO UPDATE
			SET title = EXCLUDED.title,
			    body = EXCLUDED.body,
			    state = EXCLUDED.state,
			    labels = EXCLUDED.labels,
			    author_handle = EXCLUDED.author_handle,
			    author_url = EXCLUDED.author_url,
			    html_url = EXCLUDED.html_url,
			    upstream_updated_at = EXCLUDED.upstream_updated_at,
			    fetched_at = now()
		`, projectID, i.Number, i.Title, i.Body, i.State, string(labelsJSON),
			i.AuthorHandle, i.AuthorURL, i.HTMLURL, i.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert issue #%d: %w", i.Number, err)
		}
	}

	_, _ = h.Pool.Exec(ctx, `UPDATE issue_boards SET last_synced_at = now() WHERE project_id = $1`, projectID)
	return nil
}

// loadIssueContext returns the issue plus the project context needed to
// hit the upstream provider (clone URL, configured columns).
func (h *Handlers) loadIssueContext(ctx context.Context, userID, issueID uuid.UUID) (Issue, uuid.UUID, string, []Column, error) {
	var (
		issue     Issue
		projectID uuid.UUID
		gitURL    string
		rawCols   []byte
		rawLabels []byte
	)
	err := h.Pool.QueryRow(ctx, `
		SELECT i.id, i.provider_number, i.title, i.state, i.labels,
		       COALESCE(i.author_handle, ''), i.html_url, i.upstream_updated_at,
		       p.id, p.git_url, COALESCE(b.columns, '[]'::jsonb)
		FROM issues i
		JOIN projects p   ON p.id = i.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		LEFT JOIN issue_boards b ON b.project_id = p.id
		WHERE i.id = $1 AND m.user_id = $2
	`, issueID, userID).Scan(&issue.ID, &issue.ProviderNumber, &issue.Title, &issue.State, &rawLabels,
		&issue.AuthorHandle, &issue.HTMLURL, &issue.UpstreamUpdatedAt,
		&projectID, &gitURL, &rawCols)
	if err != nil {
		return Issue{}, uuid.Nil, "", nil, err
	}
	_ = json.Unmarshal(rawLabels, &issue.Labels)
	var cols []Column
	_ = json.Unmarshal(rawCols, &cols)
	return issue, projectID, gitURL, cols, nil
}

// cacheComments upserts a comment slice into the local cache so subsequent
// reads don't have to round-trip the upstream API.
func (h *Handlers) cacheComments(ctx context.Context, issueID uuid.UUID, cs []oauth.IssueComment) error {
	for _, c := range cs {
		_, err := h.Pool.Exec(ctx, `
			INSERT INTO issue_comments (issue_id, provider_id, body, author_handle, html_url, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (issue_id, provider_id) DO UPDATE
			SET body = EXCLUDED.body, html_url = EXCLUDED.html_url, fetched_at = now()
		`, issueID, c.ProviderID, c.Body, c.AuthorHandle, c.HTMLURL, c.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// RecentIssue is a lightweight issue with project context for the workspace feed.
type RecentIssue struct {
	ID                uuid.UUID  `json:"id"`
	ProjectID         uuid.UUID  `json:"project_id"`
	ProjectName       string     `json:"project_name"`
	ProviderNumber    int        `json:"provider_number"`
	Title             string     `json:"title"`
	State             string     `json:"state"`
	Labels            []string   `json:"labels"`
	AuthorHandle      string     `json:"author_handle"`
	HTMLURL           string     `json:"html_url"`
	UpstreamUpdatedAt *time.Time `json:"upstream_updated_at"`
}

// GET /api/workspaces/{id}/recent-issues
func (h *Handlers) RecentIssues(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad workspace id", http.StatusBadRequest)
		return
	}

	var memberCheck int
	if err := h.Pool.QueryRow(r.Context(), `
		SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2
	`, wsID, uid).Scan(&memberCheck); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Background-refresh stale projects (last_issues_fetch_at > 15 min ago or null)
	staleRows, _ := h.Pool.Query(r.Context(), `
		SELECT id FROM projects
		WHERE workspace_id = $1 AND (last_issues_fetch_at IS NULL OR last_issues_fetch_at < now() - interval '15 minutes')
	`, wsID)
	if staleRows != nil {
		for staleRows.Next() {
			var pid uuid.UUID
			if err := staleRows.Scan(&pid); err == nil {
				go func(projectID uuid.UUID) {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					_ = h.syncIssuesNow(ctx, uid, projectID)
					_, _ = h.Pool.Exec(ctx, `UPDATE projects SET last_issues_fetch_at = now() WHERE id = $1`, projectID)
				}(pid)
			}
		}
		staleRows.Close()
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT i.id, i.project_id, p.name, i.provider_number, i.title,
		       i.state, i.labels, i.author_handle, i.html_url, i.upstream_updated_at
		FROM issues i
		JOIN projects p ON p.id = i.project_id
		WHERE p.workspace_id = $1
		  AND i.upstream_updated_at > now() - interval '7 days'
		ORDER BY i.upstream_updated_at DESC
		LIMIT 20
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []RecentIssue{}
	for rows.Next() {
		var ri RecentIssue
		var labelsJSON []byte
		if err := rows.Scan(&ri.ID, &ri.ProjectID, &ri.ProjectName, &ri.ProviderNumber,
			&ri.Title, &ri.State, &labelsJSON, &ri.AuthorHandle, &ri.HTMLURL, &ri.UpstreamUpdatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if len(labelsJSON) > 0 {
			_ = json.Unmarshal(labelsJSON, &ri.Labels)
		}
		if ri.Labels == nil {
			ri.Labels = []string{}
		}
		out = append(out, ri)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
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

// _intParam is a tiny helper to silence unused-imports when chi.URLParam
// returns numeric strings.
var _ = strconv.Itoa
