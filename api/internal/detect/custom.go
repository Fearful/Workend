package detect

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	dag "dagger.io/dagger"

	wdagger "workend/api/internal/dagger"
)

// CustomConfig is one entry in WORKEND_DETECTORS_FILE — a JSON array of
// these. Each entry defines: a manifest file or glob to look for, a base
// container image to run in, and a shell command whose stdout (one task
// name per line) becomes the discovered tasks.
type CustomConfig struct {
	// Source name to record on each Task (e.g., "make", "cargo").
	Source string `json:"source"`
	// Manifest filename to require at the repo root, e.g. "Makefile".
	// If absent, this detector is a no-op for that repo.
	Manifest string `json:"manifest"`
	// OCI image to run the detect command in.
	Image string `json:"image"`
	// Optional setup commands run before DetectCmd (e.g. installing tools).
	Setup []string `json:"setup,omitempty"`
	// Shell command. stdout: one task name per line. Repo is mounted at /repo.
	DetectCmd string `json:"detect_cmd"`
	// Template for the runnable command. Use {{name}} for the task name.
	// e.g., "make {{name}}"
	RunCmdTemplate string `json:"run_cmd_template"`
}

// Custom is a Detector backed by a CustomConfig. Detection runs the configured
// command inside a Dagger container against the repo dir.
type Custom struct {
	Cfg    CustomConfig
	Dagger *wdagger.Client
}

func (c Custom) Name() string { return "custom:" + c.Cfg.Source }

func (c Custom) Detect(ctx context.Context, repoPath string) ([]Task, error) {
	if c.Cfg.Manifest != "" {
		if _, err := os.Stat(filepath.Join(repoPath, c.Cfg.Manifest)); err != nil {
			return nil, nil
		}
	}
	client, err := c.Dagger.Get(ctx)
	if err != nil {
		return nil, err
	}

	container := client.Container().From(c.Cfg.Image)
	for _, step := range c.Cfg.Setup {
		container = container.WithExec([]string{"sh", "-c", step})
	}
	out, err := container.
		WithMountedDirectory("/repo", client.Host().Directory(repoPath)).
		WithWorkdir("/repo").
		WithExec(
			[]string{"sh", "-c", c.Cfg.DetectCmd},
			dag.ContainerWithExecOpts{Expect: dag.ReturnTypeAny},
		).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("custom detect (%s): %w", c.Cfg.Source, err)
	}

	tasks := []Task{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" || strings.HasPrefix(name, "#") {
			continue
		}
		runCmd := strings.ReplaceAll(c.Cfg.RunCmdTemplate, "{{name}}", shellQuote(name))
		tasks = append(tasks, Task{
			Source:     c.Cfg.Source,
			Name:       name,
			RawCommand: runCmd,
		})
	}
	return tasks, nil
}

// LoadCustom reads WORKEND_DETECTORS_FILE (JSON array of CustomConfig) and
// returns a slice of Detectors ready to register. Returns (nil, nil) if the
// env var is unset.
func LoadCustom(dc *wdagger.Client) ([]Detector, error) {
	path := os.Getenv("WORKEND_DETECTORS_FILE")
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read detectors file: %w", err)
	}
	var cfgs []CustomConfig
	if err := json.Unmarshal(data, &cfgs); err != nil {
		return nil, fmt.Errorf("parse detectors file: %w", err)
	}
	out := make([]Detector, 0, len(cfgs))
	for _, c := range cfgs {
		if c.Source == "" || c.Image == "" || c.DetectCmd == "" || c.RunCmdTemplate == "" {
			return nil, fmt.Errorf("detector entry missing required fields (source, image, detect_cmd, run_cmd_template)")
		}
		out = append(out, Custom{Cfg: c, Dagger: dc})
	}
	return out, nil
}
