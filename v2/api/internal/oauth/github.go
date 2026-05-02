// Package oauth implements third-party OAuth flows. Currently: GitHub.
//
// State is signed-cookie based to avoid needing server-side state storage.
// The cookie contains the original CSRF state value and is verified on
// callback before the code is exchanged.
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/secret"
)

const (
	githubAuthURL  = "https://github.com/login/oauth/authorize"
	githubTokenURL = "https://github.com/login/oauth/access_token"
	githubAPIRoot  = "https://api.github.com"

	stateCookieName = "workend_oauth_state"
	stateCookieTTL  = 10 * time.Minute
)

type GitHub struct {
	Pool         *pgxpool.Pool
	Box          *secret.Box
	ClientID     string
	ClientSecret string
	Scopes       []string
	RedirectURL  string // public URL: e.g., http://localhost:3000/auth/github/callback
	Secure       bool
}

// Start writes the state cookie and returns the URL to redirect to.
func (g *GitHub) StartURL(w http.ResponseWriter) (string, error) {
	if g.ClientID == "" {
		return "", errors.New("GitHub OAuth not configured")
	}
	state, err := newState()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   int(stateCookieTTL.Seconds()),
		HttpOnly: true,
		Secure:   g.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	q := url.Values{}
	q.Set("client_id", g.ClientID)
	q.Set("redirect_uri", g.RedirectURL)
	q.Set("scope", strings.Join(g.Scopes, " "))
	q.Set("state", state)
	q.Set("allow_signup", "false")
	return githubAuthURL + "?" + q.Encode(), nil
}

// Callback handles the GitHub redirect, validates state, exchanges the code
// for an access token, fetches the user handle, and persists encrypted.
func (g *GitHub) Callback(w http.ResponseWriter, r *http.Request) error {
	if g.ClientID == "" {
		return errors.New("GitHub OAuth not configured")
	}
	uid := auth.UserID(r.Context())

	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		return fmt.Errorf("missing state cookie")
	}
	if stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		return fmt.Errorf("state mismatch")
	}
	// One-shot — clear it.
	http.SetCookie(w, &http.Cookie{
		Name: stateCookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: g.Secure, SameSite: http.SameSiteLaxMode,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		return fmt.Errorf("missing code")
	}

	token, scopes, err := g.exchange(r.Context(), code)
	if err != nil {
		return err
	}
	handle, err := g.fetchHandle(r.Context(), token)
	if err != nil {
		return err
	}

	enc, err := g.Box.Seal([]byte(token))
	if err != nil {
		return err
	}

	_, err = g.Pool.Exec(r.Context(), `
		INSERT INTO user_tokens (user_id, provider, access_token, scopes, handle)
		VALUES ($1, 'github', $2, $3, $4)
		ON CONFLICT (user_id, provider) DO UPDATE
		SET access_token = EXCLUDED.access_token,
		    scopes       = EXCLUDED.scopes,
		    handle       = EXCLUDED.handle,
		    connected_at = now()
	`, uid, enc, scopes, handle)
	return err
}

// Token returns the (decrypted) GitHub token for a user, or "" if not connected.
func (g *GitHub) Token(ctx context.Context, userID string) (string, error) {
	var enc []byte
	err := g.Pool.QueryRow(ctx, `
		SELECT access_token FROM user_tokens
		WHERE user_id = $1 AND provider = 'github'
	`, userID).Scan(&enc)
	if err != nil {
		return "", err
	}
	plain, err := g.Box.Open(enc)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (g *GitHub) exchange(ctx context.Context, code string) (token, scopes string, err error) {
	body := url.Values{}
	body.Set("client_id", g.ClientID)
	body.Set("client_secret", g.ClientSecret)
	body.Set("code", code)
	body.Set("redirect_uri", g.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("token exchange status %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		AccessToken      string `json:"access_token"`
		Scope            string `json:"scope"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", "", err
	}
	if parsed.Error != "" {
		return "", "", fmt.Errorf("github: %s: %s", parsed.Error, parsed.ErrorDescription)
	}
	if parsed.AccessToken == "" {
		return "", "", errors.New("empty access token from github")
	}
	return parsed.AccessToken, parsed.Scope, nil
}

func (g *GitHub) fetchHandle(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIRoot+"/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github /user status %d", resp.StatusCode)
	}
	var u struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", err
	}
	return u.Login, nil
}

func newState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
