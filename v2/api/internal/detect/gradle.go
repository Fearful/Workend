package detect

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type Gradle struct{}

func (Gradle) Name() string { return "gradle" }

func (Gradle) Detect(_ context.Context, repoPath string) ([]Task, error) {
	if !hasGradleManifest(repoPath) {
		return nil, nil
	}

	cmd := "gradle"
	if isExecutable(filepath.Join(repoPath, "gradlew")) {
		cmd = "./gradlew"
	}

	names := []string{"build", "test", "check", "clean", "assemble", "jar", "classes", "dependencies"}
	tasks := make([]Task, 0, len(names))
	for _, n := range names {
		tasks = append(tasks, Task{
			Source:     "gradle",
			Name:       n,
			RawCommand: cmd + " " + shellQuote(n),
		})
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}

func hasGradleManifest(repoPath string) bool {
	for _, name := range []string{"build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts"} {
		if _, err := os.Stat(filepath.Join(repoPath, name)); err == nil {
			return true
		}
	}
	return false
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return info.Mode()&0o111 != 0
}
