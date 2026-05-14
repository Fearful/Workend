// Workspace-scoped aggregates: failure rate, slowest tasks, run volume by
// day, top contributors to runs, and a small "fleet health" summary that
// powers the workspace dashboard tab.
package dashboard

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
)

type WorkspaceSummary struct {
	WorkspaceID    uuid.UUID       `json:"workspace_id"`
	Window         int             `json:"window_days"`
	Totals         WorkspaceTotals `json:"totals"`
	Storage        StorageInfo     `json:"storage"`
	SlowestTasks   []SlowTask      `json:"slowest_tasks"`
	FailingTasks   []FailingTask   `json:"failing_tasks"`
	DailyRuns      []DayBucket     `json:"daily_runs"`
	ProjectsHealth []ProjectHealth `json:"projects_health"`
}

type StorageInfo struct {
	QuotaBytes     int64            `json:"quota_bytes"`
	UsedBytes      int64            `json:"used_bytes"`
	Projects       []ProjectStorage `json:"projects"`
	Disk           DiskInfo         `json:"disk"`
}

type ProjectStorage struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Bytes       int64     `json:"bytes"`
}

type DiskInfo struct {
	TotalBytes     uint64 `json:"total_bytes"`
	FreeBytes      uint64 `json:"free_bytes"`
	UsedBytes      uint64 `json:"used_bytes"`
	AvailableBytes uint64 `json:"available_bytes"`
}

type WorkspaceTotals struct {
	Projects      int     `json:"projects"`
	Members       int     `json:"members"`
	RunsLastDays  int     `json:"runs_in_window"`
	FailedLast    int     `json:"failed_in_window"`
	SuccessRate   float64 `json:"success_rate"`
	ActiveNow     int     `json:"active_now"`
}

type SlowTask struct {
	TaskID      uuid.UUID `json:"task_id"`
	TaskName    string    `json:"task_name"`
	TaskSource  string    `json:"task_source"`
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	AvgSeconds  float64   `json:"avg_seconds"`
	Runs        int       `json:"runs"`
}

type FailingTask struct {
	TaskID      uuid.UUID `json:"task_id"`
	TaskName    string    `json:"task_name"`
	TaskSource  string    `json:"task_source"`
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Failures    int       `json:"failures"`
	Total       int       `json:"total"`
	FailureRate float64   `json:"failure_rate"`
}

type DayBucket struct {
	Date      string `json:"date"`
	Total     int    `json:"total"`
	Failed    int    `json:"failed"`
	Succeeded int    `json:"succeeded"`
}

type ProjectHealth struct {
	ProjectID      uuid.UUID  `json:"project_id"`
	ProjectName    string     `json:"project_name"`
	Status         string     `json:"status"`
	LastRunStatus  *string    `json:"last_run_status"`
	LastRunAt      *time.Time `json:"last_run_at"`
	FailureRate    float64    `json:"failure_rate"`
	RunsInWindow   int        `json:"runs_in_window"`
}

