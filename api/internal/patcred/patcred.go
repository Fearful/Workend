// Package patcred owns user-managed personal access tokens used to clone
// private repos via HTTPS URLs on hosts where no OAuth provider is
// configured. Tokens are encrypted at rest with the same NaCl box used for
// OAuth tokens and SSH keys.
package patcred

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/secret"
)

// Credential is what the API returns to the UI. The token plaintext is
// never echoed back; only the host, label, and metadata.
type Credential struct {
	ID            uuid.UUID  `json:"id"`
	Host          string     `json:"host"`
	Label         string     `json:"label"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	LastRotatedAt *time.Time `json:"last_rotated_at"`
}

type Handlers struct {
	Pool *pgxpool.Pool
	Box  *secret.Box // nil when WORKEND_TOKEN_KEY is unset
}

// List returns the current user's PAT credentials (without token bytes).
// GET /api/me/pat-credentials
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, host, label, created_at, expires_at, last_rotated_at
		FROM user_pat_credentials WHERE user_id = $1 ORDER BY created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Credential{}
	for rows.Next() {
		var c Credential
		if err := rows.Scan(&c.ID, &c.Host, &c.Label, &c.CreatedAt, &c.ExpiresAt, &c.LastRotatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, c)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Create stores or replaces a PAT for (user, host).
// Body: {"host": "...", "label": "...", "token": "..."}.
// `host` is normalized to lowercase, scheme stripped, port stripped.
//
// POST /api/me/pat-credentials
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	if h.Box == nil {
		http.Error(w, "WORKEND_TOKEN_KEY must be set to store access tokens", http.StatusBadRequest)
		return
	}
	var req struct {
		Host      string     `json:"host"`
		Label     string     `json:"label"`
		Token     string     `json:"token"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	host := normalizeHost(req.Host)
	req.Label = strings.TrimSpace(req.Label)
	req.Token = strings.TrimSpace(req.Token)
	if host == "" {
		http.Error(w, "host required", http.StatusBadRequest)
		return
	}
	if req.Label == "" {
		http.Error(w, "label required", http.StatusBadRequest)
		return
	}
	if len(req.Token) < 8 {
		http.Error(w, "token looks too short to be valid", http.StatusBadRequest)
		return
	}

	enc, err := h.Box.Seal([]byte(req.Token))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var c Credential
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO user_pat_credentials (user_id, host, label, encrypted_token, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, host) DO UPDATE
		SET label           = EXCLUDED.label,
		    encrypted_token = EXCLUDED.encrypted_token,
		    expires_at      = EXCLUDED.expires_at,
		    created_at      = now()
		RETURNING id, host, label, created_at, expires_at, last_rotated_at
	`, uid, host, req.Label, enc, req.ExpiresAt).Scan(&c.ID, &c.Host, &c.Label, &c.CreatedAt, &c.ExpiresAt, &c.LastRotatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

// Delete removes one of the user's PAT credentials.
// DELETE /api/me/pat-credentials/:id
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	tag, err := h.Pool.Exec(r.Context(),
		`DELETE FROM user_pat_credentials WHERE id = $1 AND user_id = $2`, id, uid)
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

// RotateCredential marks a credential as rotated by setting last_rotated_at = now().
// Optionally updates expires_at from the request body.
// POST /api/me/pat-credentials/:id/rotate
func (h *Handlers) RotateCredential(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body struct {
		ExpiresAt *time.Time `json:"expires_at"`
	}
	// Body is optional; ignore decode errors for empty bodies.
	_ = json.NewDecoder(r.Body).Decode(&body)

	var c Credential
	err = h.Pool.QueryRow(r.Context(), `
		UPDATE user_pat_credentials
		SET last_rotated_at = now(),
		    expires_at = COALESCE($1, expires_at)
		WHERE id = $2 AND user_id = $3
		RETURNING id, host, label, created_at, expires_at, last_rotated_at
	`, body.ExpiresAt, id, uid).Scan(&c.ID, &c.Host, &c.Label, &c.CreatedAt, &c.ExpiresAt, &c.LastRotatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}

// LookupForCloneURL returns the decrypted token for whichever credential
// matches the host of rawURL, or "" if nothing matches. Used by the clone
// pipeline as a fallback after OAuth.
func LookupForCloneURL(ctx context.Context, pool *pgxpool.Pool, box *secret.Box, userID uuid.UUID, rawURL string) string {
	if box == nil {
		return ""
	}
	host := hostFromCloneURL(rawURL)
	if host == "" {
		return ""
	}
	var enc []byte
	err := pool.QueryRow(ctx, `
		SELECT encrypted_token FROM user_pat_credentials
		WHERE user_id = $1 AND host = $2
	`, userID, host).Scan(&enc)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ""
		}
		return ""
	}
	plain, err := box.Open(enc)
	if err != nil {
		return ""
	}
	return string(plain)
}

// normalizeHost lowercases and strips scheme + port from a user-provided
// host string. Accepts "github.com", "https://github.com", "github.com:443"
// and returns "github.com".
func normalizeHost(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[i+1:]
	}
	if i := strings.IndexByte(s, ':'); i >= 0 {
		s = s[:i]
	}
	return s
}

// hostFromCloneURL extracts the lowercase host from an HTTPS git URL.
// Returns "" for non-HTTP(S) URLs since PATs only apply to HTTPS clones.
func hostFromCloneURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	host := strings.ToLower(u.Host)
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	return host
}
