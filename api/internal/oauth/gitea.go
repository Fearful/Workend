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

func (g *Gitea) ListBranches(ctx context.Context, accessToken, fullName string) ([]Branch, error) {
	defaultBranch := ""
	{
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.instanceURL+"/api/v1/repos/"+fullName, nil)
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
	q.Set("limit", "50")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.instanceURL+"/api/v1/repos/"+fullName+"/branches?"+q.Encode(), nil)
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
		return nil, fmt.Errorf("gitea /branches status %d: %s", resp.StatusCode, string(raw))
	}
	var rows []struct {
		Name   string `json:"name"`
		Commit struct {
			ID string `json:"id"`
		} `json:"commit"`
		Protected bool `json:"protected"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	out := make([]Branch, 0, len(rows))
	for _, r := range rows {
		out = append(out, Branch{
			Name:      r.Name,
			CommitSHA: r.Commit.ID,
			Protected: r.Protected,
			Default:   r.Name == defaultBranch,
		})
	}
	return out, nil
}

func (g *Gitea) CreatePullRequest(ctx context.Context, accessToken, fullName string, in PullRequestInput) (*PullRequestResult, error) {
	payload := map[string]any{
		"head":  in.Source,
		"base":  in.Target,
		"title": in.Title,
		"body":  in.Body,
	}
	buf, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.instanceURL+"/api/v1/repos/"+fullName+"/pulls", strings.NewReader(string(buf)))
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
		return nil, fmt.Errorf("gitea create PR status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		HTMLURL string `json:"html_url"`
		Number  int    `json:"number"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	return &PullRequestResult{URL: parsed.HTMLURL, Number: parsed.Number}, nil
}

func (g *Gitea) ListLabels(ctx context.Context, accessToken, fullName string) ([]Label, error) {
	q := url.Values{}
	q.Set("limit", "50")
	out, err := giteaDoJSON[[]struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}](ctx, accessToken, http.MethodGet, g.instanceURL+"/api/v1/repos/"+fullName+"/labels?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	res := make([]Label, 0, len(out))
	for _, l := range out {
		res = append(res, Label{Name: l.Name, Color: l.Color})
	}
	return res, nil
}

func (g *Gitea) ListIssues(ctx context.Context, accessToken, fullName string) ([]Issue, error) {
	q := url.Values{}
	q.Set("limit", "50")
	q.Set("state", "all")
	q.Set("type", "issues") // exclude PRs
	rows, err := giteaDoJSON[[]struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		State     string `json:"state"`
		HTMLURL   string `json:"html_url"`
		UpdatedAt string `json:"updated_at"`
		User      struct {
			Login   string `json:"login"`
			HTMLURL string `json:"html_url"`
		} `json:"user"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}](ctx, accessToken, http.MethodGet, g.instanceURL+"/api/v1/repos/"+fullName+"/issues?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	out := make([]Issue, 0, len(rows))
	for _, r := range rows {
		labels := make([]string, 0, len(r.Labels))
		for _, l := range r.Labels {
			labels = append(labels, l.Name)
		}
		var ts *time.Time
		if r.UpdatedAt != "" {
			if t, err := time.Parse(time.RFC3339, r.UpdatedAt); err == nil {
				ts = &t
			}
		}
		out = append(out, Issue{
			Number: r.Number, Title: r.Title, Body: r.Body, State: r.State,
			Labels:       labels,
			AuthorHandle: r.User.Login, AuthorURL: r.User.HTMLURL,
			HTMLURL: r.HTMLURL, UpdatedAt: ts,
		})
	}
	return out, nil
}

func (g *Gitea) ListIssueComments(ctx context.Context, accessToken, fullName string, number int) ([]IssueComment, error) {
	rows, err := giteaDoJSON[[]struct {
		ID        int64  `json:"id"`
		Body      string `json:"body"`
		HTMLURL   string `json:"html_url"`
		CreatedAt string `json:"created_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
	}](ctx, accessToken, http.MethodGet,
		fmt.Sprintf("%s/api/v1/repos/%s/issues/%d/comments", g.instanceURL, fullName, number), nil)
	if err != nil {
		return nil, err
	}
	out := make([]IssueComment, 0, len(rows))
	for _, r := range rows {
		t, _ := time.Parse(time.RFC3339, r.CreatedAt)
		out = append(out, IssueComment{
			ProviderID: r.ID, Body: r.Body, AuthorHandle: r.User.Login,
			HTMLURL: r.HTMLURL, CreatedAt: t,
		})
	}
	return out, nil
}

