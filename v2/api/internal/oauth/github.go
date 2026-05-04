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

func (g *GitHub) ListRepos(ctx context.Context, accessToken string, page, perPage int) ([]Repo, error) {
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}
	if page < 1 {
		page = 1
	}
	q := url.Values{}
	q.Set("per_page", fmt.Sprintf("%d", perPage))
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("sort", "updated")
	q.Set("affiliation", "owner,collaborator,organization_member")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.apiBase()+"/user/repos?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github /user/repos status %d: %s", resp.StatusCode, string(raw))
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

func (g *GitHub) ListBranches(ctx context.Context, accessToken, fullName string) ([]Branch, error) {
	// Fetch the repo first so we know which branch is "default" — the
	// branches endpoint doesn't surface that flag.
	defaultBranch := ""
	{
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.apiBase()+"/repos/"+fullName, nil)
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Accept", "application/vnd.github+json")
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.apiBase()+"/repos/"+fullName+"/branches?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github /branches status %d: %s", resp.StatusCode, string(raw))
	}
	var rows []struct {
		Name      string `json:"name"`
		Protected bool   `json:"protected"`
		Commit    struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	out := make([]Branch, 0, len(rows))
	for _, r := range rows {
		out = append(out, Branch{
			Name:      r.Name,
			CommitSHA: r.Commit.SHA,
			Protected: r.Protected,
			Default:   r.Name == defaultBranch,
		})
	}
	return out, nil
}

