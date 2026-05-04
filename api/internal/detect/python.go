package detect

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Python struct{}

func (Python) Name() string { return "python" }

func (Python) Detect(_ context.Context, repoPath string) ([]Task, error) {
	seen := map[string]struct{}{}
	tasks := []Task{}

	if envs, ok := parseToxEnvlist(repoPath); ok {
		for _, env := range envs {
			if _, dup := seen[env]; dup {
				continue
			}
			seen[env] = struct{}{}
			tasks = append(tasks, Task{
				Source:     "python",
				Name:       env,
				RawCommand: "tox -e " + shellQuote(env),
			})
		}
	}

	if scripts, ok := parsePyprojectScripts(repoPath); ok {
		for _, name := range scripts {
			if _, dup := seen[name]; dup {
				continue
			}
			seen[name] = struct{}{}
			tasks = append(tasks, Task{
				Source:     "python",
				Name:       name,
				RawCommand: shellQuote(name),
			})
		}
	}

	hasPython := len(tasks) > 0 || hasPythonMarker(repoPath)
	if hasPython {
		for _, s := range []struct{ name, cmd string }{
			{"test", "pytest"},
			{"lint", "ruff check ."},
		} {
			if _, dup := seen[s.name]; dup {
				continue
			}
			seen[s.name] = struct{}{}
			tasks = append(tasks, Task{
				Source:     "python",
				Name:       s.name,
				RawCommand: s.cmd,
			})
		}
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}

func parseToxEnvlist(repoPath string) ([]string, bool) {
	toxPath := filepath.Join(repoPath, "tox.ini")
	f, err := os.Open(toxPath)
	if err != nil {
		return nil, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "envlist") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			raw := strings.TrimSpace(parts[1])
			var envs []string
			for _, e := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == '\n' }) {
				e = strings.TrimSpace(e)
				if e != "" {
					envs = append(envs, e)
				}
			}
			if len(envs) > 0 {
				return envs, true
			}
		}
	}
	return nil, false
}

func parsePyprojectScripts(repoPath string) ([]string, bool) {
	ppPath := filepath.Join(repoPath, "pyproject.toml")
	f, err := os.Open(ppPath)
	if err != nil {
		return nil, false
	}
	defer f.Close()

	var scripts []string
	inScripts := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "[project.scripts]" {
			inScripts = true
			continue
		}
		if inScripts {
			if strings.HasPrefix(trimmed, "[") {
				break
			}
			if idx := strings.IndexByte(trimmed, '='); idx > 0 {
				name := strings.TrimSpace(trimmed[:idx])
				if name != "" {
					scripts = append(scripts, name)
				}
			}
		}
	}
	if len(scripts) > 0 {
		return scripts, true
	}
	return nil, false
}

func hasPythonMarker(repoPath string) bool {
	for _, name := range []string{"pyproject.toml", "setup.py", "tox.ini", "noxfile.py", "requirements.txt"} {
		if _, err := os.Stat(filepath.Join(repoPath, name)); err == nil {
			return true
		} else if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "stat %s: %v\n", name, err)
		}
	}
	return false
}
