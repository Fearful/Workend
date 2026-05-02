package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/secret"
)

var (
	ErrUnknownProvider     = errors.New("unknown provider")
	ErrRefreshUnsupported  = errors.New("provider does not support refresh tokens")
	ErrConnectionNotFound  = errors.New("connection not found")
)

// Registry holds all configured Provider instances.
type Registry struct {
	providers []Provider
	byID      map[string]Provider
	box       *secret.Box
	pool      *pgxpool.Pool
}

func NewRegistry(pool *pgxpool.Pool, box *secret.Box, providers []Provider) *Registry {
	r := &Registry{
		providers: providers,
		byID:      map[string]Provider{},
		box:       box,
		pool:      pool,
	}
	for _, p := range providers {
		r.byID[p.ID()] = p
	}
	return r
}

func (r *Registry) All() []Provider { return r.providers }

func (r *Registry) ByID(id string) (Provider, bool) {
	p, ok := r.byID[id]
	return p, ok
}

// ForCloneURL returns the first provider whose host matches the URL.
// Per Stage 15 decision: host equality, not prefix.
func (r *Registry) ForCloneURL(rawURL string) (Provider, bool) {
	for _, p := range r.providers {
		if OwnsURL(p, rawURL) {
			return p, true
		}
	}
	return nil, false
}

// Connection is a stored OAuth connection.
type Connection struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	ProviderID  string     `json:"provider_id"`  // synthesized from provider+instance_url
	Provider    string     `json:"provider"`     // 'github' | 'gitlab' | 'gitea'
	InstanceURL string     `json:"instance_url"`
	Handle      string     `json:"handle"`
	Scopes      string     `json:"scopes"`
	ExpiresAt   *time.Time `json:"expires_at"`
	ConnectedAt time.Time  `json:"connected_at"`
}

// Save inserts/updates a connection for (user, provider kind, instance_url).
func (r *Registry) Save(ctx context.Context, userID uuid.UUID, p Provider, tok *Token, handle string) error {
	encAccess, err := r.box.Seal([]byte(tok.Access))
	if err != nil {
		return fmt.Errorf("seal access: %w", err)
	}
	var encRefresh []byte
	if tok.Refresh != "" {
		encRefresh, err = r.box.Seal([]byte(tok.Refresh))
		if err != nil {
			return fmt.Errorf("seal refresh: %w", err)
		}
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO provider_connections
		    (user_id, provider, instance_url, access_token, refresh_token, expires_at, scopes, handle)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, provider, instance_url) DO UPDATE
		SET access_token  = EXCLUDED.access_token,
		    refresh_token = EXCLUDED.refresh_token,
		    expires_at    = EXCLUDED.expires_at,
		    scopes        = EXCLUDED.scopes,
		    handle        = EXCLUDED.handle,
		    connected_at  = now()
	`, userID, p.Kind(), p.InstanceURL(), encAccess, encRefresh, tok.ExpiresAt, tok.Scopes, handle)
	return err
}

// LookupByID returns a stored connection.
func (r *Registry) LookupByID(ctx context.Context, userID uuid.UUID, connectionID uuid.UUID) (*Connection, error) {
	var c Connection
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, instance_url, COALESCE(handle, ''), COALESCE(scopes, ''),
		       expires_at, connected_at
		FROM provider_connections
		WHERE id = $1 AND user_id = $2
	`, connectionID, userID).Scan(&c.ID, &c.UserID, &c.Provider, &c.InstanceURL,
		&c.Handle, &c.Scopes, &c.ExpiresAt, &c.ConnectedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConnectionNotFound
	}
	if err != nil {
		return nil, err
	}
	c.ProviderID = MakeIDFromInstanceURL(c.Provider, c.InstanceURL)
	return &c, nil
}

