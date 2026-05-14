package widget

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// ---------------------------------------------------------------------------
// 1. ProjectGrid  GET /api/widgets/workspace/{id}/project-grid
// ---------------------------------------------------------------------------

type ProjectCard struct {
	ProjectID    uuid.UUID  `json:"project_id"`
	ProjectName  string     `json:"project_name"`
	TaskCount    int        `json:"task_count"`
	LastRunAt    *time.Time `json:"last_run_at"`
	LastStatus   string     `json:"last_status"`
	RecentPasses int        `json:"recent_passes"`
	RecentFails  int        `json:"recent_fails"`
	StatusColor  string     `json:"status_color"`
}

func statusColor(passes, fails int) string {
	total := passes + fails
	if total == 0 {
		return "gray"
	}
	if fails == 0 {
		return "green"
	}
	if fails*2 < total { // fail rate < 50%
		return "yellow"
	}
	return "red"
}

func (h *Handlers) ProjectGrid(w http.ResponseWriter, r *http.Request) {
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
		WITH project_tasks AS (
			SELECT project_id, COUNT(*) AS task_count
			FROM tasks
			WHERE project_id IN (SELECT id FROM projects WHERE workspace_id = $1)
			GROUP BY project_id
		),
		latest_run AS (
			SELECT DISTINCT ON (r.project_id)
				r.project_id, r.status, r.created_at
			FROM runs r
			WHERE r.project_id IN (SELECT id FROM projects WHERE workspace_id = $1)
			ORDER BY r.project_id, r.created_at DESC
		),
		recent_runs AS (
			SELECT project_id, status,
				ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY created_at DESC) AS rn
			FROM runs
			WHERE project_id IN (SELECT id FROM projects WHERE workspace_id = $1)
		),
		recent_counts AS (
			SELECT project_id,
				COUNT(*) FILTER (WHERE status = 'succeeded') AS passes,
				COUNT(*) FILTER (WHERE status = 'failed') AS fails
			FROM recent_runs
			WHERE rn <= 10
			GROUP BY project_id
		)
		SELECT p.id, p.name,
			COALESCE(pt.task_count, 0),
			lr.created_at, COALESCE(lr.status, ''),
			COALESCE(rc.passes, 0), COALESCE(rc.fails, 0)
		FROM projects p
		LEFT JOIN project_tasks pt ON pt.project_id = p.id
		LEFT JOIN latest_run lr    ON lr.project_id = p.id
		LEFT JOIN recent_counts rc ON rc.project_id = p.id
		WHERE p.workspace_id = $1
		ORDER BY p.name
	`, wsID)
	if err != nil {
		h.Logger.Error("project grid query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []ProjectCard{}
	for rows.Next() {
		var c ProjectCard
		if err := rows.Scan(
			&c.ProjectID, &c.ProjectName,
			&c.TaskCount,
			&c.LastRunAt, &c.LastStatus,
			&c.RecentPasses, &c.RecentFails,
		); err != nil {
			h.Logger.Error("project grid scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if c.LastStatus == "" {
			c.LastStatus = "none"
		}
		c.StatusColor = statusColor(c.RecentPasses, c.RecentFails)
		out = append(out, c)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 2. TeamPresence  GET /api/widgets/workspace/{id}/team-presence
// ---------------------------------------------------------------------------

type TeamMember struct {
	UserID       uuid.UUID  `json:"user_id"`
	DisplayName  string     `json:"display_name"`
	Role         string     `json:"role"`
	LastActiveAt *time.Time `json:"last_active_at"`
	ActiveRuns   int        `json:"active_runs"`
	IsActive     bool       `json:"is_active"`
}

func (h *Handlers) TeamPresence(w http.ResponseWriter, r *http.Request) {
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
		SELECT u.id, u.display_name, m.role,
			(
				SELECT MAX(r.created_at)
				FROM runs r
				JOIN tasks t ON t.id = r.task_id
				JOIN projects p ON p.id = t.project_id
				WHERE p.workspace_id = $1
				  AND r.project_id = p.id
			) AS last_active_at,
			(
				SELECT COUNT(*)
				FROM runs r
				JOIN projects p ON p.id = r.project_id
				WHERE p.workspace_id = $1
				  AND r.status = 'running'
			) AS active_runs
		FROM workspace_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = $1
		ORDER BY m.role = 'owner' DESC, u.display_name
	`, wsID)
	if err != nil {
		h.Logger.Error("team presence query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	now := time.Now()
	out := []TeamMember{}
	for rows.Next() {
		var tm TeamMember
		if err := rows.Scan(
			&tm.UserID, &tm.DisplayName, &tm.Role,
			&tm.LastActiveAt, &tm.ActiveRuns,
		); err != nil {
			h.Logger.Error("team presence scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if tm.LastActiveAt != nil {
			tm.IsActive = now.Sub(*tm.LastActiveAt) < 30*time.Minute
		}
		out = append(out, tm)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 3. DependencyGraphWidget  GET /api/widgets/workspace/{id}/dependency-overview
// ---------------------------------------------------------------------------

type DepOverview struct {
	TotalProjects int                `json:"total_projects"`
	TotalDeps     int                `json:"total_deps"`
	UnresolvedDeps int              `json:"unresolved_deps"`
	MostConnected []ConnectedProject `json:"most_connected"`
}

type ConnectedProject struct {
	ProjectID    uuid.UUID `json:"project_id"`
	ProjectName  string    `json:"project_name"`
	InboundDeps  int       `json:"inbound_deps"`
	OutboundDeps int       `json:"outbound_deps"`
}

func (h *Handlers) DependencyOverview(w http.ResponseWriter, r *http.Request) {
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

	ctx := r.Context()

	var totalProjects int
	if err := h.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM projects WHERE workspace_id = $1`, wsID,
	).Scan(&totalProjects); err != nil {
		h.Logger.Error("dep overview project count failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := DepOverview{
		TotalProjects: totalProjects,
		MostConnected: []ConnectedProject{},
	}

	// service_dependencies may not exist yet (migration pending).
	var totalDeps, unresolvedDeps int
	err = h.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE sd.target_project_id IS NULL)
		FROM service_dependencies sd
		JOIN projects p ON p.id = sd.source_project_id
		WHERE p.workspace_id = $1
	`, wsID).Scan(&totalDeps, &unresolvedDeps)
	if err != nil {
		// Table may not exist; return zeroed response.
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	resp.TotalDeps = totalDeps
	resp.UnresolvedDeps = unresolvedDeps

	rows, err := h.Pool.Query(ctx, `
		WITH ws_projects AS (
			SELECT id FROM projects WHERE workspace_id = $1
		),
		outbound AS (
			SELECT source_project_id AS pid, COUNT(*) AS cnt
			FROM service_dependencies
			WHERE source_project_id IN (SELECT id FROM ws_projects)
			GROUP BY source_project_id
		),
		inbound AS (
			SELECT target_project_id AS pid, COUNT(*) AS cnt
			FROM service_dependencies
			WHERE target_project_id IN (SELECT id FROM ws_projects)
			GROUP BY target_project_id
		),
		combined AS (
			SELECT COALESCE(o.pid, i.pid) AS pid,
				COALESCE(i.cnt, 0) AS inbound,
				COALESCE(o.cnt, 0) AS outbound
			FROM outbound o
			FULL OUTER JOIN inbound i ON i.pid = o.pid
		)
		SELECT c.pid, p.name, c.inbound, c.outbound
		FROM combined c
		JOIN projects p ON p.id = c.pid
		ORDER BY (c.inbound + c.outbound) DESC
		LIMIT 5
	`, wsID)
	if err != nil {
		h.Logger.Error("dep overview most connected query failed", "err", err)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cp ConnectedProject
		if err := rows.Scan(&cp.ProjectID, &cp.ProjectName, &cp.InboundDeps, &cp.OutboundDeps); err != nil {
			h.Logger.Error("dep overview scan failed", "err", err)
			break
		}
		resp.MostConnected = append(resp.MostConnected, cp)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// 4. SecretsSummary  GET /api/widgets/workspace/{id}/secrets-summary
// ---------------------------------------------------------------------------

type SecretsSummaryResponse struct {
	WorkspaceID     uuid.UUID  `json:"workspace_id"`
	WorkspaceName   string     `json:"workspace_name"`
	TotalSecrets    int        `json:"total_secrets"`
	OldestUpdatedAt *time.Time `json:"oldest_updated_at"`
	NewestUpdatedAt *time.Time `json:"newest_updated_at"`
	CreatorCount    int        `json:"creator_count"`
}

func (h *Handlers) SecretsSummary(w http.ResponseWriter, r *http.Request) {
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

	var resp SecretsSummaryResponse
	resp.WorkspaceID = wsID

	err = h.Pool.QueryRow(r.Context(), `
		SELECT w.name,
			COUNT(s.id),
			MIN(s.updated_at),
			MAX(s.updated_at),
			COUNT(DISTINCT s.created_by)
		FROM workspaces w
		LEFT JOIN workspace_secrets s ON s.workspace_id = w.id
		WHERE w.id = $1
		GROUP BY w.id
	`, wsID).Scan(
		&resp.WorkspaceName,
		&resp.TotalSecrets,
		&resp.OldestUpdatedAt,
		&resp.NewestUpdatedAt,
		&resp.CreatorCount,
	)
	if err != nil {
		h.Logger.Error("secrets summary query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// 5. PipelineStatusBoard  GET /api/widgets/workspace/{id}/pipeline-board
// ---------------------------------------------------------------------------

type PipelineBoardResponse struct {
	Queued  []PipelineBoardItem `json:"queued"`
	Running []PipelineBoardItem `json:"running"`
	Passed  []PipelineBoardItem `json:"passed"`
	Failed  []PipelineBoardItem `json:"failed"`
}

type PipelineBoardItem struct {
	PipelineRunID uuid.UUID  `json:"pipeline_run_id"`
	PipelineName  string     `json:"pipeline_name"`
	ProjectName   string     `json:"project_name"`
	Status        string     `json:"status"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
}

func (h *Handlers) PipelineBoard(w http.ResponseWriter, r *http.Request) {
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

	resp := PipelineBoardResponse{
		Queued:  []PipelineBoardItem{},
		Running: []PipelineBoardItem{},
		Passed:  []PipelineBoardItem{},
		Failed:  []PipelineBoardItem{},
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT pr.id, p.name, proj.name, pr.status, pr.started_at, pr.finished_at
		FROM pipeline_runs pr
		JOIN pipelines p   ON p.id = pr.pipeline_id
		JOIN projects proj ON proj.id = p.project_id
		WHERE proj.workspace_id = $1
		  AND pr.created_at > now() - interval '7 days'
		ORDER BY pr.created_at DESC
		LIMIT 50
	`, wsID)
	if err != nil {
		// Table may not exist yet; return empty board.
		h.Logger.Warn("pipeline board query failed, returning empty", "err", err)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item PipelineBoardItem
		if err := rows.Scan(
			&item.PipelineRunID, &item.PipelineName, &item.ProjectName,
			&item.Status, &item.StartedAt, &item.FinishedAt,
		); err != nil {
			h.Logger.Error("pipeline board scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		switch item.Status {
		case "queued":
			resp.Queued = append(resp.Queued, item)
		case "running", "partial":
			resp.Running = append(resp.Running, item)
		case "succeeded":
			resp.Passed = append(resp.Passed, item)
		case "failed":
			resp.Failed = append(resp.Failed, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// 6. SandboxPreviewRack  GET /api/widgets/workspace/{id}/sandbox-preview-rack
// ---------------------------------------------------------------------------

type RackItem struct {
	Type       string     `json:"type"`
	ID         uuid.UUID  `json:"id"`
	ProjectName string    `json:"project_name"`
	Branch     string     `json:"branch"`
	Status     string     `json:"status"`
	URL        string     `json:"url"`
	CreatedBy  string     `json:"created_by"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	AutoDeploy *bool      `json:"auto_deploy,omitempty"`
}

func (h *Handlers) SandboxPreviewRack(w http.ResponseWriter, r *http.Request) {
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

	ctx := r.Context()
	out := []RackItem{}

	// Sandboxes
	sRows, err := h.Pool.Query(ctx, `
		SELECT s.id, p.name, s.branch, s.status, s.url, u.display_name,
			s.expires_at, s.created_at
		FROM sandboxes s
		JOIN projects p ON p.id = s.project_id
		JOIN users u    ON u.id = s.created_by
		WHERE s.workspace_id = $1
		  AND s.status IN ('provisioning', 'running')
		ORDER BY s.created_at DESC
	`, wsID)
	if err != nil {
		h.Logger.Error("sandbox rack query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer sRows.Close()

	for sRows.Next() {
		var item RackItem
		var createdAt time.Time
		if err := sRows.Scan(
			&item.ID, &item.ProjectName, &item.Branch, &item.Status,
			&item.URL, &item.CreatedBy, &item.ExpiresAt, &createdAt,
		); err != nil {
			h.Logger.Error("sandbox rack scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		item.Type = "sandbox"
		out = append(out, item)
	}

	// Previews
	pRows, err := h.Pool.Query(ctx, `
		SELECT pv.id, p.name, pv.branch, pv.status, pv.url, u.display_name,
			pv.auto_deploy, pv.created_at
		FROM previews pv
		JOIN projects p ON p.id = pv.project_id
		JOIN users u    ON u.id = pv.created_by
		WHERE pv.workspace_id = $1
		  AND pv.status IN ('building', 'live')
		ORDER BY pv.created_at DESC
	`, wsID)
	if err != nil {
		h.Logger.Error("preview rack query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer pRows.Close()

	for pRows.Next() {
		var item RackItem
		var createdAt time.Time
		var autoDeploy bool
		if err := pRows.Scan(
			&item.ID, &item.ProjectName, &item.Branch, &item.Status,
			&item.URL, &item.CreatedBy, &autoDeploy, &createdAt,
		); err != nil {
			h.Logger.Error("preview rack scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		item.Type = "preview"
		item.AutoDeploy = &autoDeploy
		out = append(out, item)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
