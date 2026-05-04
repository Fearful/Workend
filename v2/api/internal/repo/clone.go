// Package repo handles git repository ingestion via Dagger pipelines.
package repo

import (
	"context"
	"fmt"
	"strings"

	wdagger "workend/api/internal/dagger"
)

// SSHKey holds the materialized SSH key the caller will pass when the URL
// is git+SSH. PrivateKey is PEM bytes; KnownHosts may be empty (we use
// StrictHostKeyChecking=accept-new in that case).
type SSHKey struct {
	PrivateKey string
	KnownHosts string
}

type CloneResult struct {
	LocalPath        string
	DefaultBranch    string
	LatestCommitSHA  string
	LatestCommitMsg  string
	LatestCommitAuth string
}

// Clone fetches `gitURL` at `branch` (or HEAD if empty) and exports the tree
// to `destPath` on the API container's filesystem. Then runs git log inside
// an alpine/git container against the cloned tree to extract commit metadata.
//
// If `authToken` is non-empty, it's embedded into the URL as
// https://x-access-token:<token>@host/path so private repos work.
//
// If `sshKey` is non-nil, the URL is assumed to be git+SSH and the key
// drives a sidecar `git clone` rather than the Dagger Git API (which
// expects an SSH agent socket; see CloneSSH below).
func Clone(ctx context.Context, dc *wdagger.Client, gitURL, branch, authToken string, sshKey *SSHKey, destPath string) (*CloneResult, error) {
	if sshKey != nil {
		return CloneSSH(ctx, dc, gitURL, branch, sshKey, destPath)
	}

	client, err := dc.Get(ctx)
	if err != nil {
		return nil, err
	}

	cloneURL := gitURL
	if authToken != "" {
		cloneURL = injectAuth(gitURL, authToken)
	}

	gitRepo := client.Git(cloneURL)

	var ref *commitRef
	if branch != "" {
		ref = &commitRef{branchRef: gitRepo.Branch(branch)}
	} else {
		// No branch specified — let Dagger pick the default
		ref = &commitRef{branchRef: gitRepo.Head()}
	}

	tree := ref.tree()
	if _, err := tree.Export(ctx, destPath); err != nil {
		return nil, fmt.Errorf("export tree: %w", err)
	}

	commitSHA, err := ref.commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("read commit sha: %w", err)
	}

	// Read commit message + author by running git log inside a container
	// that has the cloned tree mounted. This avoids needing git on the
	// API container.
	logOut, err := client.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/repo", tree).
		WithWorkdir("/repo").
		WithExec([]string{"git", "log", "-1", "--format=%an%x1f%s"}).
		Stdout(ctx)
	if err != nil {
		// Don't fail the whole clone if log read fails; it's nice-to-have
		logOut = ""
	}

	author, msg := parseGitLog(logOut)

	resolvedBranch := branch
	if resolvedBranch == "" {
		// Best-effort: ask the container what HEAD points to
		headOut, err := client.Container().
			From("alpine/git:latest").
			WithMountedDirectory("/repo", tree).
			WithWorkdir("/repo").
			WithExec([]string{"git", "symbolic-ref", "--short", "HEAD"}).
			Stdout(ctx)
		if err == nil {
			resolvedBranch = strings.TrimSpace(headOut)
		}
	}

	return &CloneResult{
		LocalPath:        destPath,
		DefaultBranch:    resolvedBranch,
		LatestCommitSHA:  strings.TrimSpace(commitSHA),
		LatestCommitMsg:  msg,
		LatestCommitAuth: author,
	}, nil
}

func parseGitLog(out string) (author, message string) {
	out = strings.TrimSpace(out)
	if out == "" {
		return "", ""
	}
	parts := strings.SplitN(out, "\x1f", 2)
	if len(parts) != 2 {
		return "", out
	}
	return parts[0], parts[1]
}

