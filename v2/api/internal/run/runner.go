// Package run executes detected tasks via Dagger pipelines.
//
// Each Run is a single invocation of a Task. The runner constructs a Dagger
// container from the source-appropriate base image, mounts the project's
// cloned tree, and executes the raw command. Combined stdout+stderr is
// captured to a log file on the logs volume; the exit code goes back into
// the runs table.
package run

import (
	"context"
	"fmt"
	"os"

	dag "dagger.io/dagger"

	wdagger "workend/api/internal/dagger"
)

// Spec captures everything the runner needs to execute a single task. Caller
// resolves these fields by joining the runs / tasks / projects tables.
type Spec struct {
	Source     string // 'npm' | 'just'
	Name       string // task name
	RawCommand string // verbatim command to run inside the container
	RepoPath   string // absolute path to the cloned repo on the api volume
	LogFile    string // absolute path to write combined stdout+stderr to
}

// Result is what Execute returns on completion.
type Result struct {
	ExitCode int
}

// Execute runs the task's command in a Dagger-managed container. Returns
// (Result, nil) on container completion regardless of exit code; returns
// (zero, error) only on infrastructure failures (engine unreachable,
// log write failed, etc.).
func Execute(ctx context.Context, dc *wdagger.Client, spec Spec) (Result, error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("dagger client: %w", err)
	}

	base := baseImage(spec.Source)
	setup := setupSteps(spec.Source)

	container := client.Container().From(base)

	for _, step := range setup {
		container = container.WithExec(step)
	}

	container = container.
		WithMountedDirectory("/work", client.Host().Directory(spec.RepoPath)).
		WithWorkdir("/work").
		WithExec(
			[]string{"sh", "-c", spec.RawCommand + " 2>&1"},
			dag.ContainerWithExecOpts{Expect: dag.ReturnTypeAny},
		)

	output, err := container.Stdout(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("read stdout: %w", err)
	}

	exitCode, err := container.ExitCode(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("read exit code: %w", err)
	}

	if err := os.WriteFile(spec.LogFile, []byte(output), 0o644); err != nil {
		return Result{}, fmt.Errorf("write log: %w", err)
	}

	return Result{ExitCode: exitCode}, nil
}

// baseImage chooses the OCI image used to execute tasks of a given source.
// Pinned tags only — never `latest`, so runs are reproducible.
func baseImage(source string) string {
	switch source {
	case "npm":
		return "node:22-alpine"
	case "just":
		return "alpine:3.20"
	default:
		return "alpine:3.20"
	}
}

// setupSteps returns the WithExec arg lists to run after `From` and before the
// task command. Used to install per-source tooling (e.g., the `just` binary).
func setupSteps(source string) [][]string {
	switch source {
	case "just":
		return [][]string{
			{"apk", "add", "--no-cache", "just", "bash", "git"},
		}
	default:
		return nil
	}
}
