package repo

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// SearchResult represents a single search hit.
type SearchResult struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// skipDirs is the set of directory names skipped during recursive search.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
	".next":        true,
}

// Search searches for a pattern in the repo at repoPath.
// Returns up to maxResults matching lines (case-insensitive).
func Search(repoPath, query string, maxResults int) ([]SearchResult, error) {
	if maxResults <= 0 {
		maxResults = 100
	}
	if maxResults > 500 {
		maxResults = 500
	}

	lowerQuery := strings.ToLower(query)
	var results []SearchResult

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if len(results) >= maxResults {
			return filepath.SkipAll
		}

		// Skip binary files: check first 512 bytes for null bytes.
		if isBinary(path) {
			return nil
		}

		rel, err := filepath.Rel(repoPath, path)
		if err != nil {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if strings.Contains(strings.ToLower(line), lowerQuery) {
				content := line
				if len(content) > 500 {
					content = content[:500]
				}
				results = append(results, SearchResult{
					File:    rel,
					Line:    lineNum,
					Content: content,
				})
				if len(results) >= maxResults {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// isBinary checks the first 512 bytes of a file for null bytes.
func isBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true // skip on error
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if n == 0 {
		return false
	}
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}