func (g *GitHub) CreatePullRequest(ctx context.Context, accessToken, fullName string, in PullRequestInput) (*PullRequestResult, error) {
	payload := map[string]any{
		"head":  in.Source,
		"base":  in.Target,
		"title": in.Title,
		"body":  in.Body,
	}
	buf, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.apiBase()+"/repos/"+fullName+"/pulls", strings.NewReader(string(buf)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("github create PR status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
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

func (g *GitHub) ListLabels(ctx context.Context, accessToken, fullName string) ([]Label, error) {
	q := url.Values{}
	q.Set("per_page", "100")
	out, err := ghDoJSON[[]struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}](ctx, accessToken, http.MethodGet, g.apiBase()+"/repos/"+fullName+"/labels?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	res := make([]Label, 0, len(out))
	for _, l := range out {
		res = append(res, Label{Name: l.Name, Color: l.Color})
	}
	return res, nil
}

func (g *GitHub) ListIssues(ctx context.Context, accessToken, fullName string) ([]Issue, error) {
	q := url.Values{}
	q.Set("per_page", "100")
	q.Set("state", "all")
	q.Set("sort", "updated")
	rows, err := ghDoJSON[[]struct {
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
		PullRequest *struct{} `json:"pull_request"`
	}](ctx, accessToken, http.MethodGet, g.apiBase()+"/repos/"+fullName+"/issues?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	out := make([]Issue, 0, len(rows))
	for _, r := range rows {
		// `pull_request` is non-null when the row is actually a PR; skip those.
		if r.PullRequest != nil {
			continue
		}
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

func (g *GitHub) ListIssueComments(ctx context.Context, accessToken, fullName string, number int) ([]IssueComment, error) {
	q := url.Values{}
	q.Set("per_page", "100")
	rows, err := ghDoJSON[[]struct {
		ID        int64  `json:"id"`
		Body      string `json:"body"`
		HTMLURL   string `json:"html_url"`
		CreatedAt string `json:"created_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
	}](ctx, accessToken, http.MethodGet,
		fmt.Sprintf("%s/repos/%s/issues/%d/comments?%s", g.apiBase(), fullName, number, q.Encode()), nil)
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

func (g *GitHub) CreateIssueComment(ctx context.Context, accessToken, fullName string, number int, body string) (*IssueComment, error) {
	type req struct {
		Body string `json:"body"`
	}
	row, err := ghDoJSON[struct {
		ID        int64  `json:"id"`
		Body      string `json:"body"`
		HTMLURL   string `json:"html_url"`
		CreatedAt string `json:"created_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
	}](ctx, accessToken, http.MethodPost,
		fmt.Sprintf("%s/repos/%s/issues/%d/comments", g.apiBase(), fullName, number),
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

func (g *GitHub) SetIssueLabels(ctx context.Context, accessToken, fullName string, number int, labels []string) error {
	if labels == nil {
		labels = []string{}
	}
	type req struct {
		Labels []string `json:"labels"`
	}
	_, err := ghDoJSON[any](ctx, accessToken, http.MethodPut,
		fmt.Sprintf("%s/repos/%s/issues/%d/labels", g.apiBase(), fullName, number),
		req{Labels: labels})
	return err
}

func (g *GitHub) CloseIssue(ctx context.Context, accessToken, fullName string, number int) error {
	type req struct {
		State string `json:"state"`
	}
	_, err := ghDoJSON[any](ctx, accessToken, http.MethodPatch,
		fmt.Sprintf("%s/repos/%s/issues/%d", g.apiBase(), fullName, number),
		req{State: "closed"})
	return err
}

// ghDoJSON is a generic JSON request helper for the GitHub-shaped APIs.
// Reused by Gitea (which mirrors the same routes & headers).
func ghDoJSON[T any](ctx context.Context, accessToken, method, urlStr string, body any) (T, error) {
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
	req.Header.Set("Accept", "application/vnd.github+json")
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

// TriggerCIWorkflow dispatches a workflow_dispatch event. `workflow` is the
// workflow filename (e.g. `ci.yml`); `branch` is the ref to run against.
//
// GitHub returns 204 with no body on success — the user clicks through to
// /actions to see their run.
func (g *GitHub) TriggerCIWorkflow(ctx context.Context, accessToken, fullName, workflow, branch string, inputs map[string]string) (*CIDispatchResult, error) {
	payload := map[string]any{"ref": branch}
	if len(inputs) > 0 {
		payload["inputs"] = inputs
	}
	if _, err := ghDoJSON[any](ctx, accessToken, http.MethodPost,
		fmt.Sprintf("%s/repos/%s/actions/workflows/%s/dispatches",
			g.apiBase(), fullName, url.PathEscape(workflow)),
		payload); err != nil {
		return nil, err
	}
	return &CIDispatchResult{
		URL: fmt.Sprintf("https://%s/%s/actions/workflows/%s",
			g.instanceHost, fullName, url.PathEscape(workflow)),
	}, nil
}

func (g *GitHub) ListPipelineRuns(ctx context.Context, accessToken, fullName string, limit int) ([]PipelineRunInfo, error) {
	type ghRun struct {
		ID           int64      `json:"id"`
		Status       string     `json:"status"`
		Conclusion   *string    `json:"conclusion"`
		HeadBranch   string     `json:"head_branch"`
		HeadSHA      string     `json:"head_sha"`
		Name         string     `json:"name"`
		HTMLURL      string     `json:"html_url"`
		RunStartedAt *time.Time `json:"run_started_at"`
		UpdatedAt    *time.Time `json:"updated_at"`
	}
	type ghResp struct {
		WorkflowRuns []ghRun `json:"workflow_runs"`
	}

	resp, err := ghDoJSON[ghResp](ctx, accessToken, http.MethodGet,
		fmt.Sprintf("%s/repos/%s/actions/runs?per_page=%d", g.apiBase(), fullName, limit), nil)
	if err != nil {
		return nil, err
	}

	out := make([]PipelineRunInfo, 0, len(resp.WorkflowRuns))
	for _, r := range resp.WorkflowRuns {
		status := "pending"
		if r.Status == "completed" && r.Conclusion != nil {
			switch *r.Conclusion {
			case "success":
				status = "success"
			case "failure", "timed_out":
				status = "failure"
			case "cancelled":
				status = "cancelled"
			default:
				status = *r.Conclusion
			}
		} else if r.Status == "in_progress" {
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
			StartedAt:     r.RunStartedAt,
			FinishedAt:    finished,
		})
	}
	return out, nil
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
