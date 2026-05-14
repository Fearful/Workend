package run

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
)

// AlertRule defines a user-owned alerting threshold for a specific task.
type AlertRule struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	TaskID          uuid.UUID  `json:"task_id"`
	RuleType        string     `json:"rule_type"` // "consecutive_failures", "duration_increase_pct", "success_rate_below"
	Threshold       int        `json:"threshold"`
	WindowMinutes   int        `json:"window_minutes"`
	Enabled         bool       `json:"enabled"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at"`

	// Hydrated on list for UI convenience.
	TaskName string `json:"task_name,omitempty"`
}

// validRuleTypes is the allow-list of supported alert types.
var validRuleTypes = map[string]bool{
	"consecutive_failures":  true,
	"duration_increase_pct": true,
	"success_rate_below":    true,
}

// ListAlertRules returns all alert rules owned by the authenticated user,
// with the associated task name for display.
//
// GET /api/me/alert-rules
func (h *Handlers) ListAlertRules(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	rows, err := h.Pool.Query(r.Context(), `
		SELECT a.id, a.user_id, a.task_id, a.rule_type, a.threshold,
		       a.window_minutes, a.enabled, a.last_triggered_at, a.created_at,
		       t.name
		FROM alert_rules a
		JOIN tasks t ON t.id = a.task_id
		WHERE a.user_id = $1
		ORDER BY a.created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []AlertRule{}
	for rows.Next() {
		var ar AlertRule
		if err := rows.Scan(
			&ar.ID, &ar.UserID, &ar.TaskID, &ar.RuleType, &ar.Threshold,
			&ar.WindowMinutes, &ar.Enabled, &ar.LastTriggeredAt, &ar.CreatedAt,
			&ar.TaskName,
		); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, ar)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// CreateAlertRule creates a new alerting rule for a task. The user must own
// the task (via project -> workspace -> workspace_members). Unique on
// (user_id, task_id, rule_type).
//
// POST /api/tasks/{id}/alert-rules
func (h *Handlers) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		RuleType      string `json:"rule_type"`
		Threshold     int    `json:"threshold"`
		WindowMinutes int    `json:"window_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	body.RuleType = strings.TrimSpace(body.RuleType)
	if !validRuleTypes[body.RuleType] {
		http.Error(w, "rule_type must be one of: consecutive_failures, duration_increase_pct, success_rate_below", http.StatusBadRequest)
		return
	}
	if body.Threshold < 1 || body.Threshold > 1000 {
		http.Error(w, "threshold must be between 1 and 1000", http.StatusBadRequest)
		return
	}
	if body.WindowMinutes < 5 || body.WindowMinutes > 10080 {
		http.Error(w, "window_minutes must be between 5 and 10080", http.StatusBadRequest)
		return
	}

	var ar AlertRule
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO alert_rules (user_id, task_id, rule_type, threshold, window_minutes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, task_id, rule_type, threshold, window_minutes, enabled, last_triggered_at, created_at
	`, uid, taskID, body.RuleType, body.Threshold, body.WindowMinutes).Scan(
		&ar.ID, &ar.UserID, &ar.TaskID, &ar.RuleType, &ar.Threshold,
		&ar.WindowMinutes, &ar.Enabled, &ar.LastTriggeredAt, &ar.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "alert_rules_user_task_type_idx") {
			http.Error(w, "alert rule already exists for this task and type", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ar)
}

// DeleteAlertRule removes an alert rule. The user_id must match.
//
// DELETE /api/alert-rules/{id}
func (h *Handlers) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	tag, err := h.Pool.Exec(r.Context(),
		`DELETE FROM alert_rules WHERE id = $1 AND user_id = $2`, ruleID, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ToggleAlertRule flips the enabled flag on an alert rule.
//
// POST /api/alert-rules/{id}/toggle
func (h *Handlers) ToggleAlertRule(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var enabled bool
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE alert_rules SET enabled = NOT enabled
		WHERE id = $1 AND user_id = $2
		RETURNING enabled
	`, ruleID, uid).Scan(&enabled)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"enabled": enabled})
}

