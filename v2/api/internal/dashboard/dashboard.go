// Package dashboard returns aggregated per-project health for the current
// user. Single endpoint; aggregation is done in SQL.
package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

type Card struct {
	ProjectID         uuid.UUID  `json:"project_id"`
	ProjectName       string     `json:"project_name"`
	WorkspaceID       uuid.UUID  `json:"workspace_id"`
	WorkspaceName     string     `json:"workspace_name"`
	Status            string     `json:"status"`
	LastSyncedAt      *time.Time `json:"last_synced_at"`
	LastCommitSHA     *string    `json:"last_commit_sha"`
	LastCommitMessage *string    `json:"last_commit_message"`
	TotalLines        *int       `json:"total_lines"`
	TopLanguage       *string    `json:"top_language"`
	TaskCount         int        `json:"task_count"`

	LastRunStatus     *string    `json:"last_run_status"`
	LastRunAt         *time.Time `json:"last_run_at"`
	LastRunTaskName   *string    `json:"last_run_task_name"`

	HasFailingRecentRun bool `json:"has_failing_recent_run"`
}

type Handlers struct {
	Pool *pgxpool.Pool
}

// GET /api/me/dashboard
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		WITH latest_stats AS (
			SELECT DISTINCT ON (project_id)
				project_id, total_lines, languages
			FROM project_stats
			ORDER BY project_id, computed_at DESC
		),
		latest_run AS (
			SELECT DISTINCT ON (r.project_id)
				r.project_id, r.status, r.created_at, t.name AS task_name
			FROM runs r JOIN tasks t ON t.id = r.task_id
			ORDER BY r.project_id, r.created_at DESC
		),
		recent_failed AS (
			SELECT DISTINCT project_id
			FROM runs
			WHERE status = 'failed'
			  AND created_at > now() - interval '7 days'
		),
		task_counts AS (
			SELECT project_id, COUNT(*) AS cnt FROM tasks GROUP BY project_id
		)
		SELECT p.id, p.name, w.id, w.name, p.status, p.last_synced_at,
		       p.last_commit_sha, p.last_commit_message,
		       ls.total_lines, ls.languages,
		       COALESCE(tc.cnt, 0) AS task_count,
		       lr.status, lr.created_at, lr.task_name,
		       (rf.project_id IS NOT NULL) AS has_failing_recent_run
		FROM projects p
		JOIN workspaces w     ON w.id = p.workspace_id
		LEFT JOIN latest_stats ls ON ls.project_id = p.id
		LEFT JOIN latest_run lr   ON lr.project_id = p.id
		LEFT JOIN recent_failed rf ON rf.project_id = p.id
		LEFT JOIN task_counts tc  ON tc.project_id = p.id
		WHERE w.user_id = $1
		ORDER BY w.name, p.name
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Card{}
	for rows.Next() {
		var c Card
		var langsJSON []byte
		if err := rows.Scan(&c.ProjectID, &c.ProjectName, &c.WorkspaceID, &c.WorkspaceName,
			&c.Status, &c.LastSyncedAt, &c.LastCommitSHA, &c.LastCommitMessage,
			&c.TotalLines, &langsJSON, &c.TaskCount,
			&c.LastRunStatus, &c.LastRunAt, &c.LastRunTaskName,
			&c.HasFailingRecentRun); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		c.TopLanguage = topLang(langsJSON)
		out = append(out, c)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// topLang returns the language with the most code lines, or nil if no
// stats present. Uses alphabetical name as tiebreaker for determinism.
func topLang(raw []byte) *string {
	if len(raw) == 0 {
		return nil
	}
	var langs map[string]struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(raw, &langs); err != nil {
		return nil
	}
	var name string
	maxCode := -1
	for k, v := range langs {
		if v.Code > maxCode || (v.Code == maxCode && k < name) {
			maxCode = v.Code
			name = k
		}
	}
	if name == "" {
		return nil
	}
	return &name
}
