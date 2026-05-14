// Overview returns a small bundle of "human-readable" project facts derived
// from the local clone: the README contents, top contributors, detected
// LICENSE/CONTRIBUTING/CHANGELOG presence, and recent files. Designed to be
// rendered in a single overview panel without further round-trips.
//
// Everything is read off the cached local clone — never re-runs git fetch.
// If a project hasn't cloned yet, returns 404-ish empty fields.
package project

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
)

// MAX_README_BYTES caps how much README we ship to the client. README files
// can be huge (several MB on some repos); the panel needs ~50KB at most for
// a useful preview.
const MAX_README_BYTES = 64 * 1024

// MAX_CONTRIBUTORS limits the shortlog response.
const MAX_CONTRIBUTORS = 10

type OverviewResp struct {
	Readme       *Readme       `json:"readme"`
	Contributors []Contributor `json:"contributors"`
	License      *License      `json:"license"`
	HasContrib   bool          `json:"has_contributing"`
	HasChangelog bool          `json:"has_changelog"`
	RecentFiles  []RecentFile  `json:"recent_files"`
}

type Readme struct {
	Path      string `json:"path"`        // relative path inside repo
	Bytes     int    `json:"bytes"`       // size of the (possibly truncated) content
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
}

type Contributor struct {
	Name    string `json:"name"`
	Commits int    `json:"commits"`
}

type License struct {
	Path string `json:"path"`
	// Detected SPDX-ish identifier: 'MIT', 'Apache-2.0', etc, or 'Unknown'.
	Kind string `json:"kind"`
}

type RecentFile struct {
	Path     string `json:"path"`
	Modified string `json:"modified"` // ISO8601
}

// GET /api/projects/:id/overview
func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	p, err := h.fetchOwned(r.Context(), uid, pid)
	if err != nil || p == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if p.LocalPath == nil || *p.LocalPath == "" {
		// Project not cloned yet — return empty bundle, not error.
		writeJSON(w, OverviewResp{Contributors: []Contributor{}, RecentFiles: []RecentFile{}})
		return
	}
	repo := *p.LocalPath
	resp := OverviewResp{
		Readme:       readFirstReadme(repo),
		Contributors: gitContributors(r.Context(), repo, MAX_CONTRIBUTORS),
		License:      detectLicense(repo),
		HasContrib:   exists(repo, "CONTRIBUTING.md") || exists(repo, "CONTRIBUTING") || exists(repo, "docs/CONTRIBUTING.md"),
		HasChangelog: exists(repo, "CHANGELOG.md") || exists(repo, "CHANGELOG") || exists(repo, "HISTORY.md"),
		RecentFiles:  gitRecentFiles(r.Context(), repo, 10),
	}
	if resp.Contributors == nil {
		resp.Contributors = []Contributor{}
	}
	if resp.RecentFiles == nil {
		resp.RecentFiles = []RecentFile{}
	}
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// readFirstReadme finds README* in the repo root case-insensitively. We
// check explicit candidates first (most projects use README.md) and fall
// back to a directory scan only if none match.
func readFirstReadme(repo string) *Readme {
	candidates := []string{"README.md", "README.rst", "README.txt", "README"}
	// Try the canonical names first.
	for _, name := range candidates {
		if r := tryReadme(repo, name); r != nil {
			return r
		}
	}
	// Case-insensitive fallback: scan for any README*.
	entries, err := os.ReadDir(repo)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		lc := strings.ToLower(e.Name())
		if strings.HasPrefix(lc, "readme") {
			if r := tryReadme(repo, e.Name()); r != nil {
				return r
			}
		}
	}
	return nil
}

func tryReadme(repo, name string) *Readme {
	full := filepath.Join(repo, name)
	st, err := os.Stat(full)
	if err != nil || st.IsDir() {
		return nil
	}
	f, err := os.Open(full)
	if err != nil {
		return nil
	}
	defer f.Close()
	buf := make([]byte, MAX_README_BYTES)
	n, _ := f.Read(buf)
	truncated := false
	if int64(n) < st.Size() {
		truncated = true
	}
	return &Readme{
		Path:      name,
		Bytes:     n,
		Truncated: truncated,
		Content:   string(buf[:n]),
	}
}

// gitContributors runs `git shortlog -sn --no-merges` to list authors. We
// capture stderr separately and ignore git failures (e.g. shallow clones)
// rather than 500ing the whole overview.
func gitContributors(ctx context.Context, repo string, limit int) []Contributor {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", repo, "shortlog", "-sn", "--no-merges", "HEAD")
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []Contributor
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// "  123\tAuthor Name"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}
		result = append(result, Contributor{Name: parts[1], Commits: count})
		if len(result) >= limit {
			break
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Commits > result[j].Commits })
	return result
}

// gitRecentFiles uses `git log --name-only` to collect the most recently
// touched files at the current HEAD. Bounded by limit.
func gitRecentFiles(ctx context.Context, repo string, limit int) []RecentFile {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", repo,
		"log", "--name-only", "--pretty=format:%H %cI", "-n", "20")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	seen := map[string]string{} // path -> latest ISO timestamp
	order := []string{}
	var currentDate string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Header lines look like "<sha> <iso>".
		if len(line) > 40 && line[40] == ' ' {
			currentDate = strings.TrimSpace(line[40:])
			continue
		}
		if _, ok := seen[line]; !ok {
			seen[line] = currentDate
			order = append(order, line)
			if len(order) >= limit {
				break
			}
		}
	}
	out2 := make([]RecentFile, 0, len(order))
	for _, p := range order {
		out2 = append(out2, RecentFile{Path: p, Modified: seen[p]})
	}
	return out2
}

// detectLicense reads the first found license file and tries simple keyword
// detection. Anything we can't classify is surfaced as Kind="Unknown" so the
// UI can still link to the file.
func detectLicense(repo string) *License {
	candidates := []string{"LICENSE", "LICENSE.md", "LICENSE.txt", "LICENCE", "COPYING", "COPYING.md"}
	for _, name := range candidates {
		full := filepath.Join(repo, name)
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			continue
		}
		raw, err := os.ReadFile(full)
		if err != nil {
			continue
		}
		text := strings.ToLower(string(raw))
		// quick heuristics — not authoritative, just useful labels
		switch {
		case strings.Contains(text, "mit license"):
			return &License{Path: name, Kind: "MIT"}
		case strings.Contains(text, "apache license") && strings.Contains(text, "version 2"):
			return &License{Path: name, Kind: "Apache-2.0"}
		case strings.Contains(text, "gnu general public license") && strings.Contains(text, "version 3"):
			return &License{Path: name, Kind: "GPL-3.0"}
		case strings.Contains(text, "gnu general public license") && strings.Contains(text, "version 2"):
			return &License{Path: name, Kind: "GPL-2.0"}
		case strings.Contains(text, "bsd 3-clause"), strings.Contains(text, "redistribution and use in source and binary forms") && strings.Contains(text, "neither the name"):
			return &License{Path: name, Kind: "BSD-3-Clause"}
		case strings.Contains(text, "mozilla public license"):
			return &License{Path: name, Kind: "MPL-2.0"}
		case strings.Contains(text, "isc license"):
			return &License{Path: name, Kind: "ISC"}
		case strings.Contains(text, "this is free and unencumbered software released into the public domain"):
			return &License{Path: name, Kind: "Unlicense"}
		}
		return &License{Path: name, Kind: "Unknown"}
	}
	return nil
}

func exists(repo, rel string) bool {
	_, err := os.Stat(filepath.Join(repo, rel))
	return err == nil
}
