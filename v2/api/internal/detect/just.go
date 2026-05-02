package detect

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Just struct{}

func (Just) Name() string { return "just" }

// recipeRe matches a justfile recipe definition at the start of a line:
//
//	name:
//	name(arg1, arg2):
//	name +arg:
//	@name:           (silent recipe)
//
// We extract the name only — full just parsing (params, dependencies, etc.)
// is deferred. Supporting it later means swapping this implementation to
// invoke `just --list --justfile <path>` inside a Dagger container.
var recipeRe = regexp.MustCompile(`^@?([a-zA-Z_][a-zA-Z0-9_-]*)\s*(?:\([^)]*\))?\s*(?:\+[a-zA-Z_][a-zA-Z0-9_-]*)?\s*:`)

func (Just) Detect(_ context.Context, repoPath string) ([]Task, error) {
	path, ok := findJustfile(repoPath)
	if !ok {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open justfile: %w", err)
	}
	defer f.Close()

	seen := map[string]struct{}{}
	tasks := []Task{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == ' ' || line[0] == '\t' || line[0] == '#' {
			continue
		}
		// Skip variable assignments: name := value
		if strings.Contains(line, ":=") {
			continue
		}
		m := recipeRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		tasks = append(tasks, Task{
			Source:     "just",
			Name:       name,
			RawCommand: "just " + shellQuote(name),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan justfile: %w", err)
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}

// findJustfile returns the path to a justfile, supporting both the lowercase
// and capitalized forms that just accepts.
func findJustfile(repoPath string) (string, bool) {
	for _, name := range []string{"justfile", "Justfile", ".justfile"} {
		p := filepath.Join(repoPath, name)
		if _, err := os.Stat(p); err == nil {
			return p, true
		} else if !errors.Is(err, os.ErrNotExist) {
			// permission denied or similar — skip
			continue
		}
	}
	return "", false
}
