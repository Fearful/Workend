package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		lockFile string
		want     string
	}{
		{"explicit pnpm via field", "pnpm@9.1.0", "", "pnpm"},
		{"explicit yarn via field", "yarn@4.0.0", "", "yarn"},
		{"explicit bun via field", "bun@1.0.0", "", "bun"},
		{"explicit npm via field", "npm@10.0.0", "", "npm"},
		{"field wins over lockfile", "pnpm@9.0.0", "yarn.lock", "pnpm"},
		{"field without version", "pnpm", "", "pnpm"},
		{"unknown field falls through", "rush@5.0.0", "pnpm-lock.yaml", "pnpm"},
		{"lockfile pnpm", "", "pnpm-lock.yaml", "pnpm"},
		{"lockfile yarn", "", "yarn.lock", "yarn"},
		{"lockfile bun .lockb", "", "bun.lockb", "bun"},
		{"lockfile bun .lock", "", "bun.lock", "bun"},
		{"lockfile npm", "", "package-lock.json", "npm"},
		{"no signal defaults to npm", "", "", "npm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.lockFile != "" {
				if err := os.WriteFile(filepath.Join(dir, tt.lockFile), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			got := detectPackageManager(dir, tt.field)
			if got != tt.want {
				t.Errorf("detectPackageManager(%q, %q) = %q, want %q", tt.lockFile, tt.field, got, tt.want)
			}
		})
	}
}

func TestDetectPackageManagerLockfilePriority(t *testing.T) {
	// When multiple lockfiles exist (e.g., during a migration), the more
	// specific manager wins. pnpm-lock.yaml beats package-lock.json.
	dir := t.TempDir()
	for _, f := range []string{"pnpm-lock.yaml", "package-lock.json"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := detectPackageManager(dir, ""); got != "pnpm" {
		t.Errorf("got %q, want pnpm", got)
	}
}
