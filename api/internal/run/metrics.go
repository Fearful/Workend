package run

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
)

// TaskMetrics is the aggregate performance summary for a single task over
// a rolling window. Returned by GetTaskMetrics and GetProjectMetrics.
type TaskMetrics struct {
	TaskID        uuid.UUID `json:"task_id"`
	TaskName      string    `json:"task_name"`
	Source        string    `json:"source"`
	TotalRuns     int       `json:"total_runs"`
	SuccessCount  int       `json:"success_count"`
	FailureCount  int       `json:"failure_count"`
	SuccessRate   float64   `json:"success_rate"`
	AvgDurationMs int64     `json:"avg_duration_ms"`
	P50DurationMs int64     `json:"p50_duration_ms"`
	P95DurationMs int64     `json:"p95_duration_ms"`
	MaxDurationMs int64     `json:"max_duration_ms"`
}

// DailyMetric is one row in the daily trend series.
type DailyMetric struct {
	Date          string `json:"date"` // "2026-05-14"
	TotalRuns     int    `json:"total_runs"`
	SuccessCount  int    `json:"success_count"`
	FailureCount  int    `json:"failure_count"`
	AvgDurationMs int64  `json:"avg_duration_ms"`
}

// parseDays reads the "days" query parameter, clamped to [1, 365] with a
// default of 30.
func parseDays(r *http.Request) int {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 365 {
			days = n
		}
	}
	return days
}

// GetTaskMetrics returns the aggregate run metrics for a single task over the
// last N days (default 30, max 365).
//
// GET /api/tasks/{id}/metrics?days=30
func (h *Handlers) GetTaskMetrics(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if !userOwnsTask(r.Context(), h.Pool, uid, taskID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	days := parseDays(r)
	since := time.Now().AddDate(0, 0, -days)

	var m TaskMetrics
	err = h.Pool.QueryRow(r.Context(), `
		SELECT
			t.id,
			t.name,
			t.source,
			COUNT(*)::int                                                             AS total_runs,
			COUNT(*) FILTER (WHERE r.status = 'succeeded')::int                      AS success_count,
			COUNT(*) FILTER (WHERE r.status = 'failed')::int                         AS failure_count,
			CASE WHEN COUNT(*) > 0
				THEN (COUNT(*) FILTER (WHERE r.status = 'succeeded'))::float8 / COUNT(*)
				ELSE 0
			END                                                                       AS success_rate,
			COALESCE(AVG(EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000)::bigint, 0)
			                                                                          AS avg_duration_ms,
			COALESCE((percentile_cont(0.5) WITHIN GROUP (
				ORDER BY EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000
			))::bigint, 0)                                                            AS p50_duration_ms,
			COALESCE((percentile_cont(0.95) WITHIN GROUP (
				ORDER BY EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000
			))::bigint, 0)                                                            AS p95_duration_ms,
			COALESCE(MAX(EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000)::bigint, 0)
			                                                                          AS max_duration_ms
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.task_id = $1
		  AND r.status IN ('succeeded', 'failed')
		  AND r.started_at >= $2
		  AND r.finished_at IS NOT NULL
		GROUP BY t.id, t.name, t.source
	`, taskID, since).Scan(
		&m.TaskID, &m.TaskName, &m.Source,
		&m.TotalRuns, &m.SuccessCount, &m.FailureCount, &m.SuccessRate,
		&m.AvgDurationMs, &m.P50DurationMs, &m.P95DurationMs, &m.MaxDurationMs,
	)
	if err != nil {
		// No rows is valid — task exists but has no runs in the window.
		m.TaskID = taskID
		var taskName, source string
		_ = h.Pool.QueryRow(r.Context(),
			`SELECT name, source FROM tasks WHERE id = $1`, taskID,
		).Scan(&taskName, &source)
		m.TaskName = taskName
		m.Source = source
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(m)
}

// GetTaskTrends returns a daily breakdown of run counts and durations for a
// single task over the last N days.
//
// GET /api/tasks/{id}/metrics/trends?days=30
func (h *Handlers) GetTaskTrends(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if !userOwnsTask(r.Context(), h.Pool, uid, taskID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	days := parseDays(r)
	since := time.Now().AddDate(0, 0, -days)

	rows, err := h.Pool.Query(r.Context(), `
		SELECT
			date_trunc('day', r.started_at)::date::text                              AS date,
			COUNT(*)::int                                                             AS total_runs,
			COUNT(*) FILTER (WHERE r.status = 'succeeded')::int                      AS success_count,
			COUNT(*) FILTER (WHERE r.status = 'failed')::int                         AS failure_count,
			COALESCE(AVG(EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000)::bigint, 0)
			                                                                          AS avg_duration_ms
		FROM runs r
		WHERE r.task_id = $1
		  AND r.status IN ('succeeded', 'failed')
		  AND r.started_at >= $2
		  AND r.finished_at IS NOT NULL
		GROUP BY date_trunc('day', r.started_at)
		ORDER BY date_trunc('day', r.started_at) ASC
	`, taskID, since)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []DailyMetric{}
	for rows.Next() {
		var dm DailyMetric
		if err := rows.Scan(&dm.Date, &dm.TotalRuns, &dm.SuccessCount, &dm.FailureCount, &dm.AvgDurationMs); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, dm)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GetProjectMetrics returns per-task aggregate metrics for every task in a
// project over the last N days.
//
// GET /api/projects/{id}/metrics?days=30
func (h *Handlers) GetProjectMetrics(w http.ResponseWriter, r *http.Request) {
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

	days := parseDays(r)
	since := time.Now().AddDate(0, 0, -days)

	rows, err := h.Pool.Query(r.Context(), `
		SELECT
			t.id,
			t.name,
			t.source,
			COUNT(*)::int                                                             AS total_runs,
			COUNT(*) FILTER (WHERE r.status = 'succeeded')::int                      AS success_count,
			COUNT(*) FILTER (WHERE r.status = 'failed')::int                         AS failure_count,
			CASE WHEN COUNT(*) > 0
				THEN (COUNT(*) FILTER (WHERE r.status = 'succeeded'))::float8 / COUNT(*)
				ELSE 0
			END                                                                       AS success_rate,
			COALESCE(AVG(EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000)::bigint, 0)
			                                                                          AS avg_duration_ms,
			COALESCE((percentile_cont(0.5) WITHIN GROUP (
				ORDER BY EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000
			))::bigint, 0)                                                            AS p50_duration_ms,
			COALESCE((percentile_cont(0.95) WITHIN GROUP (
				ORDER BY EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000
			))::bigint, 0)                                                            AS p95_duration_ms,
			COALESCE(MAX(EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000)::bigint, 0)
			                                                                          AS max_duration_ms
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.project_id = $1
		  AND r.status IN ('succeeded', 'failed')
		  AND r.started_at >= $2
		  AND r.finished_at IS NOT NULL
		GROUP BY t.id, t.name, t.source
		ORDER BY t.name
	`, pid, since)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []TaskMetrics{}
	for rows.Next() {
		var m TaskMetrics
		if err := rows.Scan(
			&m.TaskID, &m.TaskName, &m.Source,
			&m.TotalRuns, &m.SuccessCount, &m.FailureCount, &m.SuccessRate,
			&m.AvgDurationMs, &m.P50DurationMs, &m.P95DurationMs, &m.MaxDurationMs,
		); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, m)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
