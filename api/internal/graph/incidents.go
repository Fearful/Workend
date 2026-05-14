package graph

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// --- types ---

// Incident is a time-bounded group of failures across projects.
type Incident struct {
	ID           uuid.UUID     `json:"id"`
	WorkspaceID  uuid.UUID     `json:"workspace_id"`
	Title        string        `json:"title"`
	Severity     string        `json:"severity"`
	Status       string        `json:"status"`
	StartedAt    time.Time     `json:"started_at"`
	ResolvedAt   *time.Time    `json:"resolved_at"`
	Summary      string        `json:"summary"`
	CreatedAt    time.Time     `json:"created_at"`
	AffectedRuns []IncidentRun `json:"affected_runs,omitempty"`
}

// IncidentRun is a run linked to an incident, hydrated with task/project info.
type IncidentRun struct {
	RunID       uuid.UUID  `json:"run_id"`
	TaskName    string     `json:"task_name"`
	ProjectName string     `json:"project_name"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
}

// TimelineEntry represents one hour-bucket in the incident timeline.
type TimelineEntry struct {
	Hour       time.Time     `json:"hour"`
	FailedRuns []TimelineRun `json:"failed_runs"`
	IncidentID *uuid.UUID    `json:"incident_id"`
}

// TimelineRun is a failed run in a timeline entry.
type TimelineRun struct {
	RunID       uuid.UUID  `json:"run_id"`
	TaskName    string     `json:"task_name"`
	ProjectName string     `json:"project_name"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
}

// --- handlers ---

