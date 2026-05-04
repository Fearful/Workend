// Package sshkey owns user-managed SSH keys used to clone private repos via
// the git@host:path URL form. Private keys are encrypted at rest with the
// same NaCl box used for OAuth tokens.
package sshkey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/ssh"

	"workend/api/internal/auth"
	"workend/api/internal/secret"
)

// Key is what the API returns to the UI. The private key is never echoed
// back; only the derived public key.
type Key struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	PublicKey string    `json:"public_key"`
	CreatedAt time.Time `json:"created_at"`
}

type Handlers struct {
	Pool *pgxpool.Pool
	Box  *secret.Box // nil when WORKEND_TOKEN_KEY is unset
}

// List returns the current user's SSH keys.
// GET /api/me/ssh-keys
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, name, public_key, created_at
		FROM user_ssh_keys WHERE user_id = $1 ORDER BY created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Key{}
	for rows.Next() {
		var k Key
		if err := rows.Scan(&k.ID, &k.Name, &k.PublicKey, &k.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, k)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Create stores a new SSH key. Body: {"name": "...", "private_key": "..."}.
// The private key must be a PEM-encoded RSA / Ed25519 / ECDSA private key
// (encrypted or unencrypted; encrypted keys are rejected since we can't
// decrypt them headlessly). Public key is derived and persisted alongside.
//
// POST /api/me/ssh-keys
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	if h.Box == nil {
		http.Error(w, "WORKEND_TOKEN_KEY must be set to store SSH keys", http.StatusBadRequest)
		return
	}
	var req struct {
		Name       string `json:"name"`
		PrivateKey string `json:"private_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.PrivateKey = strings.TrimSpace(req.PrivateKey) + "\n"
	if req.Name == "" || len(req.PrivateKey) < 50 {
		http.Error(w, "name and private_key required", http.StatusBadRequest)
		return
	}

	signer, err := ssh.ParsePrivateKey([]byte(req.PrivateKey))
	if err != nil {
		// Accept the case where the key is passphrase-protected only with
		// an empty passphrase (rare but valid PEM). Anything else: reject.
		if _, isMissing := err.(*ssh.PassphraseMissingError); isMissing {
			http.Error(w, "encrypted SSH keys are not supported — supply an unencrypted key", http.StatusBadRequest)
			return
		}
		http.Error(w, "invalid private key: "+err.Error(), http.StatusBadRequest)
		return
	}
	pubKey := strings.TrimRight(string(ssh.MarshalAuthorizedKey(signer.PublicKey())), "\n")

	enc, err := h.Box.Seal([]byte(req.PrivateKey))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var k Key
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO user_ssh_keys (user_id, name, encrypted_private_key, public_key)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, public_key, created_at
	`, uid, req.Name, enc, pubKey).Scan(&k.ID, &k.Name, &k.PublicKey, &k.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "user_ssh_keys_user_name_idx") {
			http.Error(w, "you already have a key with that name", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(k)
}

// Delete removes one of the user's keys.
// DELETE /api/me/ssh-keys/:id
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	tag, err := h.Pool.Exec(r.Context(),
		`DELETE FROM user_ssh_keys WHERE id = $1 AND user_id = $2`, id, uid)
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

// LookupAnyKeyForUser returns the decrypted private-key bytes of any one
// of the user's keys (most recent first). Used by the clone pipeline when
// it sees an ssh:// or git@ URL. Returns ("", nil, nil) when no keys exist
// — caller will report "no SSH key configured" to the user.
func LookupAnyKeyForUser(ctx context.Context, pool *pgxpool.Pool, box *secret.Box, userID uuid.UUID) (privateKey string, publicKey string, err error) {
	if box == nil {
		return "", "", errors.New("token key not configured; cannot decrypt SSH keys")
	}
	var enc []byte
	if err := pool.QueryRow(ctx, `
		SELECT encrypted_private_key, public_key FROM user_ssh_keys
		WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&enc, &publicKey); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("load ssh key: %w", err)
	}
	plain, err := box.Open(enc)
	if err != nil {
		return "", "", fmt.Errorf("decrypt ssh key: %w", err)
	}
	return string(plain), publicKey, nil
}

// IsSSHURL returns true when the URL looks like git+SSH (the `git@host:path`
// shorthand or an explicit `ssh://` scheme). HTTPS URLs return false.
func IsSSHURL(u string) bool {
	u = strings.TrimSpace(u)
	if strings.HasPrefix(u, "ssh://") {
		return true
	}
	// `git@host:path/to/repo.git` form. No scheme; userinfo before colon.
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return false
	}
	at := strings.Index(u, "@")
	colon := strings.Index(u, ":")
	return at > 0 && colon > at
}
