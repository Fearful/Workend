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

// GitLab is the OAuth Provider for gitlab.com or self-hosted GitLab.
type GitLab struct {
	clientID     string
	clientSecret string
	instanceURL  string
	instanceHost string
	redirectBase string
}

func NewGitLab(clientID, clientSecret, instanceURL, redirectBase string) *GitLab {
	u, _ := url.Parse(instanceURL)
	return &GitLab{
		clientID:     clientID,
		clientSecret: clientSecret,
		instanceURL:  strings.TrimRight(instanceURL, "/"),
		instanceHost: u.Host,
		redirectBase: strings.TrimRight(redirectBase, "/"),
	}
}

func (g *GitLab) Kind() string         { return "gitlab" }
func (g *GitLab) ID() string           { return MakeID(g.Kind(), g.instanceHost) }
func (g *GitLab) InstanceURL() string  { return g.instanceURL }
func (g *GitLab) InstanceHost() string { return g.instanceHost }

func (g *GitLab) AuthorizeURL(state, redirectURL string) string {
	q := url.Values{}
	q.Set("client_id", g.clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "read_user read_repository write_repository api")
	q.Set("state", state)
	return g.instanceURL + "/oauth/authorize?" + q.Encode()
}

func (g *GitLab) ExchangeCode(ctx context.Context, code, redirectURL string) (*Token, error) {
	body := url.Values{}
	body.Set("client_id", g.clientID)
	body.Set("client_secret", g.clientSecret)
	body.Set("code", code)
	body.Set("grant_type", "authorization_code")
	body.Set("redirect_uri", redirectURL)
	return g.tokenRequest(ctx, body)
}

func (g *GitLab) RefreshAccess(ctx context.Context, refreshToken string) (*Token, error) {
	body := url.Values{}
	body.Set("client_id", g.clientID)
	body.Set("client_secret", g.clientSecret)
	body.Set("refresh_token", refreshToken)
	body.Set("grant_type", "refresh_token")
	return g.tokenRequest(ctx, body)
}

func (g *GitLab) tokenRequest(ctx context.Context, body url.Values) (*Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.instanceURL+"/oauth/token", strings.NewReader(body.Encode()))
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
		return nil, fmt.Errorf("gitlab token status %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		Scope            string `json:"scope"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != "" {
		return nil, fmt.Errorf("gitlab: %s: %s", parsed.Error, parsed.ErrorDescription)
	}
	if parsed.AccessToken == "" {
		return nil, errors.New("empty access token from gitlab")
	}
	tok := &Token{Access: parsed.AccessToken, Refresh: parsed.RefreshToken, Scopes: parsed.Scope}
	if parsed.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second)
		tok.ExpiresAt = &exp
	}
	return tok, nil
}

func (g *GitLab) FetchHandle(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.instanceURL+"/api/v4/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("gitlab /user status %d", resp.StatusCode)
	}
	var u struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", err
	}
	return u.Username, nil
}

func (g *GitLab) ListRepos(ctx context.Context, accessToken string, page, perPage int) ([]Repo, error) {
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}
	if page < 1 {
		page = 1
	}
	q := url.Values{}
	q.Set("per_page", fmt.Sprintf("%d", perPage))
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("membership", "true")
	q.Set("simple", "true")
	q.Set("order_by", "last_activity_at")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.instanceURL+"/api/v4/projects?"+q.Encode(), nil)
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
		return nil, fmt.Errorf("gitlab /projects status %d: %s", resp.StatusCode, string(raw))
	}
	var rows []struct {
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		Description       string `json:"description"`
		Visibility        string `json:"visibility"`
		WebURL            string `json:"web_url"`
		HTTPURLToRepo     string `json:"http_url_to_repo"`
		DefaultBranch     string `json:"default_branch"`
		LastActivityAt    string `json:"last_activity_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	out := make([]Repo, 0, len(rows))
	for _, r := range rows {
		var ts *time.Time
		if r.LastActivityAt != "" {
			if t, err := time.Parse(time.RFC3339, r.LastActivityAt); err == nil {
				ts = &t
			}
		}
		out = append(out, Repo{
			Name: r.Name, FullName: r.PathWithNamespace, Description: r.Description,
			Private: r.Visibility != "public", HTMLURL: r.WebURL, CloneURL: r.HTTPURLToRepo,
			DefaultBranch: r.DefaultBranch, UpdatedAt: ts,
		})
	}
	return out, nil
}

// InjectCloneAuth: https://oauth2:<token>@host/path
func (g *GitLab) InjectCloneAuth(rawURL, accessToken string) string {
	if !OwnsURL(g, rawURL) {
		return rawURL
	}
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(rawURL, scheme) {
			return scheme + "oauth2:" + accessToken + "@" + strings.TrimPrefix(rawURL, scheme)
		}
	}
	return rawURL
}
