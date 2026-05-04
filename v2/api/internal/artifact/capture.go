// Capture walks a list of glob patterns inside the runner's working directory
// and copies matched files into the artifacts root, recording metadata in
// the artifacts table. Designed to be called from the run executor after a
// run finishes.
package artifact

import (
	"context"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MaxArtifactBytes caps a single captured file's size. Larger files are
// skipped so a misconfigured pattern doesn't fill the disk.
const MaxArtifactBytes = 50 * 1024 * 1024

// MaxArtifactsPerRun limits how many files we capture per run.
const MaxArtifactsPerRun = 200

// Capture matches `patterns` (filepath.Match doublestar-lite — only literal
// + leading "**/" supported) against `repoPath` and records each match.
//
// Returns the number of files actually captured. Errors during a single
// file's copy are logged-and-skipped, not returned.
func Capture(ctx context.Context, pool *pgxpool.Pool, artifactsRoot string, runID uuid.UUID, repoPath string, patterns []string) int {
	if len(patterns) == 0 {
		return 0
	}
	matches := map[string]struct{}{}
	for _, pat := range patterns {
		ms := globMatchAll(repoPath, pat)
		for _, m := range ms {
			matches[m] = struct{}{}
			if len(matches) >= MaxArtifactsPerRun {
				break
			}
		}
		if len(matches) >= MaxArtifactsPerRun {
			break
		}
	}
	if len(matches) == 0 {
		return 0
	}

	dst := filepath.Join(artifactsRoot, runID.String())
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return 0
	}

	captured := 0
	for full := range matches {
		rel, err := filepath.Rel(repoPath, full)
		if err != nil {
			continue
		}
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			continue
		}
		if st.Size() > MaxArtifactBytes {
			continue
		}
		safeName := SanitizeRelPath(rel)
		storage := filepath.Join(dst, safeName)
		if err := os.MkdirAll(filepath.Dir(storage), 0o755); err != nil {
			continue
		}
		if err := copyFile(full, storage); err != nil {
			continue
		}
		mt := mime.TypeByExtension(filepath.Ext(full))
		var mimePtr *string
		if mt != "" {
			mimePtr = &mt
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO artifacts (run_id, relative_path, storage_path, size_bytes, mime_type)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (run_id, relative_path) DO NOTHING
		`, runID, rel, storage, st.Size(), mimePtr)
		if err == nil {
			captured++
		}
	}
	return captured
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// globMatchAll walks `root` and returns every file that matches `pattern`.
// Supports a small grammar suitable for artifact globs:
//   - "*.xml"               — at any depth
//   - "coverage/*.xml"      — exact directory + glob
//   - "**/junit.xml"        — at any depth (alias for the bare name above)
//   - "build/output.tar.gz" — literal path
func globMatchAll(root, pattern string) []string {
	pat := filepath.ToSlash(pattern)
	if pat == "" {
		return nil
	}

	// "**/" prefix means "at any depth" — same as the bare basename.
	anyDepth := false
	if rest, ok := stripPrefix(pat, "**/"); ok {
		pat = rest
		anyDepth = true
	}
	// If the user just gave us a glob without slashes, search any depth.
	if !anyDepth && !strings.Contains(pat, "/") {
		anyDepth = true
	}

	var results []string
	if anyDepth {
		_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			matched, _ := filepath.Match(pat, filepath.Base(p))
			if matched {
				results = append(results, p)
			}
			return nil
		})
		return results
	}

	// Otherwise treat as relative to root.
	abs := filepath.Join(root, filepath.FromSlash(pat))
	matches, _ := filepath.Glob(abs)
	results = append(results, matches...)
	return results
}

// stripPrefix returns (rest, true) if s starts with prefix.
func stripPrefix(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		return s, false
	}
	return s[len(prefix):], true
}
