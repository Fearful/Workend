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
	"time"
)

// Gitea is the OAuth Provider for self-hosted Gitea instances.
type Gitea struct {
	clientID     string
	clientSecret string
	instanceURL  string
	instanceHost string
	redirectBase string
}

func NewGitea(clientID, clientSecret, instanceURL, redirectBase string) *Gitea {
	u, _ := url.Parse(instanceURL)
	return &Gitea{
		clientID:     clientID,
		clientSecret: clientSecret,
		instanceURL:  strings.TrimRight(instanceURL, "/"),
		instanceHost: u.Host,
		redirectBase: strings.TrimRight(redirectBase, "/"),
	}
}

func (g *Gitea) Kind() string         { return "gitea" }
func (g *Gitea) ID() string           { return MakeID(g.Kind(), g.instanceHost) }
func (g *Gitea) InstanceURL() string  { return g.instanceURL }
func (g *Gitea) InstanceHost() string { return g.instanceHost }

func (g *Gitea) AuthorizeURL(state, redirectURL string) string {
	q := url.Values{}
	q.Set("client_id", g.clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("state", state)
	return g.instanceURL + "/login/oauth/authorize?" + q.Encode()
}

func (g *Gitea) ExchangeCode(ctx context.Context, code, redirectURL string) (*Token, error) {
	body := url.Values{}
	body.Set("client_id", g.clientID)
	body.Set("client_secret", g.clientSecret)
	body.Set("code", code)
	body.Set("grant_type", "authorization_code")
	body.Set("redirect_uri", redirectURL)
	return g.tokenRequest(ctx, body)
}

func (g *Gitea) RefreshAccess(ctx context.Context, refreshToken string) (*Token, error) {
	body := url.Values{}
	body.Set("client_id", g.clientID)
	body.Set("client_secret", g.clientSecret)
	body.Set("refresh_token", refreshToken)
	body.Set("grant_type", "refresh_token")
	return g.tokenRequest(ctx, body)
}

func (g *Gitea) tokenRequest(ctx context.Context, body url.Values) (*Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.instanceURL+"/login/oauth/access_token", strings.NewReader(body.Encode()))
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
		return nil, fmt.Errorf("gitea token status %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if parsed.AccessToken == "" {
		return nil, errors.New("empty access token from gitea")
	}
	tok := &Token{Access: parsed.AccessToken, Refresh: parsed.RefreshToken, Scopes: parsed.Scope}
	if parsed.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second)
		tok.ExpiresAt = &exp
	}
	return tok, nil
}

func (g *Gitea) FetchHandle(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.instanceURL+"/api/v1/user", nil)
	if err != nil {
		return "", err
	}
	// Gitea accepts both "token <tok>" and "Bearer <tok>"; use Bearer for consistency.
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("gitea /user status %d", resp.StatusCode)
	}
	var u struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", err
	}
	return u.Login, nil
}

func (g *Gitea) ListRepos(ctx context.Context, accessToken string, page, perPage int) ([]Repo, error) {
	if perPage < 1 || perPage > 50 {
		perPage = 50
	}
	if page < 1 {
		page = 1
	}
	q := url.Values{}
	q.Set("limit", fmt.Sprintf("%d", perPage))
	q.Set("page", fmt.Sprintf("%d", page))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.instanceURL+"/api/v1/user/repos?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitea /user/repos status %d: %s", resp.StatusCode, string(raw))
	}
	var rows []struct {
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		Private       bool   `json:"private"`
		HTMLURL       string `json:"html_url"`
		CloneURL      string `json:"clone_url"`
		DefaultBranch string `json:"default_branch"`
		UpdatedAt     string `json:"updated_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	out := make([]Repo, 0, len(rows))
	for _, r := range rows {
		var ts *time.Time
		if r.UpdatedAt != "" {
			if t, err := time.Parse(time.RFC3339, r.UpdatedAt); err == nil {
				ts = &t
			}
		}
		out = append(out, Repo{
			Name: r.Name, FullName: r.FullName, Description: r.Description,
			Private: r.Private, HTMLURL: r.HTMLURL, CloneURL: r.CloneURL,
			DefaultBranch: r.DefaultBranch, UpdatedAt: ts,
		})
	}
	return out, nil
}

// InjectCloneAuth: https://<token>@host/path  (Gitea accepts token-as-username)
func (g *Gitea) InjectCloneAuth(rawURL, accessToken string) string {
	if !OwnsURL(g, rawURL) {
		return rawURL
	}
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(rawURL, scheme) {
			return scheme + accessToken + "@" + strings.TrimPrefix(rawURL, scheme)
		}
	}
	return rawURL
}
