package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GitHub is the OAuth Provider for github.com (Enterprise via instanceURL).
type GitHub struct {
	clientID     string
	clientSecret string
	instanceURL  string // e.g. "https://github.com" or "https://github.acme.com"
	instanceHost string
	redirectBase string
}

func NewGitHub(clientID, clientSecret, instanceURL, redirectBase string) *GitHub {
	u, _ := url.Parse(instanceURL)
	return &GitHub{
		clientID:     clientID,
		clientSecret: clientSecret,
		instanceURL:  strings.TrimRight(instanceURL, "/"),
		instanceHost: u.Host,
		redirectBase: strings.TrimRight(redirectBase, "/"),
	}
}

func (g *GitHub) Kind() string         { return "github" }
func (g *GitHub) ID() string           { return MakeID(g.Kind(), g.instanceHost) }
func (g *GitHub) InstanceURL() string  { return g.instanceURL }
func (g *GitHub) InstanceHost() string { return g.instanceHost }

func (g *GitHub) authBase() string {
	if g.instanceHost == "github.com" {
		return "https://github.com"
	}
	return g.instanceURL
}

func (g *GitHub) apiBase() string {
	if g.instanceHost == "github.com" {
		return "https://api.github.com"
	}
	return g.instanceURL + "/api/v3"
}

func (g *GitHub) AuthorizeURL(state, redirectURL string) string {
	q := url.Values{}
	q.Set("client_id", g.clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("scope", "repo read:user")
	q.Set("state", state)
	q.Set("allow_signup", "false")
	return g.authBase() + "/login/oauth/authorize?" + q.Encode()
}

func (g *GitHub) ExchangeCode(ctx context.Context, code, redirectURL string) (*Token, error) {
	body := url.Values{}
	body.Set("client_id", g.clientID)
	body.Set("client_secret", g.clientSecret)
	body.Set("code", code)
	body.Set("redirect_uri", redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.authBase()+"/login/oauth/access_token", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github token exchange status %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		AccessToken      string `json:"access_token"`
		Scope            string `json:"scope"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != "" {
		return nil, fmt.Errorf("github: %s: %s", parsed.Error, parsed.ErrorDescription)
	}
	if parsed.AccessToken == "" {
		return nil, errors.New("empty access token from github")
	}
	return &Token{Access: parsed.AccessToken, Scopes: parsed.Scope}, nil
}

func (g *GitHub) RefreshAccess(ctx context.Context, refreshToken string) (*Token, error) {
	return nil, ErrRefreshUnsupported
}

func (g *GitHub) FetchHandle(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.apiBase()+"/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
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

// InjectCloneAuth: https://x-access-token:<token>@host/path
func (g *GitHub) InjectCloneAuth(rawURL, accessToken string) string {
	if !OwnsURL(g, rawURL) {
		return rawURL
	}
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(rawURL, scheme) {
			return scheme + "x-access-token:" + accessToken + "@" + strings.TrimPrefix(rawURL, scheme)
		}
	}
	return rawURL
}
