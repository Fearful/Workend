package detect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
		Scripts        map[string]string `json:"scripts"`
		PackageManager string            `json:"packageManager"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	pm := detectPackageManager(repoPath, pkg.PackageManager)

	tasks := make([]Task, 0, len(pkg.Scripts)+4)
	for name := range pkg.Scripts {
		tasks = append(tasks, Task{
			Source:     pm,
			Name:       name,
			RawCommand: pm + " run " + shellQuote(name),
		})
	}

	scriptExists := func(n string) bool { _, ok := pkg.Scripts[n]; return ok }
	for _, u := range []struct{ name, cmd string }{
		{"install", pm + " install"},
		{"update", pm + " update"},
		{"outdated", pm + " outdated"},
		{"audit", pm + " audit"},
	} {
		if !scriptExists(u.name) {
			tasks = append(tasks, Task{Source: pm, Name: u.name, RawCommand: u.cmd})
		}
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Name < tasks[j].Name })
	return tasks, nil
}

// detectPackageManager picks the Node package manager to invoke for a repo.
// Priority: explicit `packageManager` field in package.json, then lock-file
// presence, then npm as fallback. Returns one of: npm, pnpm, yarn, bun.
func detectPackageManager(repoPath, packageManagerField string) string {
	if packageManagerField != "" {
		// Spec format is "<name>@<version>"; we only care about the name.
		name := packageManagerField
		if at := strings.IndexByte(name, '@'); at > 0 {
			name = name[:at]
		}
		switch name {
		case "npm", "pnpm", "yarn", "bun":
			return name
		}
	}
	for _, lf := range []struct {
		file string
		pm   string
	}{
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"bun.lockb", "bun"},
		{"bun.lock", "bun"},
		{"package-lock.json", "npm"},
	} {
		if _, err := os.Stat(filepath.Join(repoPath, lf.file)); err == nil {
			return lf.pm
		}
	}
	return "npm"
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
