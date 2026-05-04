package detect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type Compose struct{}

func (Compose) Name() string { return "compose" }

func (Compose) Detect(_ context.Context, repoPath string) ([]Task, error) {
	path, ok := findComposeFile(repoPath)
	if !ok {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read compose file: %w", err)
	}

	var doc struct {
		Services map[string]any `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse compose file: %w", err)
	}

	tasks := make([]Task, 0, len(doc.Services))
	for name := range doc.Services {
		tasks = append(tasks, Task{
			Source:     "compose",
			Name:       name,
			RawCommand: "docker compose up " + shellQuote(name),
		})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}

func findComposeFile(repoPath string) (string, bool) {
	for _, name := range []string{"compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml"} {
		p := filepath.Join(repoPath, name)
		if _, err := os.Stat(p); err == nil {
			return p, true
		} else if !errors.Is(err, os.ErrNotExist) {
			continue
		}
	}
	return "", false
}
