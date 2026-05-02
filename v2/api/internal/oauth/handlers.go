package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
)

const (
	stateCookiePrefix = "workend_oauth_state_"
	stateCookieTTL    = 10 * time.Minute
)

type Handlers struct {
	Registry     *Registry
	RedirectBase string // e.g. "http://localhost:3000"
	Secure       bool
}

// callbackURLFor returns the public URL the provider should redirect back to.
func (h *Handlers) callbackURLFor(providerID string) string {
	return h.RedirectBase + "/auth/" + providerID + "/callback"
}

// POST /api/auth/{provider}/start
// Returns { url } for the browser to navigate to.
func (h *Handlers) Start(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")
	p, ok := h.Registry.ByID(providerID)
	if !ok {
		http.Error(w, "unknown provider", http.StatusNotFound)
		return
	}
	state, err := newState()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookiePrefix + providerID,
		Value:    state,
		Path:     "/",
		MaxAge:   int(stateCookieTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	authURL := p.AuthorizeURL(state, h.callbackURLFor(providerID))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": authURL})
}

// GET /api/auth/{provider}/callback?code=…&state=…
func (h *Handlers) Callback(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	providerID := chi.URLParam(r, "provider")
	p, ok := h.Registry.ByID(providerID)
	if !ok {
		http.Error(w, "unknown provider", http.StatusNotFound)
		return
	}
	stateCookie, err := r.Cookie(stateCookiePrefix + providerID)
	if err != nil {
		http.Error(w, "missing state cookie", http.StatusBadRequest)
		return
	}
	if stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "state mismatch", http.StatusBadRequest)
		return
	}
	// One-shot.
	http.SetCookie(w, &http.Cookie{
		Name: stateCookiePrefix + providerID, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: h.Secure, SameSite: http.SameSiteLaxMode,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	tok, err := p.ExchangeCode(r.Context(), code, h.callbackURLFor(providerID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	handle, err := p.FetchHandle(r.Context(), tok.Access)
	if err != nil {
		http.Error(w, fmt.Sprintf("fetch handle: %v", err), http.StatusBadGateway)
		return
	}
	if err := h.Registry.Save(r.Context(), uid, p, tok, handle); err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/me/connections — list of available providers + which are connected
type connectionView struct {
	ProviderID   string     `json:"provider_id"`
	Provider     string     `json:"provider"`
	InstanceURL  string     `json:"instance_url"`
	InstanceHost string     `json:"instance_host"`
	Connected    bool       `json:"connected"`
	Handle       string     `json:"handle,omitempty"`
	Scopes       string     `json:"scopes,omitempty"`
	ConnectedAt  *time.Time `json:"connected_at,omitempty"`
	ConnectionID *uuid.UUID `json:"connection_id,omitempty"`
}

func (h *Handlers) ListConnections(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	conns, err := h.Registry.ListConnections(r.Context(), uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	byProviderID := map[string]*Connection{}
	for i := range conns {
		byProviderID[conns[i].ProviderID] = &conns[i]
	}

	out := []connectionView{}
	for _, p := range h.Registry.All() {
		v := connectionView{
			ProviderID:   p.ID(),
			Provider:     p.Kind(),
			InstanceURL:  p.InstanceURL(),
			InstanceHost: p.InstanceHost(),
		}
		if c := byProviderID[p.ID()]; c != nil {
			v.Connected = true
			v.Handle = c.Handle
			v.Scopes = c.Scopes
			v.ConnectedAt = &c.ConnectedAt
			v.ConnectionID = &c.ID
		}
		out = append(out, v)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// DELETE /api/me/connections/{id}
func (h *Handlers) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.Registry.Delete(r.Context(), uid, id); err != nil {
		if errors.Is(err, ErrConnectionNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func newState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