// LookupByProviderForUser returns the connection for a (user, provider).
// Used by handlers that operate on a specific provider ID.
func (r *Registry) LookupByProviderForUser(ctx context.Context, userID uuid.UUID, providerID string) (*Connection, error) {
	p, ok := r.ByID(providerID)
	if !ok {
		return nil, ErrUnknownProvider
	}
	var c Connection
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, provider, instance_url, COALESCE(handle, ''), COALESCE(scopes, ''),
		       expires_at, connected_at
		FROM provider_connections
		WHERE user_id = $1 AND provider = $2 AND instance_url = $3
	`, userID, p.Kind(), p.InstanceURL()).Scan(&c.ID, &c.UserID, &c.Provider, &c.InstanceURL,
		&c.Handle, &c.Scopes, &c.ExpiresAt, &c.ConnectedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConnectionNotFound
	}
	if err != nil {
		return nil, err
	}
	c.ProviderID = providerID
	return &c, nil
}

// ListConnections returns all of a user's connections.
func (r *Registry) ListConnections(ctx context.Context, userID uuid.UUID) ([]Connection, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, provider, instance_url, COALESCE(handle, ''), COALESCE(scopes, ''),
		       expires_at, connected_at
		FROM provider_connections
		WHERE user_id = $1
		ORDER BY connected_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Connection{}
	for rows.Next() {
		var c Connection
		if err := rows.Scan(&c.ID, &c.UserID, &c.Provider, &c.InstanceURL,
			&c.Handle, &c.Scopes, &c.ExpiresAt, &c.ConnectedAt); err != nil {
			return nil, err
		}
		c.ProviderID = MakeIDFromInstanceURL(c.Provider, c.InstanceURL)
		out = append(out, c)
	}
	return out, nil
}

// Delete removes a connection.
func (r *Registry) Delete(ctx context.Context, userID, connectionID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM provider_connections WHERE id = $1 AND user_id = $2`, connectionID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

// AccessTokenForCloneURL returns the (decrypted) access token for whichever
// provider owns this URL, or "" if no matching connection exists.
//
// On expired access tokens with a refresh token available, attempts a
// refresh first (Stage 15 lazy-refresh). If refresh fails, returns "".
func (r *Registry) AccessTokenForCloneURL(ctx context.Context, userID uuid.UUID, rawURL string) string {
	p, ok := r.ForCloneURL(rawURL)
	if !ok {
		return ""
	}
	access, _, _ := r.accessTokenForProvider(ctx, userID, p)
	return access
}

// ListReposForUser returns a page of repos for a (user, provider) connection.
// Lazy-refreshes the access token if it's expired.
func (r *Registry) ListReposForUser(ctx context.Context, userID uuid.UUID, providerID string, page, perPage int) ([]Repo, error) {
	p, ok := r.ByID(providerID)
	if !ok {
		return nil, ErrUnknownProvider
	}
	access, _, err := r.accessTokenForProvider(ctx, userID, p)
	if err != nil {
		return nil, err
	}
	if access == "" {
		return nil, ErrConnectionNotFound
	}
	return p.ListRepos(ctx, access, page, perPage)
}

// accessTokenForProvider is the workhorse: load encrypted token from DB,
// decrypt, refresh if expired, return access string.
func (r *Registry) accessTokenForProvider(ctx context.Context, userID uuid.UUID, p Provider) (access string, expiresAt *time.Time, err error) {
	var encAccess, encRefresh []byte
	var dbExpiresAt *time.Time
	err = r.pool.QueryRow(ctx, `
		SELECT access_token, refresh_token, expires_at
		FROM provider_connections
		WHERE user_id = $1 AND provider = $2 AND instance_url = $3
	`, userID, p.Kind(), p.InstanceURL()).Scan(&encAccess, &encRefresh, &dbExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, ErrConnectionNotFound
	}
	if err != nil {
		return "", nil, err
	}

	plainAccess, err := r.box.Open(encAccess)
	if err != nil {
		return "", nil, fmt.Errorf("decrypt access: %w", err)
	}
	access = string(plainAccess)
	expiresAt = dbExpiresAt

	// Refresh if expired and we have a refresh token.
	needsRefresh := dbExpiresAt != nil && time.Now().After(*dbExpiresAt)
	if needsRefresh && len(encRefresh) > 0 {
		plainRefresh, err := r.box.Open(encRefresh)
		if err != nil {
			return access, expiresAt, nil // best-effort: return the (expired) token
		}
		newTok, err := p.RefreshAccess(ctx, string(plainRefresh))
		if err != nil {
			return access, expiresAt, nil // refresh failed; return what we have
		}
		// Persist the new token. Keep existing handle/scopes via no-op INSERT...DO UPDATE.
		_ = r.refreshSave(ctx, userID, p, newTok)
		return newTok.Access, newTok.ExpiresAt, nil
	}
	return access, expiresAt, nil
}

func (r *Registry) refreshSave(ctx context.Context, userID uuid.UUID, p Provider, tok *Token) error {
	encAccess, err := r.box.Seal([]byte(tok.Access))
	if err != nil {
		return err
	}
	var encRefresh []byte
	if tok.Refresh != "" {
		encRefresh, _ = r.box.Seal([]byte(tok.Refresh))
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE provider_connections
		SET access_token  = $1,
		    refresh_token = COALESCE(NULLIF($2, ''::bytea), refresh_token),
		    expires_at    = $3
		WHERE user_id = $4 AND provider = $5 AND instance_url = $6
	`, encAccess, encRefresh, tok.ExpiresAt, userID, p.Kind(), p.InstanceURL())
	return err
}

// MakeIDFromInstanceURL parses a URL and computes the provider ID.
func MakeIDFromInstanceURL(kind, instanceURL string) string {
	host := instanceURL
	if i := strings.Index(host, "://"); i >= 0 {
		host = host[i+3:]
	}
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	return MakeID(kind, host)
}

// --- Configuration loading ---

// InstanceConfig is one configured provider instance.
type InstanceConfig struct {
	Kind         string `json:"kind"`         // 'github' | 'gitlab' | 'gitea'
	InstanceURL  string `json:"instance_url"` // omitted for github (defaults github.com)
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// LoadConfig builds the Provider list from:
//  1. WORKEND_PROVIDERS_FILE — JSON array of InstanceConfig (multi-instance friendly)
//  2. Otherwise: WORKEND_GITHUB_/GITLAB_/GITEA_ env vars (one instance per kind)
//
// redirectBase is the public URL the web frontend serves at, used to build
// callback URLs. Each provider's callback URL is "{redirectBase}/auth/{providerID}/callback".
func LoadConfig(redirectBase string) ([]Provider, error) {
	if path := os.Getenv("WORKEND_PROVIDERS_FILE"); path != "" {
		return loadFromFile(path, redirectBase)
	}
	return loadFromEnv(redirectBase), nil
}

func loadFromFile(path, redirectBase string) ([]Provider, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read providers file: %w", err)
	}
	var configs []InstanceConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("parse providers file: %w", err)
	}
	out := make([]Provider, 0, len(configs))
	for _, c := range configs {
		p, err := buildProvider(c, redirectBase)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func loadFromEnv(redirectBase string) []Provider {
	var out []Provider
	if id, secret := os.Getenv("WORKEND_GITHUB_CLIENT_ID"), os.Getenv("WORKEND_GITHUB_CLIENT_SECRET"); id != "" && secret != "" {
		p, _ := buildProvider(InstanceConfig{
			Kind:        "github",
			InstanceURL: "https://github.com",
			ClientID:    id, ClientSecret: secret,
		}, redirectBase)
		if p != nil {
			out = append(out, p)
		}
	}
	if id, secret := os.Getenv("WORKEND_GITLAB_CLIENT_ID"), os.Getenv("WORKEND_GITLAB_CLIENT_SECRET"); id != "" && secret != "" {
		instance := os.Getenv("WORKEND_GITLAB_URL")
		if instance == "" {
			instance = "https://gitlab.com"
		}
		p, _ := buildProvider(InstanceConfig{
			Kind:        "gitlab",
			InstanceURL: instance,
			ClientID:    id, ClientSecret: secret,
		}, redirectBase)
		if p != nil {
			out = append(out, p)
		}
	}
	if id, secret := os.Getenv("WORKEND_GITEA_CLIENT_ID"), os.Getenv("WORKEND_GITEA_CLIENT_SECRET"); id != "" && secret != "" {
		instance := os.Getenv("WORKEND_GITEA_URL")
		if instance != "" {
			p, _ := buildProvider(InstanceConfig{
				Kind:        "gitea",
				InstanceURL: instance,
				ClientID:    id, ClientSecret: secret,
			}, redirectBase)
			if p != nil {
				out = append(out, p)
			}
		}
	}
	return out
}

func buildProvider(c InstanceConfig, redirectBase string) (Provider, error) {
	if c.ClientID == "" || c.ClientSecret == "" {
		return nil, fmt.Errorf("provider %s: client_id and client_secret required", c.Kind)
	}
	switch c.Kind {
	case "github":
		instance := c.InstanceURL
		if instance == "" {
			instance = "https://github.com"
		}
		return NewGitHub(c.ClientID, c.ClientSecret, instance, redirectBase), nil
	case "gitlab":
		if c.InstanceURL == "" {
			return nil, fmt.Errorf("gitlab requires instance_url")
		}
		return NewGitLab(c.ClientID, c.ClientSecret, c.InstanceURL, redirectBase), nil
	case "gitea":
		if c.InstanceURL == "" {
			return nil, fmt.Errorf("gitea requires instance_url")
		}
		return NewGitea(c.ClientID, c.ClientSecret, c.InstanceURL, redirectBase), nil
	default:
		return nil, fmt.Errorf("unknown provider kind: %s", c.Kind)
	}
}
