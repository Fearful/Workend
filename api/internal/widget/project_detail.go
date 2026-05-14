package widget

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// ---------------------------------------------------------------------------
// helper: resolve workspace from project and verify membership
// ---------------------------------------------------------------------------

func (h *Handlers) requireProjectMember(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	uid := auth.UserID(r.Context())
	projID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return uuid.Nil, false
	}

	var wsID uuid.UUID
	if err := h.Pool.QueryRow(r.Context(),
		`SELECT workspace_id FROM projects WHERE id = $1`, projID,
	).Scan(&wsID); err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return uuid.Nil, false
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not a workspace member", http.StatusForbidden)
		return uuid.Nil, false
	}
	return projID, true
}

// ---------------------------------------------------------------------------
// 7. ImpactRadar  GET /api/widgets/project/{id}/impact-radar
// ---------------------------------------------------------------------------

type ImpactRadarResponse struct {
	ProjectID     uuid.UUID   `json:"project_id"`
	ProjectName   string      `json:"project_name"`
	RecentCommits int         `json:"recent_commits"`
	AffectedTasks []RadarTask `json:"tasks"`
	UnmappedTasks int         `json:"unmapped_tasks"`
}

type RadarTask struct {
	TaskID        uuid.UUID `json:"task_id"`
	TaskName      string    `json:"task_name"`
	MatchedFiles  int       `json:"matched_files"`
	TotalInputs   int       `json:"total_inputs"`
	LastRunStatus string    `json:"last_status"`
	Patterns      []string  `json:"patterns"`
}

