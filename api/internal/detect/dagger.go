package detect

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	dag "dagger.io/dagger"

	wdagger "workend/api/internal/dagger"
)

// Dagger detects functions exposed by a Dagger module (presence of
// dagger.json at the repo root). Implementation strategy: shell out to the
// `dagger` CLI inside a container with the repo mounted, parse `dagger
// functions` output, emit one Task per function name.
//
// This is the most expensive built-in detector — pulls and runs the
// dagger CLI image — so it's gated on dagger.json being present.
type Dagger struct {
	Dagger *wdagger.Client
}

func (Dagger) Name() string { return "dagger" }

// daggerCLIImage matches versions.md. Bump in lock-step with the engine.
const daggerCLIImage = "registry.dagger.io/cli:v0.20.5"

func (d Dagger) Detect(ctx context.Context, repoPath string) ([]Task, error) {
	if _, err := os.Stat(filepath.Join(repoPath, "dagger.json")); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if d.Dagger == nil {
		return nil, nil // no dagger client wired
	}
	client, err := d.Dagger.Get(ctx)
	if err != nil {
		return nil, err
	}

	out, err := client.Container().
		From(daggerCLIImage).
		WithMountedDirectory("/repo", client.Host().Directory(repoPath)).
		WithWorkdir("/repo").
		WithExec(
			[]string{"dagger", "functions", "--no-mod-load=false"},
			dag.ContainerWithExecOpts{Expect: dag.ReturnTypeAny},
		).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("dagger functions exec: %w", err)
	}

	tasks := []Task{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		// `dagger functions` output: header line then 2-column table
		// (NAME  DESCRIPTION). Skip the header and any line that doesn't
		// look like a function name at column 0.
		if first {
			first = false
			if strings.HasPrefix(strings.ToLower(line), "name") {
				continue
			}
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		// Sanity: function names are kebab-case alphanumerics.
		if !isValidFnName(name) {
			continue
		}
		tasks = append(tasks, Task{
			Source:     "dagger",
			Name:       name,
			RawCommand: "dagger call " + shellQuote(name),
		})
	}
	return tasks, nil
}

func isValidFnName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		isLower := r >= 'a' && r <= 'z'
		isUpper := r >= 'A' && r <= 'Z'
		isDigit := r >= '0' && r <= '9'
		isDash := r == '-' || r == '_'
		if i == 0 && !(isLower || isUpper) {
			return false
		}
		if !(isLower || isUpper || isDigit || isDash) {
			return false
		}
	}
	return true
}
