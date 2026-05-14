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

// ResourcePoint is one day's aggregated resource usage across all completed
// runs of a task.
type ResourcePoint struct {
	Date        string `json:"date"` // "2026-05-14"
	AvgCPUMs    int64  `json:"avg_cpu_ms"`
	MaxCPUMs    int64  `json:"max_cpu_ms"`
	AvgMemBytes int64  `json:"avg_mem_bytes"`
	MaxMemBytes int64  `json:"max_mem_bytes"`
	TotalNetRx  int64  `json:"total_net_rx_bytes"`
	TotalNetTx  int64  `json:"total_net_tx_bytes"`
	RunCount    int    `json:"run_count"`
}

// ResourceSummary wraps daily resource points with overall averages for the
// requested period.
type ResourceSummary struct {
	TaskID      uuid.UUID       `json:"task_id"`
	TaskName    string          `json:"task_name"`
	Period      string          `json:"period"` // "30d"
	Points      []ResourcePoint `json:"points"`
	AvgCPUMs    int64           `json:"avg_cpu_ms"`
	AvgMemBytes int64           `json:"avg_mem_bytes"`
	TotalNetRx  int64           `json:"total_net_rx_bytes"`
	TotalNetTx  int64           `json:"total_net_tx_bytes"`
}

// GetResourceTrends returns daily resource-usage aggregates for a task's
// completed runs. Only runs with status 'succeeded' or 'failed' are
// included.
//
// GET /api/tasks/{id}/resources?days=30
func (h *Handlers) GetResourceTrends(w http.ResponseWriter, r *http.Request) {
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

	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 365 {
			days = n
		}
	}
	since := time.Now().AddDate(0, 0, -days)

	// Fetch task name for the response envelope.
	var taskName string
	if err := h.Pool.QueryRow(r.Context(),
		`SELECT name FROM tasks WHERE id = $1`, taskID).Scan(&taskName); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT to_char(date_trunc('day', started_at), 'YYYY-MM-DD') AS day,
		       COALESCE(AVG(cpu_ms), 0)::BIGINT            AS avg_cpu,
		       COALESCE(MAX(cpu_ms), 0)::BIGINT            AS max_cpu,
		       COALESCE(AVG(mem_peak_bytes), 0)::BIGINT    AS avg_mem,
		       COALESCE(MAX(mem_peak_bytes), 0)::BIGINT    AS max_mem,
		       COALESCE(SUM(net_rx_bytes), 0)::BIGINT      AS total_rx,
		       COALESCE(SUM(net_tx_bytes), 0)::BIGINT      AS total_tx,
		       COUNT(*)                                     AS run_count
		FROM runs
		WHERE task_id = $1
		  AND status IN ('succeeded', 'failed')
		  AND started_at IS NOT NULL
		  AND started_at >= $2
		GROUP BY day
		ORDER BY day
	`, taskID, since)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	points := []ResourcePoint{}
	var totalCPU, totalMem, totalRx, totalTx int64
	var totalRuns int
	for rows.Next() {
		var p ResourcePoint
		if err := rows.Scan(&p.Date, &p.AvgCPUMs, &p.MaxCPUMs, &p.AvgMemBytes,
			&p.MaxMemBytes, &p.TotalNetRx, &p.TotalNetTx, &p.RunCount); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		totalCPU += p.AvgCPUMs * int64(p.RunCount)
		totalMem += p.AvgMemBytes * int64(p.RunCount)
		totalRx += p.TotalNetRx
		totalTx += p.TotalNetTx
		totalRuns += p.RunCount
		points = append(points, p)
	}

	summary := ResourceSummary{
		TaskID:      taskID,
		TaskName:    taskName,
		Period:      strconv.Itoa(days) + "d",
		Points:      points,
		TotalNetRx:  totalRx,
		TotalNetTx:  totalTx,
	}
	if totalRuns > 0 {
		summary.AvgCPUMs = totalCPU / int64(totalRuns)
		summary.AvgMemBytes = totalMem / int64(totalRuns)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}
