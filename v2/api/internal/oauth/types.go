// Package oauth implements third-party OAuth flows for git hosting providers.
//
// Supports multiple instances of each provider type — e.g., one connection
// to gitlab.com and another to a self-hosted gitlab.acme.com is two
// separate Provider entries in the registry.
package oauth

import (
	"context"
	"net/url"
	"strings"
	"time"
)

// Provider is the per-instance OAuth + git-host adapter.
//
// Implementations live in github.go, gitlab.go, gitea.go. The interface
// captures only the per-vendor differences (URL shapes, header conventions,
// response field names) — the storage and HTTP-handler layers are vendor-
// agnostic and live in registry.go / handlers.go.
type Provider interface {
	// Kind is the lower-case provider family: "github" | "gitlab" | "gitea".
	Kind() string

	// ID is a stable per-instance identifier used in URL paths and DB rows.
	// e.g., "github-github-com", "gitlab-gitlab-com", "gitlab-gitlab-acme-com".
	ID() string

	// InstanceURL is the canonical base URL of this instance.
	// e.g., "https://github.com" or "https://gitlab.acme.com"
	InstanceURL() string

	// InstanceHost is the bare hostname, used for clone-URL routing.
	InstanceHost() string

	// AuthorizeURL returns the URL to redirect the user to to begin OAuth.
	AuthorizeURL(state, redirectURL string) string

	// ExchangeCode trades an authorization code for an access token.
	ExchangeCode(ctx context.Context, code, redirectURL string) (*Token, error)

	// RefreshAccess exchanges a refresh token for a new access token.
	// Returns ErrRefreshUnsupported for providers that don't issue refresh
	// tokens by default (e.g., GitHub OAuth Apps).
	RefreshAccess(ctx context.Context, refreshToken string) (*Token, error)

	// FetchHandle returns the provider-side username for the user behind
	// the access token. Used to label the connection in the UI.
	FetchHandle(ctx context.Context, accessToken string) (string, error)

	// InjectCloneAuth rewrites a clone URL to embed the access token in
	// whatever shape this provider expects. Returns the URL unchanged if
	// it isn't an HTTP(S) URL on this provider's host.
	InjectCloneAuth(rawURL, accessToken string) string

	// ListRepos returns repos accessible to the authenticated user, sorted
	// by recent activity. Page is 1-indexed; perPage capped per provider.
	ListRepos(ctx context.Context, accessToken string, page, perPage int) ([]Repo, error)

	// ListBranches returns the branches of a repo. fullName is the
	// provider-specific identifier ("owner/repo" for GitHub/Gitea,
	// "group/subgroup/project" for GitLab).
	ListBranches(ctx context.Context, accessToken, fullName string) ([]Branch, error)

	// CreatePullRequest opens a PR/MR from sourceBranch into targetBranch
	// on the given repo. Returns the public URL and provider-side number.
	CreatePullRequest(ctx context.Context, accessToken, fullName string, in PullRequestInput) (*PullRequestResult, error)

	// ListLabels returns the upstream label set for the repo. Used at
	// board first-sync to populate column choices.
	ListLabels(ctx context.Context, accessToken, fullName string) ([]Label, error)

	// ListIssues returns issues for the repo. Excludes pull requests.
	// Both open and closed are returned so the board can show closed swim
	// lanes; the caller paginates if necessary.
	ListIssues(ctx context.Context, accessToken, fullName string) ([]Issue, error)

	// ListIssueComments returns the comment thread for one issue.
	ListIssueComments(ctx context.Context, accessToken, fullName string, number int) ([]IssueComment, error)

	// CreateIssueComment posts a new comment.
	CreateIssueComment(ctx context.Context, accessToken, fullName string, number int, body string) (*IssueComment, error)

	// SetIssueLabels replaces the label set on an issue (the drag/drop
	// destination column dictates the full label list).
	SetIssueLabels(ctx context.Context, accessToken, fullName string, number int, labels []string) error

	// CloseIssue marks an issue as closed.
	CloseIssue(ctx context.Context, accessToken, fullName string, number int) error

	// TriggerCIWorkflow kicks off the upstream CI for a branch. workflow is
	// the provider-specific identifier (filename for GitHub, ref for GitLab,
	// pipeline name for Gitea Actions). Returns a public URL the user can
	// click to follow the run.
	TriggerCIWorkflow(ctx context.Context, accessToken, fullName, workflow, branch string, inputs map[string]string) (*CIDispatchResult, error)

	// ListPipelineRuns returns recent CI/CD pipeline runs for a repo.
	ListPipelineRuns(ctx context.Context, accessToken, fullName string, limit int) ([]PipelineRunInfo, error)
}

