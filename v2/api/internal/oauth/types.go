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
}

// Token is what ExchangeCode and RefreshAccess return.
type Token struct {
	Access    string
	Refresh   string     // empty if provider doesn't support refresh
	ExpiresAt *time.Time // nil if token doesn't expire
	Scopes    string     // raw scope string from provider
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
