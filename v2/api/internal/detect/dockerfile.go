package detect

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

type Dockerfile struct{}

func (Dockerfile) Name() string { return "dockerfile" }

// Detect checks for a top-level Dockerfile. Each found dockerfile becomes a
// task with name "build" (or "build:<dirname>" for non-root locations later).
// raw_command is "docker build ." conceptually; the runner recognizes the
// "dockerfile" source and routes to a Dagger BuildKit pipeline instead.
func (Dockerfile) Detect(_ context.Context, repoPath string) ([]Task, error) {
	dfPath := filepath.Join(repoPath, "Dockerfile")
	info, err := os.Stat(dfPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, nil
	}
	return []Task{{
		Source:     "dockerfile",
		Name:       "build",
		RawCommand: "Dockerfile",
	}}, nil
}
