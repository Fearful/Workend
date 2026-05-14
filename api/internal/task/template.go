package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// Template is a shareable task blueprint scoped to a workspace.
type Template struct {
	ID             uuid.UUID         `json:"id"`
	WorkspaceID    uuid.UUID         `json:"workspace_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Source         string            `json:"source"`
	RawCommand     string            `json:"raw_command"`
	BaseImage      string            `json:"base_image"`
	EnvVars        map[string]string `json:"env_vars"`
	TimeoutSeconds *int              `json:"timeout_seconds"`
	RetryMax       int               `json:"retry_max"`
	CreatedBy      uuid.UUID         `json:"created_by"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// ListTemplates returns all templates in a workspace. The caller must be a
// workspace member.
//
// GET /api/workspaces/{id}/task-templates
func (h *Handlers) ListTemplates(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}

	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, workspace_id, name, description, source, raw_command, base_image,
		       env_vars, timeout_seconds, retry_max, created_by, created_at, updated_at
		FROM task_templates
		WHERE workspace_id = $1
		ORDER BY name
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Template{}
	for rows.Next() {
		var t Template
		var rawEnv []byte
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &t.Name, &t.Description, &t.Source,
			&t.RawCommand, &t.BaseImage, &rawEnv, &t.TimeoutSeconds, &t.RetryMax,
			&t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		t.EnvVars = decodeTaskEnvVars(rawEnv)
		out = append(out, t)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// CreateTemplate creates a new task template in a workspace.
//
// POST /api/workspaces/{id}/task-templates
func (h *Handlers) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}

	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body struct {
		Name           string            `json:"name"`
		Description    string            `json:"description"`
		Source         string            `json:"source"`
		RawCommand     string            `json:"raw_command"`
		BaseImage      string            `json:"base_image"`
		EnvVars        map[string]string `json:"env_vars"`
		TimeoutSeconds *int              `json:"timeout_seconds"`
		RetryMax       *int              `json:"retry_max"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Description = strings.TrimSpace(body.Description)
	body.Source = strings.TrimSpace(body.Source)
	body.RawCommand = strings.TrimSpace(body.RawCommand)
	body.BaseImage = strings.TrimSpace(body.BaseImage)

	if body.Name == "" || len(body.Name) > 100 {
		http.Error(w, "name is required (1-100 chars)", http.StatusBadRequest)
		return
	}
	if len(body.Description) > 500 {
		http.Error(w, "description max 500 chars", http.StatusBadRequest)
		return
	}
	if body.Source == "" {
		http.Error(w, "source is required", http.StatusBadRequest)
		return
	}
	if body.RawCommand == "" || len(body.RawCommand) > 2000 {
		http.Error(w, "raw_command is required (max 2000 chars)", http.StatusBadRequest)
		return
	}
	if len(body.BaseImage) > 256 {
		http.Error(w, "base_image max 256 chars", http.StatusBadRequest)
		return
	}
	if len(body.EnvVars) > 50 {
		http.Error(w, "env_vars max 50 entries", http.StatusBadRequest)
		return
	}
	for k, v := range body.EnvVars {
		if !envKeyRegexp.MatchString(k) {
			http.Error(w, "env_vars key must match ^[A-Za-z_][A-Za-z0-9_]*$", http.StatusBadRequest)
			return
		}
		if len(v) > 4096 {
			http.Error(w, "env_vars value too long (max 4096 chars)", http.StatusBadRequest)
			return
		}
	}
	if body.TimeoutSeconds != nil {
		v := *body.TimeoutSeconds
		if v < 1 || v > 86400 {
			http.Error(w, "timeout_seconds must be 1-86400", http.StatusBadRequest)
			return
		}
	}
	retryMax := 0
	if body.RetryMax != nil {
		retryMax = *body.RetryMax
		if retryMax < 0 || retryMax > 10 {
			http.Error(w, "retry_max must be 0-10", http.StatusBadRequest)
			return
		}
	}

	envJSON, _ := json.Marshal(body.EnvVars)
	if body.EnvVars == nil {
		envJSON = []byte("{}")
	}

	var t Template
	var rawEnv []byte
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO task_templates (workspace_id, name, description, source, raw_command,
		                            base_image, env_vars, timeout_seconds, retry_max, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10)
		RETURNING id, workspace_id, name, description, source, raw_command, base_image,
		          env_vars, timeout_seconds, retry_max, created_by, created_at, updated_at
	`, wsID, body.Name, body.Description, body.Source, body.RawCommand,
		body.BaseImage, string(envJSON), body.TimeoutSeconds, retryMax, uid).Scan(
		&t.ID, &t.WorkspaceID, &t.Name, &t.Description, &t.Source, &t.RawCommand,
		&t.BaseImage, &rawEnv, &t.TimeoutSeconds, &t.RetryMax, &t.CreatedBy,
		&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "task_templates_ws_name_idx") {
			http.Error(w, "template name already exists in this workspace", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	t.EnvVars = decodeTaskEnvVars(rawEnv)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

// DeleteTemplate removes a task template. Only the creator or workspace
// owner may delete.
//
// DELETE /api/task-templates/{id}
func (h *Handlers) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tmplID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid template id", http.StatusBadRequest)
		return
	}

	var wsID, createdBy uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT workspace_id, created_by FROM task_templates WHERE id = $1
	`, tmplID).Scan(&wsID, &createdBy)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Creator can always delete; otherwise must be workspace owner.
	if createdBy != uid {
		role, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID)
		if err != nil || role != workspace.RoleOwner {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	if _, err := h.Pool.Exec(r.Context(), `DELETE FROM task_templates WHERE id = $1`, tmplID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ApplyTemplate creates a new task in the given project using a template's
// field values. The user must be a member of both the template's workspace
// and the project's workspace.
//
// POST /api/task-templates/{id}/apply
func (h *Handlers) ApplyTemplate(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tmplID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid template id", http.StatusBadRequest)
		return
	}

	var body struct {
		ProjectID uuid.UUID `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProjectID == uuid.Nil {
		http.Error(w, "project_id is required", http.StatusBadRequest)
		return
	}

	// Load the template.
	var tmpl Template
	var rawEnv []byte
	err = h.Pool.QueryRow(r.Context(), `
		SELECT id, workspace_id, name, description, source, raw_command, base_image,
		       env_vars, timeout_seconds, retry_max, created_by, created_at, updated_at
		FROM task_templates
		WHERE id = $1
	`, tmplID).Scan(&tmpl.ID, &tmpl.WorkspaceID, &tmpl.Name, &tmpl.Description, &tmpl.Source,
		&tmpl.RawCommand, &tmpl.BaseImage, &rawEnv, &tmpl.TimeoutSeconds, &tmpl.RetryMax,
		&tmpl.CreatedBy, &tmpl.CreatedAt, &tmpl.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "template not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	tmpl.EnvVars = decodeTaskEnvVars(rawEnv)

	// User must be member of the template's workspace.
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, tmpl.WorkspaceID); err != nil {
		http.Error(w, "no access to template workspace", http.StatusForbidden)
		return
	}

	// User must be member of the project's workspace.
	var projectWSID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT w.id FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		WHERE p.id = $1
	`, body.ProjectID).Scan(&projectWSID)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, projectWSID); err != nil {
		http.Error(w, "no access to project workspace", http.StatusForbidden)
		return
	}

	envJSON, _ := json.Marshal(tmpl.EnvVars)

	var task Task
	var taskRawEnv []byte
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO tasks (project_id, source, name, raw_command, base_image,
		                   env_vars, timeout_seconds, retry_max)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		RETURNING id, project_id, source, name, raw_command, detected_at, timeout_seconds,
		          retry_max, retry_backoff_sec, requires_approval, max_concurrency, supersede_policy,
		          needs_services, artifact_patterns, env_vars, base_image,
		          network_mode, exposed_ports,
		          quarantined, quarantined_at, quarantine_reason
	`, body.ProjectID, tmpl.Source, tmpl.Name, tmpl.RawCommand, tmpl.BaseImage,
		string(envJSON), tmpl.TimeoutSeconds, tmpl.RetryMax).Scan(
		&task.ID, &task.ProjectID, &task.Source, &task.Name, &task.RawCommand,
		&task.DetectedAt, &task.TimeoutSeconds, &task.RetryMax, &task.RetryBackoffSec,
		&task.RequiresApproval, &task.MaxConcurrency, &task.SupersedePolicy,
		&task.NeedsServices, &task.ArtifactPatterns, &taskRawEnv, &task.BaseImage,
		&task.NetworkMode, &task.ExposedPorts,
		&task.Quarantined, &task.QuarantinedAt, &task.QuarantineReason)
	if err != nil {
		if strings.Contains(err.Error(), "tasks_project_source_name_idx") {
			http.Error(w, "a task with this source and name already exists in the project", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	task.EnvVars = decodeTaskEnvVars(taskRawEnv)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(task)
}