// CheckAlerts evaluates all enabled alert rules for a given task. Called
// internally after a run completes — not an HTTP handler.
func (h *Handlers) CheckAlerts(ctx context.Context, taskID uuid.UUID) {
	if h.Events == nil {
		return
	}

	rows, err := h.Pool.Query(ctx, `
		SELECT id, user_id, rule_type, threshold, window_minutes
		FROM alert_rules
		WHERE task_id = $1 AND enabled = true
	`, taskID)
	if err != nil {
		h.Logger.Warn("check alerts: load rules failed", "task", taskID, "err", err)
		return
	}
	defer rows.Close()

	type rule struct {
		ID            uuid.UUID
		UserID        uuid.UUID
		RuleType      string
		Threshold     int
		WindowMinutes int
	}
	var rules []rule
	for rows.Next() {
		var rl rule
		if err := rows.Scan(&rl.ID, &rl.UserID, &rl.RuleType, &rl.Threshold, &rl.WindowMinutes); err != nil {
			continue
		}
		rules = append(rules, rl)
	}

	for _, rl := range rules {
		triggered := false

		switch rl.RuleType {
		case "consecutive_failures":
			triggered = h.checkConsecutiveFailures(ctx, taskID, rl.Threshold)
		case "duration_increase_pct":
			triggered = h.checkDurationIncrease(ctx, taskID, rl.Threshold)
		case "success_rate_below":
			triggered = h.checkSuccessRateBelow(ctx, taskID, rl.Threshold, rl.WindowMinutes)
		}

		if triggered {
			h.fireAlertEvent(ctx, rl.ID, rl.UserID, taskID, rl.RuleType, rl.Threshold)
		}
	}
}

// checkConsecutiveFailures returns true when the last N runs (ordered by
// started_at DESC) all have status = 'failed'.
func (h *Handlers) checkConsecutiveFailures(ctx context.Context, taskID uuid.UUID, threshold int) bool {
	rows, err := h.Pool.Query(ctx, `
		SELECT status FROM runs
		WHERE task_id = $1 AND status IN ('succeeded', 'failed')
		ORDER BY started_at DESC
		LIMIT $2
	`, taskID, threshold)
	if err != nil {
		return false
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return false
		}
		if status != "failed" {
			return false
		}
		count++
	}
	return count >= threshold
}

// checkDurationIncrease compares the average duration of the last 5 runs
// against the average of the previous 20 runs. Returns true if the recent
// average exceeds the baseline by threshold percent.
func (h *Handlers) checkDurationIncrease(ctx context.Context, taskID uuid.UUID, thresholdPct int) bool {
	var recentAvg, baselineAvg float64
	err := h.Pool.QueryRow(ctx, `
		WITH recent AS (
			SELECT EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000 AS dur_ms
			FROM runs
			WHERE task_id = $1 AND status IN ('succeeded', 'failed')
			  AND finished_at IS NOT NULL
			ORDER BY started_at DESC
			LIMIT 5
		),
		baseline AS (
			SELECT EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000 AS dur_ms
			FROM runs
			WHERE task_id = $1 AND status IN ('succeeded', 'failed')
			  AND finished_at IS NOT NULL
			ORDER BY started_at DESC
			LIMIT 20 OFFSET 5
		)
		SELECT COALESCE(AVG(r.dur_ms), 0), COALESCE(AVG(b.dur_ms), 0)
		FROM recent r, baseline b
	`, taskID).Scan(&recentAvg, &baselineAvg)
	if err != nil || baselineAvg <= 0 {
		return false
	}
	increase := ((recentAvg - baselineAvg) / baselineAvg) * 100
	return increase >= float64(thresholdPct)
}

// checkSuccessRateBelow returns true when the success rate over the window
// drops below threshold percent.
func (h *Handlers) checkSuccessRateBelow(ctx context.Context, taskID uuid.UUID, thresholdPct, windowMinutes int) bool {
	var total, succeeded int
	err := h.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE status = 'succeeded')::int
		FROM runs
		WHERE task_id = $1
		  AND status IN ('succeeded', 'failed')
		  AND started_at >= now() - make_interval(mins := $2)
	`, taskID, windowMinutes).Scan(&total, &succeeded)
	if err != nil || total == 0 {
		return false
	}
	rate := (float64(succeeded) / float64(total)) * 100
	return rate < float64(thresholdPct)
}

// fireAlertEvent emits an event via the outbound webhook dispatcher and
// updates last_triggered_at on the rule.
func (h *Handlers) fireAlertEvent(ctx context.Context, ruleID, userID, taskID uuid.UUID, ruleType string, threshold int) {
	var taskName string
	_ = h.Pool.QueryRow(ctx, `SELECT name FROM tasks WHERE id = $1`, taskID).Scan(&taskName)

	h.Events.Emit(userID, "alert.triggered", map[string]any{
		"rule_id":   ruleID.String(),
		"task_id":   taskID.String(),
		"task_name": taskName,
		"rule_type": ruleType,
		"threshold": threshold,
	})

	_, _ = h.Pool.Exec(ctx,
		`UPDATE alert_rules SET last_triggered_at = now() WHERE id = $1`, ruleID)
}
