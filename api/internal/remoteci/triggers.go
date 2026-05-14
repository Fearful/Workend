package remoteci

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"workend/api/internal/auth"
)

// TriggerRule describes a rule that fires a remote CI pipeline when a local
// event occurs.
type TriggerRule struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	TriggerOn      string    `json:"trigger_on"`
	RemotePipeline string    `json:"remote_pipeline"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
}

// ListTriggerRules returns all trigger rules for a project.
// GET /api/projects/{id}/ci-triggers
func (h *Handlers) ListTriggerRules(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if !userOwnsProject(r.Context(), h.Pool, uid, projectID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, project_id, trigger_on, remote_pipeline, enabled, created_at
		FROM ci_trigger_rules
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []TriggerRule{}
	for rows.Next() {
		var t TriggerRule
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.TriggerOn, &t.RemotePipeline, &t.Enabled, &t.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, t)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// CreateTriggerRule creates a new CI trigger rule for a project.
// POST /api/projects/{id}/ci-triggers
//
//	body: {"trigger_on": "run_success", "remote_pipeline": "deploy.yml", "enabled": true}
func (h *Handlers) CreateTriggerRule(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if !userOwnsProject(r.Context(), h.Pool, uid, projectID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body struct {
		TriggerOn      string `json:"trigger_on"`
		RemotePipeline string `json:"remote_pipeline"`
		Enabled        *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.TriggerOn = strings.TrimSpace(body.TriggerOn)
	body.RemotePipeline = strings.TrimSpace(body.RemotePipeline)

	if !validTriggerOn(body.TriggerOn) {
		http.Error(w, "trigger_on must be run_success, run_failure, or run_complete", http.StatusBadRequest)
		return
	}
	if body.RemotePipeline == "" || len(body.RemotePipeline) > 256 {
		http.Error(w, "remote_pipeline is required and must be at most 256 chars", http.StatusBadRequest)
		return
	}

	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	var t TriggerRule
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO ci_trigger_rules (project_id, trigger_on, remote_pipeline, enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, project_id, trigger_on, remote_pipeline, enabled, created_at
	`, projectID, body.TriggerOn, body.RemotePipeline, enabled).Scan(
		&t.ID, &t.ProjectID, &t.TriggerOn, &t.RemotePipeline, &t.Enabled, &t.CreatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

// DeleteTriggerRule removes a trigger rule.
// DELETE /api/ci-triggers/{id}
func (h *Handlers) DeleteTriggerRule(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if !h.userOwnsTriggerRule(r, uid, ruleID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `DELETE FROM ci_trigger_rules WHERE id = $1`, ruleID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ToggleTriggerRule flips the enabled flag on a trigger rule.
// POST /api/ci-triggers/{id}/toggle
func (h *Handlers) ToggleTriggerRule(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if !h.userOwnsTriggerRule(r, uid, ruleID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var t TriggerRule
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE ci_trigger_rules SET enabled = NOT enabled
		WHERE id = $1
		RETURNING id, project_id, trigger_on, remote_pipeline, enabled, created_at
	`, ruleID).Scan(&t.ID, &t.ProjectID, &t.TriggerOn, &t.RemotePipeline, &t.Enabled, &t.CreatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func (h *Handlers) userOwnsTriggerRule(r *http.Request, userID, ruleID uuid.UUID) bool {
	var n int
	err := h.Pool.QueryRow(r.Context(), `
		SELECT 1 FROM ci_trigger_rules tr
		JOIN projects p          ON p.id = tr.project_id
		JOIN workspaces w        ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE tr.id = $1 AND m.user_id = $2
	`, ruleID, userID).Scan(&n)
	return err == nil && err != pgx.ErrNoRows
}

func validTriggerOn(s string) bool {
	switch s {
	case "run_success", "run_failure", "run_complete":
		return true
	}
	return false
}