// CloneSSH performs a git clone over SSH using a sidecar alpine/git
// container with the user's private key mounted as a Dagger secret. Used
// instead of the Dagger Git API because Git over SSH inside the engine
// requires an SSH agent socket, which we'd have to pre-provision per
// session — the sidecar approach is simpler and equally secure.
func CloneSSH(ctx context.Context, dc *wdagger.Client, gitURL, branch string, sshKey *SSHKey, destPath string) (*CloneResult, error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return nil, err
	}
	keySecret := client.SetSecret("workend-ssh-key", sshKey.PrivateKey)

	// Build a one-line ssh command: identity from the mounted secret, no
	// host-key prompts. accept-new persists known_hosts in-container so
	// repeat fetches in this run reuse the trusted key.
	sshCmd := "ssh -i /tmp/ssh-key " +
		"-o IdentitiesOnly=yes " +
		"-o StrictHostKeyChecking=accept-new " +
		"-o UserKnownHostsFile=/tmp/known_hosts"

	args := []string{"git", "clone"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, gitURL, "/work")

	container := client.Container().
		From("alpine/git:latest").
		WithExec([]string{"sh", "-c", "mkdir -p /tmp && touch /tmp/known_hosts && chmod 600 /tmp/known_hosts"}).
		WithMountedSecret("/tmp/ssh-key", keySecret).
		WithEnvVariable("GIT_SSH_COMMAND", sshCmd).
		WithExec(args)

	tree := container.Directory("/work")
	if _, err := tree.Export(ctx, destPath); err != nil {
		return nil, fmt.Errorf("export tree: %w", err)
	}

	commitSHA, err := container.
		WithWorkdir("/work").
		WithExec([]string{"git", "rev-parse", "HEAD"}).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("read commit sha: %w", err)
	}

	logOut, err := container.
		WithWorkdir("/work").
		WithExec([]string{"git", "log", "-1", "--format=%an%x1f%s"}).
		Stdout(ctx)
	if err != nil {
		logOut = ""
	}
	author, msg := parseGitLog(logOut)

	resolvedBranch := branch
	if resolvedBranch == "" {
		headOut, err := container.
			WithWorkdir("/work").
			WithExec([]string{"git", "symbolic-ref", "--short", "HEAD"}).
			Stdout(ctx)
		if err == nil {
			resolvedBranch = strings.TrimSpace(headOut)
		}
	}

	return &CloneResult{
		LocalPath:        destPath,
		DefaultBranch:    resolvedBranch,
		LatestCommitSHA:  strings.TrimSpace(commitSHA),
		LatestCommitMsg:  msg,
		LatestCommitAuth: author,
	}, nil
}

// LsRemoteBranches enumerates branch names of a remote without doing a full
// clone. Runs `git ls-remote --heads` inside an alpine/git container; works
// for any reachable repo (public or token-injected). Returns refs sorted by
// name with their commit SHAs.
func LsRemoteBranches(ctx context.Context, dc *wdagger.Client, gitURL, authToken string) ([]RemoteBranch, error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return nil, err
	}
	url := gitURL
	if authToken != "" {
		url = injectAuth(gitURL, authToken)
	}
	out, err := client.Container().
		From("alpine/git:latest").
		WithExec([]string{"git", "ls-remote", "--heads", url}).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("git ls-remote: %w", err)
	}
	var branches []RemoteBranch
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		// Format: "<sha>\trefs/heads/<branch>"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		ref := strings.TrimPrefix(parts[1], "refs/heads/")
		branches = append(branches, RemoteBranch{Name: ref, CommitSHA: parts[0]})
	}
	return branches, nil
}

// RemoteBranch is the minimal representation returned by LsRemoteBranches.
// Used as the fallback when no OAuth provider is connected.
type RemoteBranch struct {
	Name      string `json:"name"`
	CommitSHA string `json:"commit_sha"`
}

// LsRemoteBranchSHA resolves a single branch name to its commit SHA via
// `git ls-remote`. Faster than fetching the full branch list when we only
// care about one branch (per-branch run resolution path).
func LsRemoteBranchSHA(ctx context.Context, dc *wdagger.Client, gitURL, authToken, branch string) (string, error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return "", err
	}
	url := gitURL
	if authToken != "" {
		url = injectAuth(gitURL, authToken)
	}
	out, err := client.Container().
		From("alpine/git:latest").
		WithExec([]string{"git", "ls-remote", url, "refs/heads/" + branch}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("git ls-remote: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		return parts[0], nil
	}
	return "", fmt.Errorf("branch %q not found", branch)
}

// injectAuth turns "https://github.com/foo/bar.git" into
// "https://x-access-token:<token>@github.com/foo/bar.git". A simple,
// well-supported way to authenticate HTTPS git clones without needing
// Dagger Secret bindings.
func injectAuth(rawURL, token string) string {
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(rawURL, scheme) {
			return scheme + "x-access-token:" + token + "@" + strings.TrimPrefix(rawURL, scheme)
		}
	}
	return rawURL
}
