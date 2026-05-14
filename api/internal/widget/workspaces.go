package widget

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"workend/api/internal/auth"
)

// ---------------------------------------------------------------------------
// 7. HealthCards  GET /api/widgets/workspaces/health
// ---------------------------------------------------------------------------

// WorkspaceHealth is a per-workspace health summary card.
type WorkspaceHealth struct {
	WorkspaceID    uuid.UUID  `json:"workspace_id"`
	WorkspaceName  string     `json:"workspace_name"`
	MemberCount    int        `json:"member_count"`
	ProjectCount   int        `json:"project_count"`
	ActiveRuns     int        `json:"active_runs"`
	RecentFailRate float64    `json:"recent_fail_rate"`
	TotalRuns7d    int        `json:"total_runs_7d"`
	LastActivityAt *time.Time `json:"last_activity_at"`
}

func (h *Handlers) HealthCards(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT
			w.id, w.name,
			(SELECT COUNT(*) FROM workspace_members WHERE workspace_id = w.id)      AS member_count,
			(SELECT COUNT(*) FROM projects WHERE workspace_id = w.id)               AS project_count,
			(SELECT COUNT(*) FROM runs r
			 JOIN projects p ON p.id = r.project_id
			 WHERE p.workspace_id = w.id AND r.status IN ('running','queued'))       AS active_runs,
			(SELECT COUNT(*) FILTER (WHERE r2.status = 'failed')
			 FROM runs r2
			 JOIN projects p2 ON p2.id = r2.project_id
			 WHERE p2.workspace_id = w.id
			   AND r2.created_at >= now() - interval '7 days')                      AS fail_7d,
			(SELECT COUNT(*)
			 FROM runs r3
			 JOIN projects p3 ON p3.id = r3.project_id
			 WHERE p3.workspace_id = w.id
			   AND r3.created_at >= now() - interval '7 days')                      AS total_7d,
			(SELECT MAX(r4.created_at)
			 FROM runs r4
			 JOIN projects p4 ON p4.id = r4.project_id
			 WHERE p4.workspace_id = w.id)                                          AS last_activity
		FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id
		WHERE wm.user_id = $1
		ORDER BY w.name
	`, uid)
	if err != nil {
		h.Logger.Error("widget.HealthCards query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []WorkspaceHealth{}
	for rows.Next() {
		var wh WorkspaceHealth
		var fail7d, total7d int
		if err := rows.Scan(&wh.WorkspaceID, &wh.WorkspaceName,
			&wh.MemberCount, &wh.ProjectCount, &wh.ActiveRuns,
			&fail7d, &total7d, &wh.LastActivityAt); err != nil {
			h.Logger.Error("widget.HealthCards scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		wh.TotalRuns7d = total7d
		if total7d > 0 {
			wh.RecentFailRate = float64(fail7d) / float64(total7d)
		}
		out = append(out, wh)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 8. Comparison  GET /api/widgets/workspaces/comparison?days=7
// ---------------------------------------------------------------------------

// WorkspaceComparison holds side-by-side metrics for one workspace.
type WorkspaceComparison struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	RunsPerDay    float64   `json:"runs_per_day"`
	SuccessRate   float64   `json:"success_rate"`
	AvgDurationMs int64     `json:"avg_duration_ms"`
	TotalRuns     int       `json:"total_runs"`
}

func (h *Handlers) Comparison(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	days := parseDays(r, 7, 90)

	rows, err := h.Pool.Query(r.Context(), `
		SELECT w.id, w.name,
		       COUNT(r.id)                                                         AS total,
		       COUNT(r.id) FILTER (WHERE r.status = 'succeeded')                   AS ok,
		       AVG(EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000)
		           FILTER (WHERE r.started_at IS NOT NULL AND r.finished_at IS NOT NULL) AS avg_ms
		FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id
		LEFT JOIN projects p      ON p.workspace_id = w.id
		LEFT JOIN runs r          ON r.project_id = p.id
		                          AND r.created_at >= now() - make_interval(days => $2)
		WHERE wm.user_id = $1
		GROUP BY w.id, w.name
		ORDER BY w.name
	`, uid, days)
	if err != nil {
		h.Logger.Error("widget.Comparison query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []WorkspaceComparison{}
	for rows.Next() {
		var wc WorkspaceComparison
		var okCount int
		var avgMs *float64
		if err := rows.Scan(&wc.WorkspaceID, &wc.WorkspaceName,
			&wc.TotalRuns, &okCount, &avgMs); err != nil {
			h.Logger.Error("widget.Comparison scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if wc.TotalRuns > 0 {
			wc.RunsPerDay = float64(wc.TotalRuns) / float64(days)
			wc.SuccessRate = float64(okCount) / float64(wc.TotalRuns)
		}
		if avgMs != nil {
			wc.AvgDurationMs = int64(*avgMs)
		}
		out = append(out, wc)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 9. IncidentBanner  GET /api/widgets/workspaces/incidents
// ---------------------------------------------------------------------------

// WorkspaceIncident summarises open incidents for a workspace.
type WorkspaceIncident struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	OpenIncidents int       `json:"open_incidents"`
	MaxSeverity   string    `json:"max_severity"`
}

func (h *Handlers) IncidentBanner(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT w.id, w.name,
		       COUNT(i.id) AS open_count,
		       COALESCE(
		           MIN(CASE i.severity
		               WHEN 'critical' THEN 1
		               WHEN 'high'     THEN 2
		               WHEN 'medium'   THEN 3
		               WHEN 'low'      THEN 4
		               ELSE 5
		           END), 5
		       ) AS sev_rank
		FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id
		JOIN incidents i          ON i.workspace_id = w.id
		WHERE wm.user_id = $1
		  AND i.status != 'resolved'
		GROUP BY w.id, w.name
		ORDER BY sev_rank, open_count DESC
	`, uid)
	if err != nil {
		h.Logger.Error("widget.IncidentBanner query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []WorkspaceIncident{}
	for rows.Next() {
		var wi WorkspaceIncident
		var sevRank int
		if err := rows.Scan(&wi.WorkspaceID, &wi.WorkspaceName,
			&wi.OpenIncidents, &sevRank); err != nil {
			h.Logger.Error("widget.IncidentBanner scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		wi.MaxSeverity = sevRankToString(sevRank)
		out = append(out, wi)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 10. StorageBreakdown  GET /api/widgets/workspaces/storage
// ---------------------------------------------------------------------------

// StorageInfo holds entity counts as a storage proxy for one workspace.
type StorageInfo struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	ProjectCount  int       `json:"project_count"`
	TaskCount     int       `json:"task_count"`
	RunCount      int       `json:"run_count"`
	ArtifactCount int       `json:"artifact_count"`
}

func (h *Handlers) StorageBreakdown(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT w.id, w.name,
		       (SELECT COUNT(*) FROM projects WHERE workspace_id = w.id) AS projects,
		       (SELECT COUNT(*) FROM tasks t
		        JOIN projects p ON p.id = t.project_id
		        WHERE p.workspace_id = w.id) AS tasks,
		       (SELECT COUNT(*) FROM runs r
		        JOIN projects p ON p.id = r.project_id
		        WHERE p.workspace_id = w.id) AS runs,
		       (SELECT COUNT(*) FROM artifacts a
		        JOIN runs r ON r.id = a.run_id
		        JOIN projects p ON p.id = r.project_id
		        WHERE p.workspace_id = w.id) AS artifacts
		FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id
		WHERE wm.user_id = $1
		ORDER BY w.name
	`, uid)
	if err != nil {
		h.Logger.Error("widget.StorageBreakdown query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []StorageInfo{}
	for rows.Next() {
		var si StorageInfo
		if err := rows.Scan(&si.WorkspaceID, &si.WorkspaceName,
			&si.ProjectCount, &si.TaskCount, &si.RunCount, &si.ArtifactCount); err != nil {
			h.Logger.Error("widget.StorageBreakdown scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, si)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 11. Templates  GET /api/widgets/workspaces/templates
// ---------------------------------------------------------------------------

// TemplateOption aggregates task templates per workspace.
type TemplateOption struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	TemplateCount int       `json:"template_count"`
	TemplateNames []string  `json:"template_names"`
}

func (h *Handlers) Templates(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT w.id, w.name,
		       COUNT(tt.id) AS template_count,
		       COALESCE(
		           (array_agg(tt.name ORDER BY tt.name) FILTER (WHERE tt.name IS NOT NULL))[1:5],
		           '{}'
		       ) AS top_names
		FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id
		LEFT JOIN task_templates tt ON tt.workspace_id = w.id
		WHERE wm.user_id = $1
		GROUP BY w.id, w.name
		HAVING COUNT(tt.id) > 0
		ORDER BY w.name
	`, uid)
	if err != nil {
		h.Logger.Error("widget.Templates query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []TemplateOption{}
	for rows.Next() {
		var to TemplateOption
		if err := rows.Scan(&to.WorkspaceID, &to.WorkspaceName,
			&to.TemplateCount, &to.TemplateNames); err != nil {
			h.Logger.Error("widget.Templates scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, to)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 12. InvitePending  GET /api/widgets/workspaces/pending-members
// ---------------------------------------------------------------------------

// PendingInvite counts outstanding invitations per workspace.
type PendingInvite struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	PendingCount  int       `json:"pending_count"`
}

func (h *Handlers) InvitePending(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT w.id, w.name, COUNT(wi.id) AS pending
		FROM workspaces w
		JOIN workspace_members wm        ON wm.workspace_id = w.id
		JOIN workspace_invitations wi    ON wi.workspace_id = w.id
		WHERE wm.user_id = $1
		  AND wm.role = 'owner'
		  AND wi.status = 'pending'
		  AND wi.expires_at > now()
		GROUP BY w.id, w.name
		ORDER BY w.name
	`, uid)
	if err != nil {
		h.Logger.Error("widget.InvitePending query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []PendingInvite{}
	for rows.Next() {
		var pi PendingInvite
		if err := rows.Scan(&pi.WorkspaceID, &pi.WorkspaceName, &pi.PendingCount); err != nil {
			h.Logger.Error("widget.InvitePending scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, pi)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func sevRankToString(rank int) string {
	switch rank {
	case 1:
		return "critical"
	case 2:
		return "high"
	case 3:
		return "medium"
	case 4:
		return "low"
	default:
		return "unknown"
	}
}
