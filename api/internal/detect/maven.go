package detect

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type Maven struct{}

func (Maven) Name() string { return "maven" }

func (Maven) Detect(_ context.Context, repoPath string) ([]Task, error) {
	pomPath := filepath.Join(repoPath, "pom.xml")
	if _, err := os.Stat(pomPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	cmd := "mvn"
	if isExecutable(filepath.Join(repoPath, "mvnw")) {
		cmd = "./mvnw"
	}

	names := []string{"clean", "compile", "test", "package", "install", "deploy", "verify", "site"}
	tasks := make([]Task, 0, len(names))
	for _, n := range names {
		tasks = append(tasks, Task{
			Source:     "maven",
			Name:       n,
			RawCommand: cmd + " " + shellQuote(n),
		})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}
