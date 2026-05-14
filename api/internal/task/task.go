// Package task owns task (= a runnable command discovered in a project) read
// endpoints. Task creation happens implicitly via the detector after every
// project sync (see internal/detect).
package task

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

var envKeyRegexp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type Task struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	Source           string    `json:"source"`
	Name             string    `json:"name"`
	RawCommand       string    `json:"raw_command"`
	DetectedAt       time.Time `json:"detected_at"`
	TimeoutSeconds   *int      `json:"timeout_seconds"`
	RetryMax         int       `json:"retry_max"`
	RetryBackoffSec  int       `json:"retry_backoff_sec"`
	RequiresApproval bool      `json:"requires_approval"`
	MaxConcurrency   int               `json:"max_concurrency"`
	SupersedePolicy  string            `json:"supersede_policy"`
	NeedsServices    []string          `json:"needs_services"`
	ArtifactPatterns []string          `json:"artifact_patterns"`
	EnvVars          map[string]string `json:"env_vars"`
	BaseImage        string            `json:"base_image"`
	NetworkMode      string            `json:"network_mode"`
	ExposedPorts     []string          `json:"exposed_ports"`
	Quarantined      bool              `json:"quarantined"`
	QuarantinedAt    *time.Time        `json:"quarantined_at"`
	QuarantineReason string            `json:"quarantine_reason"`
}

type Handlers struct {
	Pool *pgxpool.Pool
}

