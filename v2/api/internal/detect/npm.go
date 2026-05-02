package detect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type NPM struct{}

func (NPM) Name() string { return "npm" }

func (NPM) Detect(_ context.Context, repoPath string) ([]Task, error) {
	pkgPath := filepath.Join(repoPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read package.json: %w", err)
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	tasks := make([]Task, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		tasks = append(tasks, Task{
			Source:     "npm",
			Name:       name,
			RawCommand: "npm run " + shellQuote(name),
		})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}

// shellQuote wraps `s` in single quotes if it contains shell-significant
// characters. Avoids surprises with names like "test:unit".
func shellQuote(s string) string {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		// has a special char — quote
		return "'" + s + "'"
	}
	return s
}
