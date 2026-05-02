// Package repo handles git repository ingestion via Dagger pipelines.
package repo

import (
	"context"
	"fmt"
	"strings"

	wdagger "workend/api/internal/dagger"
)

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
func Clone(ctx context.Context, dc *wdagger.Client, gitURL, branch, authToken, destPath string) (*CloneResult, error) {
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
