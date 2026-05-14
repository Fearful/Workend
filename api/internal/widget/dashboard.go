// Package widget provides lightweight, read-only aggregate endpoints that
// power dashboard and workspace-list UI widgets. Each handler runs one or a
// few SQL queries and returns a small JSON payload -- no mutations.
package widget

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

// Handlers holds shared dependencies for every widget endpoint.
type Handlers struct {
	Pool      *pgxpool.Pool
	ReposRoot string
	Logger    *slog.Logger
}

// ---------------------------------------------------------------------------
// 1. RunPulse  GET /api/widgets/dashboard/run-pulse
// ---------------------------------------------------------------------------

// RunPulseItem is one actively running or queued run visible to the user.
type RunPulseItem struct {
	RunID       uuid.UUID  `json:"run_id"`
	TaskName    string     `json:"task_name"`
	ProjectName string     `json:"project_name"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at"`
}

func (h *Handlers) RunPulse(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT r.id, t.name, p.name, p.workspace_id, r.status, r.started_at
		FROM runs r
		JOIN tasks t              ON t.id = r.task_id
		JOIN projects p           ON p.id = r.project_id
		JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
		WHERE wm.user_id = $1
		  AND r.status IN ('running','queued')
		ORDER BY r.created_at DESC
		LIMIT 50
	`, uid)
	if err != nil {
		h.Logger.Error("widget.RunPulse query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []RunPulseItem{}
	for rows.Next() {
		var item RunPulseItem
		if err := rows.Scan(&item.RunID, &item.TaskName, &item.ProjectName,
			&item.WorkspaceID, &item.Status, &item.StartedAt); err != nil {
			h.Logger.Error("widget.RunPulse scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, item)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 2. FailureHeatmap  GET /api/widgets/dashboard/failure-heatmap?days=7
// ---------------------------------------------------------------------------

// HeatmapCell is one (day-of-week, hour) bucket of failure counts.
type HeatmapCell struct {
	DayOfWeek int `json:"day_of_week"` // 0=Sunday
	Hour      int `json:"hour"`        // 0-23
	Count     int `json:"count"`
}

func (h *Handlers) FailureHeatmap(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	days := parseDays(r, 7, 30)

	rows, err := h.Pool.Query(r.Context(), `
		SELECT EXTRACT(DOW  FROM r.started_at)::int AS dow,
		       EXTRACT(HOUR FROM r.started_at)::int AS hr,
		       COUNT(*)
		FROM runs r
		JOIN projects p           ON p.id = r.project_id
		JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
		WHERE wm.user_id = $1
		  AND r.status = 'failed'
		  AND r.started_at >= now() - make_interval(days => $2)
		GROUP BY dow, hr
		ORDER BY dow, hr
	`, uid, days)
	if err != nil {
		h.Logger.Error("widget.FailureHeatmap query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []HeatmapCell{}
	for rows.Next() {
		var c HeatmapCell
		if err := rows.Scan(&c.DayOfWeek, &c.Hour, &c.Count); err != nil {
			h.Logger.Error("widget.FailureHeatmap scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, c)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 3. MyQueue  GET /api/widgets/dashboard/my-queue
// ---------------------------------------------------------------------------

// QueueResponse is a unified action list for the current user.
type QueueResponse struct {
	PendingApprovals  []QueueApproval `json:"pending_approvals"`
	ExpiringSandboxes []QueueSandbox  `json:"expiring_sandboxes"`
	UnreadMentions    int             `json:"unread_mentions"`
}

// QueueApproval is a run awaiting the user's approval.
type QueueApproval struct {
	RunID       uuid.UUID `json:"run_id"`
	TaskName    string    `json:"task_name"`
	ProjectName string    `json:"project_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// QueueSandbox is a sandbox expiring within two hours.
type QueueSandbox struct {
	SandboxID   uuid.UUID `json:"sandbox_id"`
	Branch      string    `json:"branch"`
	ExpiresAt   time.Time `json:"expires_at"`
	ProjectName string    `json:"project_name"`
}

func (h *Handlers) MyQueue(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	ctx := r.Context()
	resp := QueueResponse{
		PendingApprovals:  []QueueApproval{},
		ExpiringSandboxes: []QueueSandbox{},
	}

	// Pending approvals in user's workspaces.
	rows, err := h.Pool.Query(ctx, `
		SELECT r.id, t.name, p.name, r.created_at
		FROM runs r
		JOIN tasks t              ON t.id = r.task_id
		JOIN projects p           ON p.id = r.project_id
		JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
		WHERE wm.user_id = $1
		  AND r.status = 'pending_approval'
		ORDER BY r.created_at DESC
		LIMIT 20
	`, uid)
	if err != nil {
		h.Logger.Error("widget.MyQueue approvals query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var a QueueApproval
		if err := rows.Scan(&a.RunID, &a.TaskName, &a.ProjectName, &a.CreatedAt); err != nil {
			h.Logger.Error("widget.MyQueue approvals scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		resp.PendingApprovals = append(resp.PendingApprovals, a)
	}

	// Sandboxes expiring within 2 hours.
	rows2, err := h.Pool.Query(ctx, `
		SELECT s.id, s.branch, s.expires_at, p.name
		FROM sandboxes s
		JOIN projects p ON p.id = s.project_id
		WHERE s.created_by = $1
		  AND s.status = 'running'
		  AND s.expires_at < now() + interval '2 hours'
		ORDER BY s.expires_at
		LIMIT 10
	`, uid)
	if err != nil {
		h.Logger.Error("widget.MyQueue sandboxes query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var s QueueSandbox
		if err := rows2.Scan(&s.SandboxID, &s.Branch, &s.ExpiresAt, &s.ProjectName); err != nil {
			h.Logger.Error("widget.MyQueue sandboxes scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		resp.ExpiringSandboxes = append(resp.ExpiringSandboxes, s)
	}

	// Unread mention count.
	if err := h.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM mentions
		WHERE user_id = $1 AND read_at IS NULL
	`, uid).Scan(&resp.UnreadMentions); err != nil {
		h.Logger.Error("widget.MyQueue mentions query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// 4. SprintVelocity  GET /api/widgets/dashboard/sprint-velocity
// ---------------------------------------------------------------------------

// VelocityResponse compares this week to last week.
type VelocityResponse struct {
	ThisWeek  WeekStats `json:"this_week"`
	LastWeek  WeekStats `json:"last_week"`
	RunsDelta int       `json:"runs_delta"`
	RateDelta float64   `json:"rate_delta"`
}

// WeekStats holds success/failure counts for a calendar week.
type WeekStats struct {
	TotalRuns    int     `json:"total_runs"`
	SuccessCount int     `json:"success_count"`
	FailCount    int     `json:"fail_count"`
	SuccessRate  float64 `json:"success_rate"`
}

func (h *Handlers) SprintVelocity(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	// Single query returning two rows (this week, last week).
	rows, err := h.Pool.Query(r.Context(), `
		WITH bounds AS (
			SELECT date_trunc('week', now()) AS this_start,
			       date_trunc('week', now()) - interval '7 days' AS last_start
		)
		SELECT
			CASE WHEN r.created_at >= b.this_start THEN 'this' ELSE 'last' END AS bucket,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE r.status = 'succeeded') AS ok,
			COUNT(*) FILTER (WHERE r.status = 'failed')    AS fail
		FROM runs r
		JOIN projects p           ON p.id = r.project_id
		JOIN workspace_members wm ON wm.workspace_id = p.workspace_id
		CROSS JOIN bounds b
		WHERE wm.user_id = $1
		  AND r.created_at >= b.last_start
		GROUP BY bucket
	`, uid)
	if err != nil {
		h.Logger.Error("widget.SprintVelocity query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	resp := VelocityResponse{}
	for rows.Next() {
		var bucket string
		var total, ok, fail int
		if err := rows.Scan(&bucket, &total, &ok, &fail); err != nil {
			h.Logger.Error("widget.SprintVelocity scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		ws := WeekStats{TotalRuns: total, SuccessCount: ok, FailCount: fail}
		if total > 0 {
			ws.SuccessRate = float64(ok) / float64(total)
		}
		switch bucket {
		case "this":
			resp.ThisWeek = ws
		case "last":
			resp.LastWeek = ws
		}
	}
	resp.RunsDelta = resp.ThisWeek.TotalRuns - resp.LastWeek.TotalRuns
	resp.RateDelta = resp.ThisWeek.SuccessRate - resp.LastWeek.SuccessRate

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// 5. SandboxStatus  GET /api/widgets/dashboard/sandbox-status
// ---------------------------------------------------------------------------

// SandboxWidget is one active sandbox owned by the current user.
type SandboxWidget struct {
	ID          uuid.UUID `json:"id"`
	ProjectName string    `json:"project_name"`
	Branch      string    `json:"branch"`
	Status      string    `json:"status"`
	URL         string    `json:"url"`
	ExpiresAt   time.Time `json:"expires_at"`
	MinutesLeft int       `json:"minutes_left"`
}

func (h *Handlers) SandboxStatus(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT s.id, p.name, s.branch, s.status, s.url, s.expires_at,
		       GREATEST(0, EXTRACT(EPOCH FROM (s.expires_at - now())) / 60)::int AS minutes_left
		FROM sandboxes s
		JOIN projects p ON p.id = s.project_id
		WHERE s.created_by = $1
		  AND s.status IN ('provisioning','running')
		ORDER BY s.created_at DESC
	`, uid)
	if err != nil {
		h.Logger.Error("widget.SandboxStatus query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []SandboxWidget{}
	for rows.Next() {
		var sb SandboxWidget
		if err := rows.Scan(&sb.ID, &sb.ProjectName, &sb.Branch, &sb.Status,
			&sb.URL, &sb.ExpiresAt, &sb.MinutesLeft); err != nil {
			h.Logger.Error("widget.SandboxStatus scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, sb)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ---------------------------------------------------------------------------
// 6. QuotaMeter  GET /api/widgets/dashboard/quota-meter
// ---------------------------------------------------------------------------

// QuotaWidget is the top-level response for the quota meter widget.
type QuotaWidget struct {
	Workspaces []WorkspaceQuota `json:"workspaces"`
}

// WorkspaceQuota holds lightweight usage proxy metrics per workspace.
type WorkspaceQuota struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	ProjectCount  int       `json:"project_count"`
	RunCount      int       `json:"run_count"`
	QuotaBytes    int64     `json:"quota_bytes"`
	UsagePercent  float64   `json:"usage_percent"`
}

func (h *Handlers) QuotaMeter(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	var quotaBytes int64
	if err := h.Pool.QueryRow(r.Context(),
		`SELECT quota_bytes FROM users WHERE id = $1`, uid).Scan(&quotaBytes); err != nil {
		h.Logger.Error("widget.QuotaMeter user query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT w.id, w.name,
		       (SELECT COUNT(*) FROM projects WHERE workspace_id = w.id) AS project_count,
		       (SELECT COUNT(*) FROM runs r
		        JOIN projects p ON p.id = r.project_id
		        WHERE p.workspace_id = w.id) AS run_count
		FROM workspaces w
		WHERE w.user_id = $1
		ORDER BY w.name
	`, uid)
	if err != nil {
		h.Logger.Error("widget.QuotaMeter workspaces query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	resp := QuotaWidget{Workspaces: []WorkspaceQuota{}}
	var totalRuns int
	for rows.Next() {
		var wq WorkspaceQuota
		if err := rows.Scan(&wq.WorkspaceID, &wq.WorkspaceName,
			&wq.ProjectCount, &wq.RunCount); err != nil {
			h.Logger.Error("widget.QuotaMeter scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		wq.QuotaBytes = quotaBytes
		totalRuns += wq.RunCount
		resp.Workspaces = append(resp.Workspaces, wq)
	}

	// Compute per-workspace usage percent as share of total runs (a
	// lightweight proxy -- actual disk usage requires a background job).
	if totalRuns > 0 {
		for i := range resp.Workspaces {
			resp.Workspaces[i].UsagePercent = math.Round(
				float64(resp.Workspaces[i].RunCount)/float64(totalRuns)*10000) / 100
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// parseDays reads the "days" query param, clamping between 1 and max.
func parseDays(r *http.Request, def, max int) int {
	v := r.URL.Query().Get("days")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	if n > max {
		return max
	}
	return n
}