// CIDispatchResult is what the provider returned after triggering CI.
type CIDispatchResult struct {
	URL string `json:"url"` // best-effort link to the workflow runs page
}

// PipelineRunInfo is a normalized CI/CD run across providers.
type PipelineRunInfo struct {
	ProviderRunID string     `json:"provider_run_id"`
	Status        string     `json:"status"` // "success" | "failure" | "running" | "pending" | "cancelled"
	Branch        string     `json:"branch"`
	CommitSHA     string     `json:"commit_sha"`
	WorkflowName  string     `json:"workflow_name"`
	HTMLURL       string     `json:"html_url"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
}

// Label is a normalized upstream label.
type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Issue is the normalized cross-provider issue.
type Issue struct {
	Number          int        `json:"number"`
	Title           string     `json:"title"`
	Body            string     `json:"body"`
	State           string     `json:"state"` // 'open' | 'closed'
	Labels          []string   `json:"labels"`
	AuthorHandle    string     `json:"author_handle"`
	AuthorURL       string     `json:"author_url"`
	HTMLURL         string     `json:"html_url"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

// IssueComment is one comment on an issue.
type IssueComment struct {
	ProviderID   int64     `json:"provider_id"`
	Body         string    `json:"body"`
	AuthorHandle string    `json:"author_handle"`
	HTMLURL      string    `json:"html_url"`
	CreatedAt    time.Time `json:"created_at"`
}

// Branch is the normalized cross-provider branch summary.
type Branch struct {
	Name      string `json:"name"`
	CommitSHA string `json:"commit_sha"`
	Protected bool   `json:"protected"`
	Default   bool   `json:"default"`
}

// PullRequestInput is the user-supplied content for an outgoing PR/MR.
type PullRequestInput struct {
	Source string // source branch, e.g. "feature/x"
	Target string // target branch, e.g. "main"
	Title  string
	Body   string
}

// PullRequestResult is what the provider returned after creation.
type PullRequestResult struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
}

// Token is what ExchangeCode and RefreshAccess return.
type Token struct {
	Access    string
	Refresh   string     // empty if provider doesn't support refresh
	ExpiresAt *time.Time // nil if token doesn't expire
	Scopes    string     // raw scope string from provider
}

// Repo is the normalized cross-provider repo summary returned from
// Provider.ListRepos. Mapped from GitHub/GitLab/Gitea response shapes.
type Repo struct {
	Name          string     `json:"name"`           // e.g. "kit"
	FullName      string     `json:"full_name"`      // e.g. "sveltejs/kit"
	Description   string     `json:"description"`
	Private       bool       `json:"private"`
	HTMLURL       string     `json:"html_url"`       // browser link
	CloneURL      string     `json:"clone_url"`      // https URL for git clone
	DefaultBranch string     `json:"default_branch"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

// OwnsURL is a default implementation: equality match on URL host.
// Per Stage 15 decision (host equality, not prefix or fuzzy match).
func OwnsURL(p Provider, rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	// strip port if any — match "gitlab.com" against "gitlab.com:8080"
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	want := strings.ToLower(p.InstanceHost())
	if i := strings.IndexByte(want, ':'); i >= 0 {
		want = want[:i]
	}
	return host == want
}

// MakeID derives a stable provider ID from kind + instance host.
// "gitlab" + "gitlab.acme.com" -> "gitlab-gitlab-acme-com"
func MakeID(kind, instanceHost string) string {
	clean := strings.ToLower(instanceHost)
	clean = strings.ReplaceAll(clean, ".", "-")
	clean = strings.ReplaceAll(clean, ":", "-")
	return kind + "-" + clean
}
