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

// ActivityDay is one day's worth of run counts for a project.
type ActivityDay struct {
	Date      string `json:"date"` // YYYY-MM-DD
	Total     int    `json:"total"`
	Succeeded int    `json:"succeeded"`
	Failed    int    `json:"failed"`
	Cancelled int    `json:"cancelled"`
}

// Activity returns per-day run counts for a project over a rolling window.
// Only days that had at least one run are returned — the frontend fills the
// gaps for the heatmap.
//
// GET /api/projects/:project_id/activity?days=N (default 365, max 365)
func (h *Handlers) Activity(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "project_id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}
	if !userOwnsProject(r.Context(), h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	days := 365
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 365 {
			days = n
		}
	}
	since := time.Now().AddDate(0, 0, -days)

	rows, err := h.Pool.Query(r.Context(), `
		SELECT to_char(date_trunc('day', created_at), 'YYYY-MM-DD') AS day,
		       COUNT(*) AS total,
		       COUNT(*) FILTER (WHERE status = 'succeeded') AS ok,
		       COUNT(*) FILTER (WHERE status = 'failed')    AS bad,
		       COUNT(*) FILTER (WHERE status = 'cancelled') AS canc
		FROM runs
		WHERE project_id = $1 AND created_at >= $2
		GROUP BY day
		ORDER BY day
	`, pid, since)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []ActivityDay{}
	for rows.Next() {
		var d ActivityDay
		if err := rows.Scan(&d.Date, &d.Total, &d.Succeeded, &d.Failed, &d.Cancelled); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, d)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"days":  days,
		"cells": out,
	})
}