func (h *Handlers) ListByProject(w http.ResponseWriter, r *http.Request) {
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

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, project_id, source, name, raw_command, detected_at, timeout_seconds,
		       retry_max, retry_backoff_sec, requires_approval, max_concurrency, supersede_policy,
		       needs_services, artifact_patterns, env_vars, base_image,
		       network_mode, exposed_ports,
		       quarantined, quarantined_at, quarantine_reason
		FROM tasks
		WHERE project_id = $1
		ORDER BY source, name
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Task{}
	for rows.Next() {
		var t Task
		var rawEnv []byte
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Source, &t.Name, &t.RawCommand, &t.DetectedAt, &t.TimeoutSeconds,
			&t.RetryMax, &t.RetryBackoffSec, &t.RequiresApproval, &t.MaxConcurrency, &t.SupersedePolicy,
			&t.NeedsServices, &t.ArtifactPatterns, &rawEnv, &t.BaseImage,
			&t.NetworkMode, &t.ExposedPorts,
			&t.Quarantined, &t.QuarantinedAt, &t.QuarantineReason); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		t.EnvVars = decodeTaskEnvVars(rawEnv)
		out = append(out, t)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Update applies user-editable fields on a task. All fields are optional;
// only the ones present in the body are updated.
// PATCH /api/tasks/:id
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var body struct {
		TimeoutSeconds   *int               `json:"timeout_seconds"`
		RetryMax         *int               `json:"retry_max"`
		RetryBackoffSec  *int               `json:"retry_backoff_sec"`
		RequiresApproval *bool              `json:"requires_approval"`
		MaxConcurrency   *int               `json:"max_concurrency"`
		SupersedePolicy  *string            `json:"supersede_policy"`
		NeedsServices    *[]string          `json:"needs_services"`
		ArtifactPatterns *[]string          `json:"artifact_patterns"`
		EnvVars          *map[string]string `json:"env_vars"`
		BaseImage        *string            `json:"base_image"`
		NetworkMode      *string            `json:"network_mode"`
		ExposedPorts     *[]string          `json:"exposed_ports"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.TimeoutSeconds != nil {
		v := *body.TimeoutSeconds
		if v < 1 || v > 24*60*60 {
			http.Error(w, "timeout_seconds must be between 1 and 86400", http.StatusBadRequest)
			return
		}
	}
	if body.RetryMax != nil && (*body.RetryMax < 0 || *body.RetryMax > 10) {
		http.Error(w, "retry_max must be 0..10", http.StatusBadRequest)
		return
	}
	if body.RetryBackoffSec != nil && (*body.RetryBackoffSec < 1 || *body.RetryBackoffSec > 3600) {
		http.Error(w, "retry_backoff_sec must be 1..3600", http.StatusBadRequest)
		return
	}
	if body.MaxConcurrency != nil && (*body.MaxConcurrency < 0 || *body.MaxConcurrency > 100) {
		http.Error(w, "max_concurrency must be 0..100", http.StatusBadRequest)
		return
	}
	if body.SupersedePolicy != nil {
		v := *body.SupersedePolicy
		if v != "queue" && v != "cancel-old" && v != "reject" {
			http.Error(w, "supersede_policy must be queue|cancel-old|reject", http.StatusBadRequest)
			return
		}
	}
	if body.NeedsServices != nil {
		if len(*body.NeedsServices) > 32 {
			http.Error(w, "needs_services has too many entries (max 32)", http.StatusBadRequest)
			return
		}
		for _, s := range *body.NeedsServices {
			if s == "" || len(s) > 100 {
				http.Error(w, "needs_services entries must be 1..100 chars", http.StatusBadRequest)
				return
			}
		}
	}
	if body.ArtifactPatterns != nil {
		if len(*body.ArtifactPatterns) > 32 {
			http.Error(w, "artifact_patterns has too many entries (max 32)", http.StatusBadRequest)
			return
		}
		for _, s := range *body.ArtifactPatterns {
			if s == "" || len(s) > 200 {
				http.Error(w, "artifact_patterns entries must be 1..200 chars", http.StatusBadRequest)
				return
			}
		}
	}
	if body.EnvVars != nil {
		if len(*body.EnvVars) > 50 {
			http.Error(w, "env_vars has too many entries (max 50)", http.StatusBadRequest)
			return
		}
		for k, v := range *body.EnvVars {
			if !envKeyRegexp.MatchString(k) {
				http.Error(w, "env_vars key must match ^[A-Za-z_][A-Za-z0-9_]*$", http.StatusBadRequest)
				return
			}
			if len(v) > 4096 {
				http.Error(w, "env_vars value too long (max 4096 chars)", http.StatusBadRequest)
				return
			}
		}
	}
	if body.BaseImage != nil {
		v := *body.BaseImage
		if len(v) > 256 {
			http.Error(w, "base_image must be at most 256 chars", http.StatusBadRequest)
			return
		}
		if strings.ContainsAny(v, " \t\n\r") {
			http.Error(w, "base_image must not contain whitespace", http.StatusBadRequest)
			return
		}
	}
	if body.NetworkMode != nil {
		v := *body.NetworkMode
		if v != "default" && v != "host" && v != "none" {
			http.Error(w, "network_mode must be default|host|none", http.StatusBadRequest)
			return
		}
	}
	if body.ExposedPorts != nil {
		if len(*body.ExposedPorts) > 20 {
			http.Error(w, "exposed_ports has too many entries (max 20)", http.StatusBadRequest)
			return
		}
		for _, p := range *body.ExposedPorts {
			if len(p) > 20 {
				http.Error(w, "exposed_ports entry too long (max 20 chars)", http.StatusBadRequest)
				return
			}
		}
	}

	if !userOwnsTask(r.Context(), h.Pool, uid, tid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Marshal env_vars to JSON string for the COALESCE update. When nil
	// (field absent from body), we pass nil so COALESCE keeps the current
	// value.
	var envJSON *string
	if body.EnvVars != nil {
		b, _ := json.Marshal(*body.EnvVars)
		s := string(b)
		envJSON = &s
	}

	var t Task
	var rawEnv []byte
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE tasks
		SET timeout_seconds    = COALESCE($1, timeout_seconds),
		    retry_max          = COALESCE($2, retry_max),
		    retry_backoff_sec  = COALESCE($3, retry_backoff_sec),
		    requires_approval  = COALESCE($4, requires_approval),
		    max_concurrency    = COALESCE($5, max_concurrency),
		    supersede_policy   = COALESCE($6, supersede_policy),
		    needs_services     = COALESCE($7, needs_services),
		    artifact_patterns  = COALESCE($8, artifact_patterns),
		    env_vars           = COALESCE($9::jsonb, env_vars),
		    base_image         = COALESCE($10, base_image),
		    network_mode       = COALESCE($11, network_mode),
		    exposed_ports      = COALESCE($12, exposed_ports)
		WHERE id = $13
		RETURNING id, project_id, source, name, raw_command, detected_at, timeout_seconds,
		          retry_max, retry_backoff_sec, requires_approval, max_concurrency, supersede_policy,
		          needs_services, artifact_patterns, env_vars, base_image,
		          quarantined, quarantined_at, quarantine_reason
	`, body.TimeoutSeconds, body.RetryMax, body.RetryBackoffSec, body.RequiresApproval,
		body.MaxConcurrency, body.SupersedePolicy, body.NeedsServices, body.ArtifactPatterns,
		envJSON, body.BaseImage, body.NetworkMode, body.ExposedPorts, tid).Scan(
		&t.ID, &t.ProjectID, &t.Source, &t.Name, &t.RawCommand, &t.DetectedAt, &t.TimeoutSeconds,
		&t.RetryMax, &t.RetryBackoffSec, &t.RequiresApproval, &t.MaxConcurrency, &t.SupersedePolicy,
		&t.NeedsServices, &t.ArtifactPatterns, &rawEnv, &t.BaseImage,
		&t.Quarantined, &t.QuarantinedAt, &t.QuarantineReason)
	t.EnvVars = decodeTaskEnvVars(rawEnv)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

// Quarantine marks a task as quarantined with a reason.
// POST /api/tasks/{id}/quarantine
func (h *Handlers) Quarantine(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.Reason = strings.TrimSpace(body.Reason)
	if len(body.Reason) > 500 {
		http.Error(w, "reason max 500 chars", http.StatusBadRequest)
		return
	}

	if !userOwnsTask(r.Context(), h.Pool, uid, tid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var t Task
	var rawEnv []byte
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE tasks
		SET quarantined = true, quarantined_at = now(), quarantine_reason = $1
		WHERE id = $2
		RETURNING id, project_id, source, name, raw_command, detected_at, timeout_seconds,
		          retry_max, retry_backoff_sec, requires_approval, max_concurrency, supersede_policy,
		          needs_services, artifact_patterns, env_vars, base_image,
		          network_mode, exposed_ports,
		          quarantined, quarantined_at, quarantine_reason
	`, body.Reason, tid).Scan(
		&t.ID, &t.ProjectID, &t.Source, &t.Name, &t.RawCommand, &t.DetectedAt, &t.TimeoutSeconds,
		&t.RetryMax, &t.RetryBackoffSec, &t.RequiresApproval, &t.MaxConcurrency, &t.SupersedePolicy,
		&t.NeedsServices, &t.ArtifactPatterns, &rawEnv, &t.BaseImage,
		&t.NetworkMode, &t.ExposedPorts,
		&t.Quarantined, &t.QuarantinedAt, &t.QuarantineReason)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	t.EnvVars = decodeTaskEnvVars(rawEnv)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

// Unquarantine removes quarantine status from a task.
// POST /api/tasks/{id}/unquarantine
func (h *Handlers) Unquarantine(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if !userOwnsTask(r.Context(), h.Pool, uid, tid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var t Task
	var rawEnv []byte
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE tasks
		SET quarantined = false, quarantined_at = NULL, quarantine_reason = ''
		WHERE id = $1
		RETURNING id, project_id, source, name, raw_command, detected_at, timeout_seconds,
		          retry_max, retry_backoff_sec, requires_approval, max_concurrency, supersede_policy,
		          needs_services, artifact_patterns, env_vars, base_image,
		          network_mode, exposed_ports,
		          quarantined, quarantined_at, quarantine_reason
	`, tid).Scan(
		&t.ID, &t.ProjectID, &t.Source, &t.Name, &t.RawCommand, &t.DetectedAt, &t.TimeoutSeconds,
		&t.RetryMax, &t.RetryBackoffSec, &t.RequiresApproval, &t.MaxConcurrency, &t.SupersedePolicy,
		&t.NeedsServices, &t.ArtifactPatterns, &rawEnv, &t.BaseImage,
		&t.NetworkMode, &t.ExposedPorts,
		&t.Quarantined, &t.QuarantinedAt, &t.QuarantineReason)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	t.EnvVars = decodeTaskEnvVars(rawEnv)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func userOwnsProject(ctx context.Context, pool *pgxpool.Pool, uid, pid uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&n)
	return err == nil
}

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

// decodeTaskEnvVars unmarshals a JSONB column into a map. Returns an empty
// map on any decode error so callers can always iterate safely.
func decodeTaskEnvVars(raw []byte) map[string]string {
	m := map[string]string{}
	if len(raw) == 0 {
		return m
	}
	_ = json.Unmarshal(raw, &m)
	return m
}
