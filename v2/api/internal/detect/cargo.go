package detect

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type Cargo struct{}

func (Cargo) Name() string { return "cargo" }

func (Cargo) Detect(_ context.Context, repoPath string) ([]Task, error) {
	cargoPath := filepath.Join(repoPath, "Cargo.toml")
	if _, err := os.Stat(cargoPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	names := []string{"build", "test", "check", "run", "clippy", "fmt", "doc", "bench"}
	tasks := make([]Task, 0, len(names))
	for _, n := range names {
		tasks = append(tasks, Task{
			Source:     "cargo",
			Name:       n,
			RawCommand: "cargo " + shellQuote(n),
		})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}
