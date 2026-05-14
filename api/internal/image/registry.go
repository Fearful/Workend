package image

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
)

// RegistryConfig represents a project's container registry push configuration.
type RegistryConfig struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"project_id"`
	RegistryURL string    `json:"registry_url"`
	Repository  string    `json:"repository"`
	Username    string    `json:"username"`
	TagPattern  string    `json:"tag_pattern"`
	AutoPush    bool      `json:"auto_push"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetRegistryConfig returns the registry configuration for a project.
// GET /api/projects/{id}/registry
func (h *Handlers) GetRegistryConfig(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if !userOwnsProject(r, h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var rc RegistryConfig
	err = h.Pool.QueryRow(r.Context(), `
		SELECT id, project_id, registry_url, repository, username, tag_pattern, auto_push, created_at
		FROM registry_configs
		WHERE project_id = $1
	`, pid).Scan(&rc.ID, &rc.ProjectID, &rc.RegistryURL, &rc.Repository, &rc.Username, &rc.TagPattern, &rc.AutoPush, &rc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rc)
}

type setRegistryReq struct {
	RegistryURL string  `json:"registry_url"`
	Repository  string  `json:"repository"`
	Username    string  `json:"username"`
	Password    *string `json:"password"`
	TagPattern  string  `json:"tag_pattern"`
	AutoPush    bool    `json:"auto_push"`
}

// SetRegistryConfig upserts the registry configuration for a project.
// PUT /api/projects/{id}/registry
func (h *Handlers) SetRegistryConfig(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if !userOwnsProject(r, h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var req setRegistryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.RegistryURL = strings.TrimSpace(req.RegistryURL)
	req.Repository = strings.TrimSpace(req.Repository)
	req.Username = strings.TrimSpace(req.Username)
	req.TagPattern = strings.TrimSpace(req.TagPattern)

	if req.RegistryURL == "" {
		http.Error(w, "registry_url required", http.StatusBadRequest)
		return
	}
	if len(req.RegistryURL) > 512 {
		http.Error(w, "registry_url max 512 chars", http.StatusBadRequest)
		return
	}
	if req.Repository == "" {
		http.Error(w, "repository required", http.StatusBadRequest)
		return
	}
	if len(req.Repository) > 256 {
		http.Error(w, "repository max 256 chars", http.StatusBadRequest)
		return
	}
	if len(req.Username) > 256 {
		http.Error(w, "username max 256 chars", http.StatusBadRequest)
		return
	}
	if req.TagPattern == "" {
		req.TagPattern = "{{branch}}-{{short_sha}}"
	}
	if len(req.TagPattern) > 256 {
		http.Error(w, "tag_pattern max 256 chars", http.StatusBadRequest)
		return
	}

	// Encrypt password if provided and Box is available.
	var passwordEnc []byte
	if req.Password != nil && *req.Password != "" {
		if h.Box == nil {
			http.Error(w, "WORKEND_TOKEN_KEY must be set to store registry passwords", http.StatusBadRequest)
			return
		}
		enc, err := h.Box.Seal([]byte(*req.Password))
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		passwordEnc = enc
	}

	var rc RegistryConfig
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO registry_configs (project_id, registry_url, repository, username, password_enc, tag_pattern, auto_push)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (project_id) DO UPDATE
		SET registry_url = EXCLUDED.registry_url,
		    repository   = EXCLUDED.repository,
		    username     = EXCLUDED.username,
		    password_enc = EXCLUDED.password_enc,
		    tag_pattern  = EXCLUDED.tag_pattern,
		    auto_push    = EXCLUDED.auto_push
		RETURNING id, project_id, registry_url, repository, username, tag_pattern, auto_push, created_at
	`, pid, req.RegistryURL, req.Repository, req.Username, passwordEnc, req.TagPattern, req.AutoPush).Scan(
		&rc.ID, &rc.ProjectID, &rc.RegistryURL, &rc.Repository, &rc.Username, &rc.TagPattern, &rc.AutoPush, &rc.CreatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rc)
}

// DeleteRegistryConfig removes the registry configuration for a project.
// DELETE /api/projects/{id}/registry
func (h *Handlers) DeleteRegistryConfig(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if !userOwnsProject(r, h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	tag, err := h.Pool.Exec(r.Context(),
		`DELETE FROM registry_configs WHERE project_id = $1`, pid)
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
