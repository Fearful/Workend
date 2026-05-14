package workspace

import (
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
	"workend/api/internal/secret"
)

var secretKeyRegexp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// SecretHandlers manages workspace-level encrypted secrets.
type SecretHandlers struct {
	Pool *pgxpool.Pool
	Box  *secret.Box
}

// SecretMeta is the public representation of a secret — value is never
// exposed through the API.
type SecretMeta struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListSecrets returns all secrets for a workspace without their values.
// Membership is required.
//
// GET /api/workspaces/{id}/secrets
func (h *SecretHandlers) ListSecrets(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}

	if _, err := RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, workspace_id, key, description, created_by, created_at, updated_at
		FROM workspace_secrets
		WHERE workspace_id = $1
		ORDER BY key
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []SecretMeta{}
	for rows.Next() {
		var s SecretMeta
		if err := rows.Scan(&s.ID, &s.WorkspaceID, &s.Key, &s.Description,
			&s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, s)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// CreateSecret encrypts and stores a new workspace secret. Only workspace
// owners may create secrets.
//
// POST /api/workspaces/{id}/secrets
func (h *SecretHandlers) CreateSecret(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}

	if h.Box == nil {
		http.Error(w, "encryption not configured", http.StatusNotImplemented)
		return
	}

	role, err := RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if role != RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	body.Key = strings.TrimSpace(body.Key)
	body.Description = strings.TrimSpace(body.Description)

	if body.Key == "" || len(body.Key) > 100 {
		http.Error(w, "key is required (1-100 chars)", http.StatusBadRequest)
		return
	}
	if !secretKeyRegexp.MatchString(body.Key) {
		http.Error(w, "key must match ^[A-Za-z_][A-Za-z0-9_]*$", http.StatusBadRequest)
		return
	}
	if body.Value == "" || len(body.Value) > 10000 {
		http.Error(w, "value is required (max 10000 chars)", http.StatusBadRequest)
		return
	}
	if len(body.Description) > 500 {
		http.Error(w, "description max 500 chars", http.StatusBadRequest)
		return
	}

	enc, err := h.Box.Seal([]byte(body.Value))
	if err != nil {
		http.Error(w, "encryption failed", http.StatusInternalServerError)
		return
	}

	var s SecretMeta
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO workspace_secrets (workspace_id, key, value_enc, description, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, workspace_id, key, description, created_by, created_at, updated_at
	`, wsID, body.Key, enc, body.Description, uid).Scan(
		&s.ID, &s.WorkspaceID, &s.Key, &s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "workspace_secrets_ws_key_idx") {
			http.Error(w, "secret key already exists in this workspace", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(s)
}

// UpdateSecret re-encrypts a secret value and optionally updates its
// description. Only workspace owners may update.
//
// PUT /api/workspaces/{id}/secrets/{secret_id}
func (h *SecretHandlers) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	secretID, err := uuid.Parse(chi.URLParam(r, "secret_id"))
	if err != nil {
		http.Error(w, "invalid secret id", http.StatusBadRequest)
		return
	}

	if h.Box == nil {
		http.Error(w, "encryption not configured", http.StatusNotImplemented)
		return
	}

	role, err := RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if role != RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		Value       string  `json:"value"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if body.Value == "" || len(body.Value) > 10000 {
		http.Error(w, "value is required (max 10000 chars)", http.StatusBadRequest)
		return
	}
	if body.Description != nil && len(*body.Description) > 500 {
		http.Error(w, "description max 500 chars", http.StatusBadRequest)
		return
	}

	enc, err := h.Box.Seal([]byte(body.Value))
	if err != nil {
		http.Error(w, "encryption failed", http.StatusInternalServerError)
		return
	}

	var s SecretMeta
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE workspace_secrets
		SET value_enc = $1,
		    description = COALESCE($2, description),
		    updated_at = now()
		WHERE id = $3 AND workspace_id = $4
		RETURNING id, workspace_id, key, description, created_by, created_at, updated_at
	`, enc, body.Description, secretID, wsID).Scan(
		&s.ID, &s.WorkspaceID, &s.Key, &s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}

// DeleteSecret removes a workspace secret. Only workspace owners may delete.
//
// DELETE /api/workspaces/{id}/secrets/{secret_id}
func (h *SecretHandlers) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	secretID, err := uuid.Parse(chi.URLParam(r, "secret_id"))
	if err != nil {
		http.Error(w, "invalid secret id", http.StatusBadRequest)
		return
	}

	role, err := RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if role != RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	tag, err := h.Pool.Exec(r.Context(), `
		DELETE FROM workspace_secrets WHERE id = $1 AND workspace_id = $2
	`, secretID, wsID)
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
