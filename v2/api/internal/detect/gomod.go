package detect

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type GoMod struct{}

func (GoMod) Name() string { return "go" }

func (GoMod) Detect(_ context.Context, repoPath string) ([]Task, error) {
	modPath := filepath.Join(repoPath, "go.mod")
	if _, err := os.Stat(modPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	static := []struct {
		name string
		cmd  string
	}{
		{"build", "go build ./..."},
		{"test", "go test ./..."},
		{"vet", "go vet ./..."},
		{"fmt", "go fmt ./..."},
	}
	tasks := make([]Task, 0, len(static))
	for _, s := range static {
		tasks = append(tasks, Task{
			Source:     "go",
			Name:       s.name,
			RawCommand: s.cmd,
		})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}
