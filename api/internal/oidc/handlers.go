package oidc

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/audit"
)

type Handlers struct {
	Pool         *pgxpool.Pool
	Doc          *DiscoveryDoc
	ClientID     string
	ClientSecret string
	RedirectBase string
	ProviderName string
	Secure       bool
	Audit        *audit.Logger
	Logger       *slog.Logger
}

const stateCookieName = "workend_oidc_state"

func (h *Handlers) redirectURI() string {
	return strings.TrimRight(h.RedirectBase, "/") + "/api/auth/oidc/callback"
}

func (h *Handlers) Start(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	state := base64.RawURLEncoding.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	url := AuthorizeURL(h.Doc, h.ClientID, h.redirectURI(), state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handlers) Callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || stateCookie.Value == "" {
		http.Error(w, "missing state cookie", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   stateCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	queryState := r.URL.Query().Get("state")
	if queryState == "" || queryState != stateCookie.Value {
		http.Error(w, "state mismatch", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errMsg := r.URL.Query().Get("error_description")
		if errMsg == "" {
			errMsg = r.URL.Query().Get("error")
		}
		if errMsg == "" {
			errMsg = "no authorization code"
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	tok, err := ExchangeCode(r.Context(), h.Doc, h.ClientID, h.ClientSecret, h.redirectURI(), code)
	if err != nil {
		h.Logger.Warn("oidc token exchange", "err", err)
		http.Error(w, "token exchange failed", http.StatusBadGateway)
		return
	}

	info, err := FetchUserInfo(r.Context(), h.Doc, tok.AccessToken)
	if err != nil {
		h.Logger.Warn("oidc userinfo", "err", err)
		http.Error(w, "userinfo fetch failed", http.StatusBadGateway)
		return
	}

	email := strings.TrimSpace(strings.ToLower(info.Email))
	if email == "" {
		http.Error(w, "oidc provider did not return an email", http.StatusBadGateway)
		return
	}
	displayName := info.Name
	if displayName == "" {
		displayName = email
	}

	var userID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT id FROM users WHERE oidc_sub = $1 AND oidc_issuer = $2
	`, info.Sub, h.Doc.Issuer).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		err = h.Pool.QueryRow(r.Context(), `
			SELECT id FROM users WHERE lower(email) = $1
		`, email).Scan(&userID)

		if errors.Is(err, pgx.ErrNoRows) {
			err = h.Pool.QueryRow(r.Context(), `
				INSERT INTO users (email, display_name, oidc_sub, oidc_issuer)
				VALUES ($1, $2, $3, $4)
				RETURNING id
			`, email, displayName, info.Sub, h.Doc.Issuer).Scan(&userID)
			if err != nil {
				h.Logger.Error("oidc create user", "err", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			h.Logger.Error("oidc user lookup", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		} else {
			if _, err := h.Pool.Exec(r.Context(), `
				UPDATE users SET oidc_sub = $1, oidc_issuer = $2 WHERE id = $3
			`, info.Sub, h.Doc.Issuer, userID); err != nil {
				h.Logger.Warn("oidc link existing user", "err", err)
			}
		}
	} else if err != nil {
		h.Logger.Error("oidc sub lookup", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	raw, expiresAt, err := auth.CreateSession(r.Context(), h.Pool, userID, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    raw,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	if h.Audit != nil {
		h.Audit.Record(r.Context(), userID, audit.UserLogin, "user", userID.String(), r.RemoteAddr,
			map[string]any{"method": "oidc", "issuer": h.Doc.Issuer})
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

type AuthConfigResponse struct {
	OIDCConfigured  bool   `json:"oidc_configured"`
	OIDCProviderName string `json:"oidc_provider_name,omitempty"`
}

func AuthConfigHandler(configured bool, providerName string) http.HandlerFunc {
	resp := AuthConfigResponse{
		OIDCConfigured:  configured,
		OIDCProviderName: providerName,
	}
	body, _ := json.Marshal(resp)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=60")
		_, _ = w.Write(body)
	}
}


