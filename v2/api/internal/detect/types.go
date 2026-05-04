// Package detect discovers runnable tasks in a cloned repository.
//
// A Detector inspects a repo path on disk (which the api container can read
// because /repos is a shared volume) and returns a list of detected Tasks.
// Detection happens once per clone, results are persisted in the tasks table.
package detect

import "context"

type Task struct {
	Source     string // "npm" | "pnpm" | "yarn" | "bun" | "just" | "dockerfile" | "dagger" | custom
	Name       string
	RawCommand string // verbatim command string the runner will execute
}

type Detector interface {
	// Name identifies this detector in logs.
	Name() string

	// Detect inspects repoPath and returns discovered tasks.
	// Returns (nil, nil) if the detector's manifest is absent (not an error).
	Detect(ctx context.Context, repoPath string) ([]Task, error)
}
