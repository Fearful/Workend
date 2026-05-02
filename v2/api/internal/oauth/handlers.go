package oauth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"workend/api/internal/auth"
)

// POST /api/auth/github/start  → returns { url } to navigate to
func (g *GitHub) HandleStart(w http.ResponseWriter, r *http.Request) {
	url, err := g.StartURL(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}

// GET /api/auth/github/callback?code=…&state=…  (proxied from web)
// Closes the OAuth loop. Returns 204 on success.
func (g *GitHub) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if err := g.Callback(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/me/github  → connection status
func (g *GitHub) HandleStatus(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var handle, scopes string
	err := g.Pool.QueryRow(r.Context(), `
		SELECT handle, scopes FROM user_tokens
		WHERE user_id = $1 AND provider = 'github'
	`, uid).Scan(&handle, &scopes)
	if errors.Is(err, pgx.ErrNoRows) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"connected": false})
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"connected": true,
		"handle":    handle,
		"scopes":    scopes,
	})
}

// DELETE /api/me/github  → disconnect
func (g *GitHub) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	if _, err := g.Pool.Exec(r.Context(),
		`DELETE FROM user_tokens WHERE user_id = $1 AND provider = 'github'`, uid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