// GET /api/workspaces/:id/dashboard?days=30
func (h *Handlers) WorkspaceSummary(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	// Membership check.
	var member bool
	_ = h.Pool.QueryRow(r.Context(), `
		SELECT EXISTS (SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2)
	`, wsID, uid).Scan(&member)
	if !member {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	days := parseDaysParam(r, 30)
	since := time.Now().AddDate(0, 0, -days)

	resp := WorkspaceSummary{
		WorkspaceID:  wsID,
		Window:       days,
		SlowestTasks: []SlowTask{},
		FailingTasks: []FailingTask{},
		DailyRuns:    []DayBucket{},
		ProjectsHealth: []ProjectHealth{},
	}

	// Totals.
	_ = h.Pool.QueryRow(r.Context(), `
		SELECT
			(SELECT COUNT(*) FROM projects WHERE workspace_id = $1),
			(SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1)
	`, wsID).Scan(&resp.Totals.Projects, &resp.Totals.Members)

	_ = h.Pool.QueryRow(r.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE r.created_at >= $2),
			COUNT(*) FILTER (WHERE r.created_at >= $2 AND r.status = 'failed'),
			COUNT(*) FILTER (WHERE r.status IN ('queued','running','pending_approval'))
		FROM runs r
		JOIN projects p ON p.id = r.project_id
		WHERE p.workspace_id = $1
	`, wsID, since).Scan(&resp.Totals.RunsLastDays, &resp.Totals.FailedLast, &resp.Totals.ActiveNow)
	if resp.Totals.RunsLastDays > 0 {
		resp.Totals.SuccessRate = 1 - float64(resp.Totals.FailedLast)/float64(resp.Totals.RunsLastDays)
	}

	// Slowest tasks (avg duration over the window, min 3 runs).
	rows, err := h.Pool.Query(r.Context(), `
		SELECT t.id, t.name, t.source, p.id, p.name,
		       AVG(EXTRACT(EPOCH FROM (r.finished_at - r.started_at))) AS avg_sec,
		       COUNT(*) AS runs
		FROM runs r
		JOIN tasks t    ON t.id = r.task_id
		JOIN projects p ON p.id = r.project_id
		WHERE p.workspace_id = $1
		  AND r.created_at >= $2
		  AND r.started_at IS NOT NULL AND r.finished_at IS NOT NULL
		GROUP BY t.id, t.name, t.source, p.id, p.name
		HAVING COUNT(*) >= 3
		ORDER BY avg_sec DESC NULLS LAST
		LIMIT 5
	`, wsID, since)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st SlowTask
			var avg *float64
			if err := rows.Scan(&st.TaskID, &st.TaskName, &st.TaskSource, &st.ProjectID, &st.ProjectName, &avg, &st.Runs); err == nil {
				if avg != nil {
					st.AvgSeconds = *avg
				}
				resp.SlowestTasks = append(resp.SlowestTasks, st)
			}
		}
	}

	// Failing tasks (highest failure rate over the window, min 3 runs).
	rows2, err := h.Pool.Query(r.Context(), `
		SELECT t.id, t.name, t.source, p.id, p.name,
		       SUM(CASE WHEN r.status = 'failed' THEN 1 ELSE 0 END) AS failures,
		       COUNT(*) AS total
		FROM runs r
		JOIN tasks t    ON t.id = r.task_id
		JOIN projects p ON p.id = r.project_id
		WHERE p.workspace_id = $1
		  AND r.created_at >= $2
		GROUP BY t.id, t.name, t.source, p.id, p.name
		HAVING COUNT(*) >= 3 AND SUM(CASE WHEN r.status='failed' THEN 1 ELSE 0 END) > 0
		ORDER BY (SUM(CASE WHEN r.status='failed' THEN 1.0 ELSE 0 END) / COUNT(*)) DESC, failures DESC
		LIMIT 5
	`, wsID, since)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var ft FailingTask
			if err := rows2.Scan(&ft.TaskID, &ft.TaskName, &ft.TaskSource, &ft.ProjectID, &ft.ProjectName, &ft.Failures, &ft.Total); err == nil {
				if ft.Total > 0 {
					ft.FailureRate = float64(ft.Failures) / float64(ft.Total)
				}
				resp.FailingTasks = append(resp.FailingTasks, ft)
			}
		}
	}

	// Daily run buckets — sparkline data.
	rows3, err := h.Pool.Query(r.Context(), `
		SELECT to_char(date_trunc('day', r.created_at), 'YYYY-MM-DD') AS d,
		       COUNT(*) AS total,
		       SUM(CASE WHEN r.status = 'failed'    THEN 1 ELSE 0 END) AS failed,
		       SUM(CASE WHEN r.status = 'succeeded' THEN 1 ELSE 0 END) AS succeeded
		FROM runs r
		JOIN projects p ON p.id = r.project_id
		WHERE p.workspace_id = $1 AND r.created_at >= $2
		GROUP BY d
		ORDER BY d
	`, wsID, since)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var b DayBucket
			if err := rows3.Scan(&b.Date, &b.Total, &b.Failed, &b.Succeeded); err == nil {
				resp.DailyRuns = append(resp.DailyRuns, b)
			}
		}
	}

	// Per-project health.
	rows4, err := h.Pool.Query(r.Context(), `
		WITH win AS (
			SELECT r.project_id,
			       COUNT(*) AS total,
			       SUM(CASE WHEN r.status = 'failed' THEN 1 ELSE 0 END) AS failed
			FROM runs r
			JOIN projects p ON p.id = r.project_id
			WHERE p.workspace_id = $1 AND r.created_at >= $2
			GROUP BY r.project_id
		),
		last_run AS (
			SELECT DISTINCT ON (r.project_id) r.project_id, r.status, r.created_at
			FROM runs r
			JOIN projects p ON p.id = r.project_id
			WHERE p.workspace_id = $1
			ORDER BY r.project_id, r.created_at DESC
		)
		SELECT p.id, p.name, p.status,
		       lr.status, lr.created_at,
		       COALESCE(w.total, 0),
		       COALESCE(w.failed, 0)
		FROM projects p
		LEFT JOIN win      w  ON w.project_id  = p.id
		LEFT JOIN last_run lr ON lr.project_id = p.id
		WHERE p.workspace_id = $1
		ORDER BY p.created_at DESC
	`, wsID, since)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var ph ProjectHealth
			var failed int
			if err := rows4.Scan(&ph.ProjectID, &ph.ProjectName, &ph.Status, &ph.LastRunStatus, &ph.LastRunAt, &ph.RunsInWindow, &failed); err == nil {
				if ph.RunsInWindow > 0 {
					ph.FailureRate = float64(failed) / float64(ph.RunsInWindow)
				}
				resp.ProjectsHealth = append(resp.ProjectsHealth, ph)
			}
		}
	}

	// Storage: per-project disk sizes + user quota + instance disk.
	if h.ReposRoot != "" {
		var quotaBytes int64
		_ = h.Pool.QueryRow(r.Context(),
			`SELECT quota_bytes FROM users WHERE id = $1`, uid).Scan(&quotaBytes)
		resp.Storage.QuotaBytes = quotaBytes

		wsDir := filepath.Join(h.ReposRoot, wsID.String())
		projRows, err := h.Pool.Query(r.Context(),
			`SELECT id, name FROM projects WHERE workspace_id = $1 ORDER BY name`, wsID)
		if err == nil {
			defer projRows.Close()
			for projRows.Next() {
				var ps ProjectStorage
				if err := projRows.Scan(&ps.ProjectID, &ps.ProjectName); err == nil {
					ps.Bytes = dirSize(filepath.Join(wsDir, ps.ProjectID.String()))
					resp.Storage.UsedBytes += ps.Bytes
					resp.Storage.Projects = append(resp.Storage.Projects, ps)
				}
			}
		}
		if resp.Storage.Projects == nil {
			resp.Storage.Projects = []ProjectStorage{}
		}

		resp.Storage.Disk = diskInfo(h.ReposRoot)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func dirSize(root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

func diskInfo(path string) DiskInfo {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskInfo{}
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	avail := stat.Bavail * uint64(stat.Bsize)
	return DiskInfo{
		TotalBytes:     total,
		FreeBytes:      free,
		UsedBytes:      total - free,
		AvailableBytes: avail,
	}
}

func parseDaysParam(r *http.Request, def int) int {
	q := r.URL.Query().Get("days")
	if q == "" {
		return def
	}
	var n int
	if _, err := fmtSscanf(q, &n); err != nil {
		return def
	}
	if n < 1 {
		return 1
	}
	if n > 365 {
		return 365
	}
	return n
}

// fmtSscanf is a tiny indirection so we don't need to import "fmt" everywhere.
func fmtSscanf(s string, v *int) (int, error) {
	var n, sign int
	sign = 1
	i := 0
	if i < len(s) && s[i] == '-' {
		sign = -1
		i++
	}
	any := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
		any = true
	}
	if !any {
		return 0, errBadInt
	}
	*v = sign * n
	return 1, nil
}

var errBadInt = newErr("bad int")

type strErr string

func (s strErr) Error() string { return string(s) }
func newErr(s string) error    { return strErr(s) }