// DetectIncidents finds failure clusters across workspace projects and creates
// incidents for them. Looks at the last 24 hours in 30-minute windows.
//
// POST /api/workspaces/{id}/detect-incidents
func (h *Handlers) DetectIncidents(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Query failed runs in the last 24h grouped by 30-minute windows.
	rows, err := h.Pool.Query(r.Context(), `
		SELECT r.id, r.project_id, p.name AS project_name,
		       date_trunc('hour', r.started_at) +
		         INTERVAL '30 min' * floor(extract(minute FROM r.started_at) / 30) AS window_start
		FROM runs r
		JOIN projects p ON p.id = r.project_id
		WHERE p.workspace_id = $1
		  AND r.status = 'failed'
		  AND r.started_at >= now() - INTERVAL '24 hours'
		ORDER BY window_start
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type windowRun struct {
		runID     uuid.UUID
		projectID uuid.UUID
	}
	type window struct {
		start    time.Time
		runs     []windowRun
		projects map[uuid.UUID]struct{}
	}

	windows := map[time.Time]*window{}
	for rows.Next() {
		var runID, projectID uuid.UUID
		var projectName string
		var windowStart time.Time
		if err := rows.Scan(&runID, &projectID, &projectName, &windowStart); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		win, ok := windows[windowStart]
		if !ok {
			win = &window{
				start:    windowStart,
				projects: map[uuid.UUID]struct{}{},
			}
			windows[windowStart] = win
		}
		win.runs = append(win.runs, windowRun{runID: runID, projectID: projectID})
		win.projects[projectID] = struct{}{}
	}

	// Load existing incidents so we can skip overlapping windows.
	existingRows, err := h.Pool.Query(r.Context(), `
		SELECT started_at, started_at + INTERVAL '30 min' AS window_end
		FROM incidents
		WHERE workspace_id = $1
		  AND started_at >= now() - INTERVAL '24 hours'
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer existingRows.Close()

	type existingWindow struct {
		start time.Time
		end   time.Time
	}
	var existing []existingWindow
	for existingRows.Next() {
		var ew existingWindow
		if err := existingRows.Scan(&ew.start, &ew.end); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		existing = append(existing, ew)
	}

	created := []Incident{}
	for _, win := range windows {
		if len(win.projects) < 2 {
			continue
		}

		// Skip if an existing incident overlaps this window.
		windowEnd := win.start.Add(30 * time.Minute)
		skip := false
		for _, ew := range existing {
			if ew.start.Before(windowEnd) && ew.end.After(win.start) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		severity := "warning"
		if len(win.projects) >= 3 {
			severity = "critical"
		}
		title := fmt.Sprintf("Multiple failures across %d projects at %s",
			len(win.projects), win.start.Format("15:04 UTC"))

		var inc Incident
		err := h.Pool.QueryRow(r.Context(), `
			INSERT INTO incidents (workspace_id, title, severity, status, started_at)
			VALUES ($1, $2, $3, 'open', $4)
			RETURNING id, workspace_id, title, severity, status, started_at, resolved_at, summary, created_at
		`, wsID, title, severity, win.start).Scan(
			&inc.ID, &inc.WorkspaceID, &inc.Title, &inc.Severity,
			&inc.Status, &inc.StartedAt, &inc.ResolvedAt, &inc.Summary, &inc.CreatedAt)
		if err != nil {
			h.Logger.Error("create incident failed", "err", err)
			continue
		}

		for _, wr := range win.runs {
			_, _ = h.Pool.Exec(r.Context(), `
				INSERT INTO incident_runs (incident_id, run_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, inc.ID, wr.runID)
		}

		// Track new incident window to avoid double-creating in this batch.
		existing = append(existing, existingWindow{
			start: win.start,
			end:   windowEnd,
		})
		created = append(created, inc)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(created)
}

// ListIncidents returns incidents for a workspace, filtered by status and time.
//
// GET /api/workspaces/{id}/incidents?status=open&days=7
func (h *Handlers) ListIncidents(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 90 {
			days = n
		}
	}

	statusFilter := r.URL.Query().Get("status")

	args := []any{wsID, days}
	statusClause := ""
	if statusFilter != "" {
		statusClause = " AND i.status = $3"
		args = append(args, statusFilter)
	}

	q := `
		SELECT i.id, i.workspace_id, i.title, i.severity, i.status,
		       i.started_at, i.resolved_at, i.summary, i.created_at
		FROM incidents i
		WHERE i.workspace_id = $1
		  AND i.started_at >= now() - make_interval(days => $2)
	` + statusClause + `
		ORDER BY i.started_at DESC
	`

	rows, err := h.Pool.Query(r.Context(), q, args...)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Incident{}
	for rows.Next() {
		var inc Incident
		if err := rows.Scan(&inc.ID, &inc.WorkspaceID, &inc.Title, &inc.Severity,
			&inc.Status, &inc.StartedAt, &inc.ResolvedAt, &inc.Summary, &inc.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, inc)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GetIncident returns a single incident with its affected runs hydrated.
//
// GET /api/incidents/{id}
func (h *Handlers) GetIncident(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	incID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var inc Incident
	err = h.Pool.QueryRow(r.Context(), `
		SELECT i.id, i.workspace_id, i.title, i.severity, i.status,
		       i.started_at, i.resolved_at, i.summary, i.created_at
		FROM incidents i
		WHERE i.id = $1
	`, incID).Scan(&inc.ID, &inc.WorkspaceID, &inc.Title, &inc.Severity,
		&inc.Status, &inc.StartedAt, &inc.ResolvedAt, &inc.Summary, &inc.CreatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Auth: check workspace membership.
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, inc.WorkspaceID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Hydrate affected runs.
	runRows, err := h.Pool.Query(r.Context(), `
		SELECT r.id, t.name, p.name, r.status, r.started_at, r.finished_at
		FROM incident_runs ir
		JOIN runs r ON r.id = ir.run_id
		JOIN tasks t ON t.id = r.task_id
		JOIN projects p ON p.id = r.project_id
		WHERE ir.incident_id = $1
		ORDER BY r.started_at DESC
	`, incID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer runRows.Close()

	inc.AffectedRuns = []IncidentRun{}
	for runRows.Next() {
		var ar IncidentRun
		if err := runRows.Scan(&ar.RunID, &ar.TaskName, &ar.ProjectName,
			&ar.Status, &ar.StartedAt, &ar.FinishedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		inc.AffectedRuns = append(inc.AffectedRuns, ar)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(inc)
}

// UpdateIncident modifies an incident's mutable fields.
//
// PATCH /api/incidents/{id}
func (h *Handlers) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	incID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Check incident exists and user has access.
	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(),
		`SELECT workspace_id FROM incidents WHERE id = $1`, incID).Scan(&wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body struct {
		Status   *string `json:"status"`
		Summary  *string `json:"summary"`
		Severity *string `json:"severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validate status values.
	if body.Status != nil {
		s := *body.Status
		if s != "open" && s != "investigating" && s != "resolved" {
			http.Error(w, "status must be open|investigating|resolved", http.StatusBadRequest)
			return
		}
	}
	if body.Severity != nil {
		s := *body.Severity
		if s != "info" && s != "warning" && s != "critical" {
			http.Error(w, "severity must be info|warning|critical", http.StatusBadRequest)
			return
		}
	}

	q := `
		UPDATE incidents
		SET status   = COALESCE($1, status),
		    summary  = COALESCE($2, summary),
		    severity = COALESCE($3, severity),
		    resolved_at = CASE WHEN $1 = 'resolved' THEN now() ELSE resolved_at END
		WHERE id = $4
		RETURNING id, workspace_id, title, severity, status,
		          started_at, resolved_at, summary, created_at
	`

	var inc Incident
	err = h.Pool.QueryRow(r.Context(), q,
		body.Status, body.Summary, body.Severity, incID).Scan(
		&inc.ID, &inc.WorkspaceID, &inc.Title, &inc.Severity,
		&inc.Status, &inc.StartedAt, &inc.ResolvedAt, &inc.Summary, &inc.CreatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(inc)
}

// GetTimeline returns a timeline of failed runs grouped by hour with incident
// annotations.
//
// GET /api/workspaces/{id}/incident-timeline?days=7
func (h *Handlers) GetTimeline(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	days := 7
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 90 {
			days = n
		}
	}

	// Fetch all failed runs in the period, grouped by hour.
	runRows, err := h.Pool.Query(r.Context(), `
		SELECT r.id, t.name, p.name, r.started_at, r.finished_at,
		       date_trunc('hour', r.started_at) AS hour_bucket
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		JOIN projects p ON p.id = r.project_id
		WHERE p.workspace_id = $1
		  AND r.status = 'failed'
		  AND r.started_at >= now() - make_interval(days => $2)
		ORDER BY hour_bucket, r.started_at
	`, wsID, days)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer runRows.Close()

	buckets := map[time.Time]*TimelineEntry{}
	bucketOrder := []time.Time{}
	for runRows.Next() {
		var tr TimelineRun
		var hour time.Time
		if err := runRows.Scan(&tr.RunID, &tr.TaskName, &tr.ProjectName,
			&tr.StartedAt, &tr.FinishedAt, &hour); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		entry, ok := buckets[hour]
		if !ok {
			entry = &TimelineEntry{Hour: hour, FailedRuns: []TimelineRun{}}
			buckets[hour] = entry
			bucketOrder = append(bucketOrder, hour)
		}
		entry.FailedRuns = append(entry.FailedRuns, tr)
	}

	// Annotate buckets with incident IDs. An incident covers a bucket if
	// its started_at falls within the same hour.
	incRows, err := h.Pool.Query(r.Context(), `
		SELECT id, started_at
		FROM incidents
		WHERE workspace_id = $1
		  AND started_at >= now() - make_interval(days => $2)
	`, wsID, days)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer incRows.Close()

	for incRows.Next() {
		var incID uuid.UUID
		var startedAt time.Time
		if err := incRows.Scan(&incID, &startedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		hour := startedAt.Truncate(time.Hour)
		if entry, ok := buckets[hour]; ok {
			id := incID // copy for pointer
			entry.IncidentID = &id
		}
	}

	// Build ordered output.
	out := make([]TimelineEntry, 0, len(bucketOrder))
	for _, h := range bucketOrder {
		out = append(out, *buckets[h])
	}

	// Suppress variable name collision with method receiver by using
	// explicit strings.Builder pattern would be overkill; just use w directly.
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
