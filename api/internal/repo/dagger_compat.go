package repo

import (
	"context"

	dag "dagger.io/dagger"
)

// commitRef is a thin wrapper that lets us treat both Branch() and Head() the
// same way — both expose Tree() and Commit(). Dagger's GitRef interface
// changed shape across versions; isolating the calls here makes future SDK
// bumps a single-file change.
type commitRef struct {
	branchRef *dag.GitRef
}

func (c *commitRef) tree() *dag.Directory {
	return c.branchRef.Tree()
}

func (c *commitRef) commit(ctx context.Context) (string, error) {
	return c.branchRef.Commit(ctx)
}
