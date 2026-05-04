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

type Composer struct{}

func (Composer) Name() string { return "composer" }

func (Composer) Detect(_ context.Context, repoPath string) ([]Task, error) {
	cPath := filepath.Join(repoPath, "composer.json")
	data, err := os.ReadFile(cPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read composer.json: %w", err)
	}

	var comp struct {
		Scripts map[string]any `json:"scripts"`
	}
	if err := json.Unmarshal(data, &comp); err != nil {
		return nil, fmt.Errorf("parse composer.json: %w", err)
	}

	tasks := make([]Task, 0, len(comp.Scripts)+3)
	for name := range comp.Scripts {
		tasks = append(tasks, Task{
			Source:     "composer",
			Name:       name,
			RawCommand: "composer " + shellQuote(name),
		})
	}

	scriptExists := func(n string) bool { _, ok := comp.Scripts[n]; return ok }
	for _, u := range []struct{ name, cmd string }{
		{"install", "composer install"},
		{"update", "composer update"},
		{"dump-autoload", "composer dump-autoload"},
	} {
		if !scriptExists(u.name) {
			tasks = append(tasks, Task{Source: "composer", Name: u.name, RawCommand: u.cmd})
		}
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}
