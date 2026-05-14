package run

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

// Preset is a saved run configuration for a task.
type Preset struct {
	ID        uuid.UUID         `json:"id"`
	TaskID    uuid.UUID         `json:"task_id"`
	Name      string            `json:"name"`
	EnvVars   map[string]string `json:"env_vars"`
	ExtraArgs []string          `json:"extra_args"`
	CreatedBy uuid.UUID         `json:"created_by"`
	CreatedAt time.Time         `json:"created_at"`
}

// ListPresets returns all presets for a given task.
// GET /api/tasks/{id}/presets
func (h *Handlers) ListPresets(w http.ResponseWriter, r *http.Request) {
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

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, task_id, name, env_vars, extra_args, created_by, created_at
		FROM run_presets
		WHERE task_id = $1
		ORDER BY name
	`, taskID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Preset{}
	for rows.Next() {
		var p Preset
		var rawEnv []byte
		if err := rows.Scan(&p.ID, &p.TaskID, &p.Name, &rawEnv, &p.ExtraArgs, &p.CreatedBy, &p.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		p.EnvVars = decodeEnvVars(rawEnv)
		out = append(out, p)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// CreatePreset creates a new preset for a task.
// POST /api/tasks/{id}/presets
func (h *Handlers) CreatePreset(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var body struct {
		Name      string            `json:"name"`
		EnvVars   map[string]string `json:"env_vars"`
		ExtraArgs []string          `json:"extra_args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" || len(body.Name) > 100 {
		http.Error(w, "name is required and must be 1..100 chars", http.StatusBadRequest)
		return
	}
	if len(body.EnvVars) > 50 {
		http.Error(w, "env_vars has too many entries (max 50)", http.StatusBadRequest)
		return
	}
	if err := validateRunInputs(body.EnvVars, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body.ExtraArgs) > 20 {
		http.Error(w, "extra_args has too many entries (max 20)", http.StatusBadRequest)
		return
	}

	if !userOwnsTask(r.Context(), h.Pool, uid, taskID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	envJSON, _ := json.Marshal(body.EnvVars)
	if body.ExtraArgs == nil {
		body.ExtraArgs = []string{}
	}

	var p Preset
	var rawEnv []byte
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO run_presets (task_id, name, env_vars, extra_args, created_by)
		VALUES ($1, $2, $3::jsonb, $4, $5)
		RETURNING id, task_id, name, env_vars, extra_args, created_by, created_at
	`, taskID, body.Name, string(envJSON), body.ExtraArgs, uid).Scan(
		&p.ID, &p.TaskID, &p.Name, &rawEnv, &p.ExtraArgs, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "run_presets_task_name_idx") {
			http.Error(w, "preset with this name already exists for task", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	p.EnvVars = decodeEnvVars(rawEnv)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(p)
}

// DeletePreset removes a preset. The user must be the creator or the task
// owner.
// DELETE /api/presets/{id}
func (h *Handlers) DeletePreset(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	presetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid preset id", http.StatusBadRequest)
		return
	}

	// Check that the user is either the creator or the task owner.
	var taskID, createdBy uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT task_id, created_by FROM run_presets WHERE id = $1
	`, presetID).Scan(&taskID, &createdBy)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if createdBy != uid && !userOwnsTask(r.Context(), h.Pool, uid, taskID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	tag, err := h.Pool.Exec(r.Context(), `DELETE FROM run_presets WHERE id = $1`, presetID)
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

// userOwnsTask checks that the authenticated user has access to the task
// through workspace membership. Mirrors task.userOwnsTask.
func userOwnsTask(ctx context.Context, pool *pgxpool.Pool, uid, tid uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM tasks t
		JOIN projects p   ON p.id = t.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE t.id = $1 AND m.user_id = $2
	`, tid, uid).Scan(&n)
	return err == nil
}

// decodeEnvVars unmarshals a JSONB column into a map. Returns an empty map
// on any decode error so callers can always iterate safely.
func decodeEnvVars(raw []byte) map[string]string {
	m := map[string]string{}
	if len(raw) == 0 {
		return m
	}
	_ = json.Unmarshal(raw, &m)
	return m
}
