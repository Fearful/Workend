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
func Clone(ctx context.Context, dc *wdagger.Client, gitURL, branch, destPath string) (*CloneResult, error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return nil, err
	}

	gitRepo := client.Git(gitURL)

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
