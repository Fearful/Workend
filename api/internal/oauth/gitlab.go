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

func (g *GitLab) ListBranches(ctx context.Context, accessToken, fullName string) ([]Branch, error) {
	defaultBranch := ""
	{
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.instanceURL+"/api/v4/projects/"+url.PathEscape(fullName), nil)
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+accessToken)
			if resp, err := http.DefaultClient.Do(req); err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					var meta struct {
						DefaultBranch string `json:"default_branch"`
					}
					if json.NewDecoder(resp.Body).Decode(&meta) == nil {
						defaultBranch = meta.DefaultBranch
					}
				}
			}
		}
	}

	q := url.Values{}
	q.Set("per_page", "100")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.instanceURL+"/api/v4/projects/"+url.PathEscape(fullName)+"/repository/branches?"+q.Encode(), nil)
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
		return nil, fmt.Errorf("gitlab /branches status %d: %s", resp.StatusCode, string(raw))
	}
	var rows []struct {
		Name      string `json:"name"`
		Protected bool   `json:"protected"`
		Default   bool   `json:"default"`
		Commit    struct {
			ID string `json:"id"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	out := make([]Branch, 0, len(rows))
	for _, r := range rows {
		isDefault := r.Default
		if defaultBranch != "" && r.Name == defaultBranch {
			isDefault = true
		}
		out = append(out, Branch{
			Name:      r.Name,
			CommitSHA: r.Commit.ID,
			Protected: r.Protected,
			Default:   isDefault,
		})
	}
	return out, nil
}

func (g *GitLab) CreatePullRequest(ctx context.Context, accessToken, fullName string, in PullRequestInput) (*PullRequestResult, error) {
	payload := map[string]any{
		"source_branch": in.Source,
		"target_branch": in.Target,
		"title":         in.Title,
		"description":   in.Body,
	}
	buf, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.instanceURL+"/api/v4/projects/"+url.PathEscape(fullName)+"/merge_requests",
		strings.NewReader(string(buf)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("gitlab create MR status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		WebURL string `json:"web_url"`
		IID    int    `json:"iid"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	return &PullRequestResult{URL: parsed.WebURL, Number: parsed.IID}, nil
}

func (g *GitLab) ListLabels(ctx context.Context, accessToken, fullName string) ([]Label, error) {
	q := url.Values{}
	q.Set("per_page", "100")
	out, err := glDoJSON[[]struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}](ctx, accessToken, http.MethodGet,
		g.instanceURL+"/api/v4/projects/"+url.PathEscape(fullName)+"/labels?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	res := make([]Label, 0, len(out))
	for _, l := range out {
		c := strings.TrimPrefix(l.Color, "#")
		res = append(res, Label{Name: l.Name, Color: c})
	}
	return res, nil
}

func (g *GitLab) ListIssues(ctx context.Context, accessToken, fullName string) ([]Issue, error) {
	q := url.Values{}
	q.Set("per_page", "100")
	q.Set("scope", "all")
	q.Set("order_by", "updated_at")
	rows, err := glDoJSON[[]struct {
		IID         int      `json:"iid"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		State       string   `json:"state"` // 'opened' | 'closed'
		WebURL      string   `json:"web_url"`
		UpdatedAt   string   `json:"updated_at"`
		Author      struct {
			Username string `json:"username"`
			WebURL   string `json:"web_url"`
		} `json:"author"`
		Labels []string `json:"labels"`
	}](ctx, accessToken, http.MethodGet,
		g.instanceURL+"/api/v4/projects/"+url.PathEscape(fullName)+"/issues?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	out := make([]Issue, 0, len(rows))
	for _, r := range rows {
		state := "open"
		if r.State == "closed" {
			state = "closed"
		}
		var ts *time.Time
		if r.UpdatedAt != "" {
			if t, err := time.Parse(time.RFC3339, r.UpdatedAt); err == nil {
				ts = &t
			}
		}
		out = append(out, Issue{
			Number: r.IID, Title: r.Title, Body: r.Description, State: state,
			Labels:       append([]string{}, r.Labels...),
			AuthorHandle: r.Author.Username, AuthorURL: r.Author.WebURL,
			HTMLURL: r.WebURL, UpdatedAt: ts,
		})
	}
	return out, nil
}

func (g *GitLab) ListIssueComments(ctx context.Context, accessToken, fullName string, number int) ([]IssueComment, error) {
	rows, err := glDoJSON[[]struct {
		ID        int64  `json:"id"`
		Body      string `json:"body"`
		System    bool   `json:"system"`
		CreatedAt string `json:"created_at"`
		Author    struct {
			Username string `json:"username"`
		} `json:"author"`
	}](ctx, accessToken, http.MethodGet,
		fmt.Sprintf("%s/api/v4/projects/%s/issues/%d/notes?per_page=100",
			g.instanceURL, url.PathEscape(fullName), number), nil)
	if err != nil {
		return nil, err
	}
	out := make([]IssueComment, 0, len(rows))
	for _, r := range rows {
		// GitLab "system" notes are activity entries (label changed,
		// closed, etc.) — they're noisy and not what users mean by
		// comments. Drop them.
		if r.System {
			continue
		}
		t, _ := time.Parse(time.RFC3339, r.CreatedAt)
		out = append(out, IssueComment{
			ProviderID: r.ID, Body: r.Body, AuthorHandle: r.Author.Username,
			CreatedAt: t,
		})
	}
	return out, nil
}

func (g *GitLab) CreateIssueComment(ctx context.Context, accessToken, fullName string, number int, body string) (*IssueComment, error) {
	type req struct {
		Body string `json:"body"`
	}
	row, err := glDoJSON[struct {
		ID        int64  `json:"id"`
		Body      string `json:"body"`
		CreatedAt string `json:"created_at"`
		Author    struct {
			Username string `json:"username"`
		} `json:"author"`
	}](ctx, accessToken, http.MethodPost,
		fmt.Sprintf("%s/api/v4/projects/%s/issues/%d/notes",
			g.instanceURL, url.PathEscape(fullName), number),
		req{Body: body})
	if err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, row.CreatedAt)
	return &IssueComment{
		ProviderID: row.ID, Body: row.Body, AuthorHandle: row.Author.Username,
		CreatedAt: t,
	}, nil
}

func (g *GitLab) SetIssueLabels(ctx context.Context, accessToken, fullName string, number int, labels []string) error {
	type req struct {
		Labels string `json:"labels"`
	}
	_, err := glDoJSON[any](ctx, accessToken, http.MethodPut,
		fmt.Sprintf("%s/api/v4/projects/%s/issues/%d", g.instanceURL, url.PathEscape(fullName), number),
		req{Labels: strings.Join(labels, ",")})
	return err
}

func (g *GitLab) CloseIssue(ctx context.Context, accessToken, fullName string, number int) error {
	type req struct {
		StateEvent string `json:"state_event"`
	}
	_, err := glDoJSON[any](ctx, accessToken, http.MethodPut,
		fmt.Sprintf("%s/api/v4/projects/%s/issues/%d", g.instanceURL, url.PathEscape(fullName), number),
		req{StateEvent: "close"})
	return err
}

// glDoJSON is the GitLab-flavored generic JSON request helper.
func glDoJSON[T any](ctx context.Context, accessToken, method, urlStr string, body any) (T, error) {
	var zero T
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return zero, err
		}
		reader = strings.NewReader(string(buf))
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, reader)
	if err != nil {
		return zero, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return zero, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if len(raw) == 0 {
		return zero, nil
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, err
	}
	return out, nil
}

// TriggerCIWorkflow runs a GitLab pipeline against `branch`. GitLab doesn't
// have per-workflow files — the whole pipeline kicks off from .gitlab-ci.yml
// — so the `workflow` argument is ignored. Variables map to pipeline vars.
func (g *GitLab) TriggerCIWorkflow(ctx context.Context, accessToken, fullName, workflow, branch string, inputs map[string]string) (*CIDispatchResult, error) {
	payload := map[string]any{"ref": branch}
	if len(inputs) > 0 {
		vars := make([]map[string]string, 0, len(inputs))
		for k, v := range inputs {
			vars = append(vars, map[string]string{"key": k, "value": v})
		}
		payload["variables"] = vars
	}
	row, err := glDoJSON[struct {
		WebURL string `json:"web_url"`
		ID     int    `json:"id"`
	}](ctx, accessToken, http.MethodPost,
		fmt.Sprintf("%s/api/v4/projects/%s/pipeline", g.instanceURL, url.PathEscape(fullName)),
		payload)
	if err != nil {
		return nil, err
	}
	return &CIDispatchResult{URL: row.WebURL}, nil
}

func (g *GitLab) ListPipelineRuns(ctx context.Context, accessToken, fullName string, limit int) ([]PipelineRunInfo, error) {
	type glPipeline struct {
		ID        int64      `json:"id"`
		Status    string     `json:"status"`
		Ref       string     `json:"ref"`
		SHA       string     `json:"sha"`
		WebURL    string     `json:"web_url"`
		CreatedAt *time.Time `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
	}

	encoded := url.PathEscape(fullName)
	pipelines, err := glDoJSON[[]glPipeline](ctx, accessToken, http.MethodGet,
		fmt.Sprintf("%s/api/v4/projects/%s/pipelines?per_page=%d&order_by=id&sort=desc",
			g.instanceURL, encoded, limit), nil)
	if err != nil {
		return nil, err
	}

	out := make([]PipelineRunInfo, 0, len(pipelines))
	for _, p := range pipelines {
		status := "pending"
		switch p.Status {
		case "success":
			status = "success"
		case "failed":
			status = "failure"
		case "running":
			status = "running"
		case "canceled":
			status = "cancelled"
		default:
			status = "pending"
		}

		var finished *time.Time
		if status == "success" || status == "failure" || status == "cancelled" {
			finished = p.UpdatedAt
		}

		out = append(out, PipelineRunInfo{
			ProviderRunID: fmt.Sprintf("%d", p.ID),
			Status:        status,
			Branch:        p.Ref,
			CommitSHA:     p.SHA,
			WorkflowName:  "pipeline",
			HTMLURL:       p.WebURL,
			StartedAt:     p.CreatedAt,
			FinishedAt:    finished,
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