func (g *Gitea) CreateIssueComment(ctx context.Context, accessToken, fullName string, number int, body string) (*IssueComment, error) {
	type req struct {
		Body string `json:"body"`
	}
	row, err := giteaDoJSON[struct {
		ID        int64  `json:"id"`
		Body      string `json:"body"`
		HTMLURL   string `json:"html_url"`
		CreatedAt string `json:"created_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
	}](ctx, accessToken, http.MethodPost,
		fmt.Sprintf("%s/api/v1/repos/%s/issues/%d/comments", g.instanceURL, fullName, number),
		req{Body: body})
	if err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, row.CreatedAt)
	return &IssueComment{
		ProviderID: row.ID, Body: row.Body, AuthorHandle: row.User.Login,
		HTMLURL: row.HTMLURL, CreatedAt: t,
	}, nil
}

func (g *Gitea) SetIssueLabels(ctx context.Context, accessToken, fullName string, number int, labels []string) error {
	if labels == nil {
		labels = []string{}
	}
	// Gitea expects label IDs, not names. Resolve in two steps.
	all, err := g.ListLabels(ctx, accessToken, fullName)
	if err != nil {
		return err
	}
	idByName, err := giteaDoJSON[[]struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}](ctx, accessToken, http.MethodGet, g.instanceURL+"/api/v1/repos/"+fullName+"/labels?limit=50", nil)
	if err != nil {
		return err
	}
	_ = all
	want := map[string]int64{}
	for _, l := range idByName {
		want[l.Name] = l.ID
	}
	ids := make([]int64, 0, len(labels))
	for _, n := range labels {
		if id, ok := want[n]; ok {
			ids = append(ids, id)
		}
	}
	type req struct {
		Labels []int64 `json:"labels"`
	}
	_, err = giteaDoJSON[any](ctx, accessToken, http.MethodPut,
		fmt.Sprintf("%s/api/v1/repos/%s/issues/%d/labels", g.instanceURL, fullName, number),
		req{Labels: ids})
	return err
}

func (g *Gitea) CloseIssue(ctx context.Context, accessToken, fullName string, number int) error {
	type req struct {
		State string `json:"state"`
	}
	_, err := giteaDoJSON[any](ctx, accessToken, http.MethodPatch,
		fmt.Sprintf("%s/api/v1/repos/%s/issues/%d", g.instanceURL, fullName, number),
		req{State: "closed"})
	return err
}

// giteaDoJSON is a Gitea-flavored variant of the GitHub helper. Differs in
// the Accept header (plain JSON; Gitea ignores vendor types).
func giteaDoJSON[T any](ctx context.Context, accessToken, method, urlStr string, body any) (T, error) {
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
	req.Header.Set("Accept", "application/json")
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

// TriggerCIWorkflow dispatches a Gitea Actions workflow. Gitea mirrors the
// GitHub Actions API, so this is essentially the GitHub call against a
// different host.
func (g *Gitea) TriggerCIWorkflow(ctx context.Context, accessToken, fullName, workflow, branch string, inputs map[string]string) (*CIDispatchResult, error) {
	payload := map[string]any{"ref": branch}
	if len(inputs) > 0 {
		payload["inputs"] = inputs
	}
	if _, err := giteaDoJSON[any](ctx, accessToken, http.MethodPost,
		fmt.Sprintf("%s/api/v1/repos/%s/actions/workflows/%s/dispatches",
			g.instanceURL, fullName, url.PathEscape(workflow)),
		payload); err != nil {
		return nil, err
	}
	return &CIDispatchResult{
		URL: fmt.Sprintf("%s/%s/actions", g.instanceURL, fullName),
	}, nil
}

func (g *Gitea) ListPipelineRuns(ctx context.Context, accessToken, fullName string, limit int) ([]PipelineRunInfo, error) {
	type giteaRun struct {
		ID         int64      `json:"id"`
		Status     string     `json:"status"`
		Conclusion string     `json:"conclusion"`
		HeadBranch string     `json:"head_branch"`
		HeadSHA    string     `json:"head_sha"`
		Name       string     `json:"name"`
		HTMLURL    string     `json:"html_url"`
		RunStarted *time.Time `json:"run_started_at"`
		UpdatedAt  *time.Time `json:"updated_at"`
	}
	type giteaResp struct {
		WorkflowRuns []giteaRun `json:"workflow_runs"`
	}

	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid fullName: %s", fullName)
	}

	resp, err := giteaDoJSON[giteaResp](ctx, accessToken, http.MethodGet,
		fmt.Sprintf("%s/api/v1/repos/%s/%s/actions/runs?limit=%d",
			g.instanceURL, parts[0], parts[1], limit), nil)
	if err != nil {
		return nil, err
	}

	out := make([]PipelineRunInfo, 0, len(resp.WorkflowRuns))
	for _, r := range resp.WorkflowRuns {
		status := "pending"
		switch {
		case r.Status == "completed" && r.Conclusion == "success":
			status = "success"
		case r.Status == "completed" && (r.Conclusion == "failure" || r.Conclusion == "timed_out"):
			status = "failure"
		case r.Status == "completed" && r.Conclusion == "cancelled":
			status = "cancelled"
		case r.Status == "in_progress":
			status = "running"
		}

		var finished *time.Time
		if r.Status == "completed" {
			finished = r.UpdatedAt
		}

		out = append(out, PipelineRunInfo{
			ProviderRunID: fmt.Sprintf("%d", r.ID),
			Status:        status,
			Branch:        r.HeadBranch,
			CommitSHA:     r.HeadSHA,
			WorkflowName:  r.Name,
			HTMLURL:       r.HTMLURL,
			StartedAt:     r.RunStarted,
			FinishedAt:    finished,
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
