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

type Makefile struct{}

func (Makefile) Name() string { return "make" }

var makeTargetRe = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)

func (Makefile) Detect(_ context.Context, repoPath string) ([]Task, error) {
	path, ok := findMakefile(repoPath)
	if !ok {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open makefile: %w", err)
	}
	defer f.Close()

	seen := map[string]struct{}{}
	tasks := []Task{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":=") {
			continue
		}
		m := makeTargetRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		if strings.HasPrefix(name, ".") {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		tasks = append(tasks, Task{
			Source:     "make",
			Name:       name,
			RawCommand: "make " + shellQuote(name),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan makefile: %w", err)
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}

func findMakefile(repoPath string) (string, bool) {
	for _, name := range []string{"Makefile", "GNUmakefile", "makefile"} {
		p := filepath.Join(repoPath, name)
		if _, err := os.Stat(p); err == nil {
			return p, true
		} else if !errors.Is(err, os.ErrNotExist) {
			continue
		}
	}
	return "", false
}