func (h *Handlers) ImpactRadar(w http.ResponseWriter, r *http.Request) {
	projID, ok := h.requireProjectMember(w, r)
	if !ok {
		return
	}

	ctx := r.Context()

	var projName string
	if err := h.Pool.QueryRow(ctx,
		`SELECT name FROM projects WHERE id = $1`, projID,
	).Scan(&projName); err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	resp := ImpactRadarResponse{
		ProjectID:     projID,
		ProjectName:   projName,
		AffectedTasks: []RadarTask{},
	}

	// task_file_inputs may not exist yet (migration pending).
	rows, err := h.Pool.Query(ctx, `
		WITH task_inputs AS (
			SELECT task_id, COUNT(*) AS total_inputs
			FROM task_file_inputs
			WHERE task_id IN (SELECT id FROM tasks WHERE project_id = $1)
			GROUP BY task_id
		),
		latest_run AS (
			SELECT DISTINCT ON (task_id)
				task_id, status
			FROM runs
			WHERE project_id = $1
			ORDER BY task_id, created_at DESC
		)
		SELECT t.id, t.name,
			COALESCE(ti.total_inputs, 0),
			COALESCE(lr.status, 'none')
		FROM tasks t
		LEFT JOIN task_inputs ti ON ti.task_id = t.id
		LEFT JOIN latest_run lr  ON lr.task_id = t.id
		WHERE t.project_id = $1
		  AND ti.task_id IS NOT NULL
		ORDER BY t.name
	`, projID)
	if err != nil {
		// Table may not exist; return graceful empty response.
		h.Logger.Warn("impact radar query failed, table may not exist", "err", err)

		// Still count unmapped tasks from the tasks table directly.
		var unmapped int
		_ = h.Pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM tasks WHERE project_id = $1`, projID,
		).Scan(&unmapped)
		resp.UnmappedTasks = unmapped

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	defer rows.Close()

	mappedSet := map[uuid.UUID]bool{}
	for rows.Next() {
		var rt RadarTask
		if err := rows.Scan(&rt.TaskID, &rt.TaskName, &rt.TotalInputs, &rt.LastRunStatus); err != nil {
			h.Logger.Error("impact radar scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		rt.MatchedFiles = rt.TotalInputs
		rt.Patterns = []string{}
		mappedSet[rt.TaskID] = true
		resp.AffectedTasks = append(resp.AffectedTasks, rt)
	}

	// Count unmapped tasks (tasks with no file-input patterns).
	var totalTasks int
	if err := h.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tasks WHERE project_id = $1`, projID,
	).Scan(&totalTasks); err == nil {
		resp.UnmappedTasks = totalTasks - len(mappedSet)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// 8. DependencyTree  GET /api/widgets/project/{id}/dependency-tree
// ---------------------------------------------------------------------------

type DepTreeResponse struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Upstream    []DepLink `json:"upstream"`
	Downstream  []DepLink `json:"downstream"`
}

type DepLink struct {
	ProjectID   *uuid.UUID `json:"project_id"`
	ProjectName string     `json:"project_name"`
	DepType     string     `json:"dep_type"`
	Reference   string     `json:"reference"`
}

func (h *Handlers) DependencyTree(w http.ResponseWriter, r *http.Request) {
	projID, ok := h.requireProjectMember(w, r)
	if !ok {
		return
	}

	ctx := r.Context()

	var projName string
	if err := h.Pool.QueryRow(ctx,
		`SELECT name FROM projects WHERE id = $1`, projID,
	).Scan(&projName); err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	resp := DepTreeResponse{
		ProjectID:   projID,
		ProjectName: projName,
		Upstream:    []DepLink{},
		Downstream:  []DepLink{},
	}

	// Upstream: projects this depends on (source = this project).
	upRows, err := h.Pool.Query(ctx, `
		SELECT sd.target_project_id, COALESCE(p.name, 'unresolved'),
			sd.dep_type, sd.reference
		FROM service_dependencies sd
		LEFT JOIN projects p ON p.id = sd.target_project_id
		WHERE sd.source_project_id = $1
		ORDER BY sd.dep_type, sd.reference
	`, projID)
	if err != nil {
		// Table may not exist; return empty.
		h.Logger.Warn("dependency tree upstream query failed", "err", err)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	defer upRows.Close()

	for upRows.Next() {
		var dl DepLink
		if err := upRows.Scan(&dl.ProjectID, &dl.ProjectName, &dl.DepType, &dl.Reference); err != nil {
			h.Logger.Error("dependency tree upstream scan failed", "err", err)
			break
		}
		resp.Upstream = append(resp.Upstream, dl)
	}

	// Downstream: projects that depend on this (target = this project).
	downRows, err := h.Pool.Query(ctx, `
		SELECT sd.source_project_id, p.name,
			sd.dep_type, sd.reference
		FROM service_dependencies sd
		JOIN projects p ON p.id = sd.source_project_id
		WHERE sd.target_project_id = $1
		ORDER BY sd.dep_type, sd.reference
	`, projID)
	if err != nil {
		h.Logger.Warn("dependency tree downstream query failed", "err", err)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	defer downRows.Close()

	for downRows.Next() {
		var dl DepLink
		var srcID uuid.UUID
		if err := downRows.Scan(&srcID, &dl.ProjectName, &dl.DepType, &dl.Reference); err != nil {
			h.Logger.Error("dependency tree downstream scan failed", "err", err)
			break
		}
		dl.ProjectID = &srcID
		resp.Downstream = append(resp.Downstream, dl)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// 9. LiveLogPreview  GET /api/widgets/project/{id}/live-run
// ---------------------------------------------------------------------------

type LiveRunResponse struct {
	HasActiveRun bool     `json:"has_active_run"`
	Run          *LiveRun `json:"run,omitempty"`
}

type LiveRun struct {
	RunID     uuid.UUID  `json:"run_id"`
	TaskName  string     `json:"task_name"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at"`
	LogTail   string     `json:"log_tail"`
	LogLines  int        `json:"log_lines"`
}

const logTailLines = 50

func (h *Handlers) LiveLogPreview(w http.ResponseWriter, r *http.Request) {
	projID, ok := h.requireProjectMember(w, r)
	if !ok {
		return
	}

	var lr LiveRun
	var logPath string
	err := h.Pool.QueryRow(r.Context(), `
		SELECT r.id, t.name, r.status, r.started_at, r.log_path
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.project_id = $1 AND r.status = 'running'
		ORDER BY r.created_at DESC
		LIMIT 1
	`, projID).Scan(&lr.RunID, &lr.TaskName, &lr.Status, &lr.StartedAt, &logPath)

	if err == pgx.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LiveRunResponse{HasActiveRun: false})
		return
	}
	if err != nil {
		h.Logger.Error("live run query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Read log tail from disk.
	if logPath != "" {
		data, err := os.ReadFile(logPath)
		if err == nil {
			lines := strings.Split(string(data), "\n")
			lr.LogLines = len(lines)
			if len(lines) > logTailLines {
				lines = lines[len(lines)-logTailLines:]
			}
			lr.LogTail = strings.Join(lines, "\n")
		} else {
			h.Logger.Warn("could not read log file", "path", logPath, "err", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(LiveRunResponse{
		HasActiveRun: true,
		Run:          &lr,
	})
}

// ---------------------------------------------------------------------------
// 10. AlertRulesPanel  GET /api/widgets/project/{id}/alert-rules
// ---------------------------------------------------------------------------

type AlertRuleWidget struct {
	RuleID          uuid.UUID  `json:"rule_id"`
	TaskID          uuid.UUID  `json:"task_id"`
	TaskName        string     `json:"task_name"`
	RuleType        string     `json:"rule_type"`
	Threshold       int        `json:"threshold"`
	Enabled         bool       `json:"enabled"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	OwnerName       string     `json:"owner_name"`
}

func (h *Handlers) AlertRulesPanel(w http.ResponseWriter, r *http.Request) {
	projID, ok := h.requireProjectMember(w, r)
	if !ok {
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT ar.id, ar.task_id, t.name, ar.rule_type,
			ar.threshold, ar.enabled, ar.last_triggered_at,
			u.display_name
		FROM alert_rules ar
		JOIN tasks t ON t.id = ar.task_id
		JOIN users u ON u.id = ar.user_id
		WHERE t.project_id = $1
		ORDER BY t.name, ar.rule_type
	`, projID)
	if err != nil {
		h.Logger.Error("alert rules query failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []AlertRuleWidget{}
	for rows.Next() {
		var ar AlertRuleWidget
		if err := rows.Scan(
			&ar.RuleID, &ar.TaskID, &ar.TaskName,
			&ar.RuleType, &ar.Threshold, &ar.Enabled,
			&ar.LastTriggeredAt, &ar.OwnerName,
		); err != nil {
			h.Logger.Error("alert rules scan failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, ar)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
