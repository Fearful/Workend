// Package deps parses common lockfiles into project_dependencies rows. Runs
// after every successful project sync so the dependency graph view always
// matches the working tree.
package deps

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

type Dependency struct {
	ID        uuid.UUID `json:"id"`
	Ecosystem string    `json:"ecosystem"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Depth     int       `json:"depth"`
	ParsedAt  time.Time `json:"parsed_at"`
}

// item is the internal tuple parsers emit. Promoted to a package-level
// type so the parser helpers can return the same slice shape.
type item struct {
	ecosystem string
	name      string
	version   string
	depth     int
}

// Parse walks the working tree, recognizes lockfiles, and upserts the
// resulting dependencies. Existing rows for ecosystems we re-parsed are
// pruned first so removed deps disappear.
//
// Deliberately conservative: we only emit what the lockfile shape makes
// obvious. Resolved transitive trees are not built — pipelines that need
// them can stream from the SBOM instead.
func Parse(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID, repoPath string) error {
	var found []item
	parsedEcosystems := map[string]struct{}{}

	if items, ok := parseNPMLock(filepath.Join(repoPath, "package-lock.json")); ok {
		found = append(found, items...)
		parsedEcosystems["npm"] = struct{}{}
	}
	if items, ok := parseGoSum(filepath.Join(repoPath, "go.sum")); ok {
		found = append(found, items...)
		parsedEcosystems["go"] = struct{}{}
	}
	if items, ok := parseCargoLock(filepath.Join(repoPath, "Cargo.lock")); ok {
		found = append(found, items...)
		parsedEcosystems["cargo"] = struct{}{}
	}
	if items, ok := parsePythonRequirements(filepath.Join(repoPath, "requirements.txt")); ok {
		found = append(found, items...)
		parsedEcosystems["pypi"] = struct{}{}
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for eco := range parsedEcosystems {
		if _, err := tx.Exec(ctx,
			`DELETE FROM project_dependencies WHERE project_id = $1 AND ecosystem = $2`,
			projectID, eco); err != nil {
			return err
		}
	}
	for _, it := range found {
		if _, err := tx.Exec(ctx, `
			INSERT INTO project_dependencies (project_id, ecosystem, name, version, depth)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (project_id, ecosystem, name, version) DO UPDATE SET depth = EXCLUDED.depth, parsed_at = now()
		`, projectID, it.ecosystem, it.name, it.version, it.depth); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// --- per-ecosystem parsers ---

// parseNPMLock walks `packages` in package-lock.json v2/v3, marking the
// keys directly listed under packages[""].dependencies as direct (depth 0).
func parseNPMLock(path string) ([]item, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var doc struct {
		Packages map[string]struct {
			Version      string            `json:"version"`
			Dependencies map[string]string `json:"dependencies"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, false
	}
	direct := map[string]struct{}{}
	if root, ok := doc.Packages[""]; ok {
		for name := range root.Dependencies {
			direct[name] = struct{}{}
		}
	}
	var out []item
	for key, pkg := range doc.Packages {
		if key == "" || pkg.Version == "" {
			continue
		}
		name := key
		if i := strings.LastIndex(name, "node_modules/"); i >= 0 {
			name = name[i+len("node_modules/"):]
		}
		depth := 1
		if _, ok := direct[name]; ok {
			depth = 0
		}
		out = append(out, item{ecosystem: "npm", name: name, version: pkg.Version, depth: depth})
	}
	return out, true
}

// parseGoSum dedupes lines (the `/go.mod` companion entries) and emits at
// depth 0 — go's lock is flat.
func parseGoSum(path string) ([]item, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	seen := map[string]struct{}{}
	var out []item
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		version := strings.TrimSuffix(parts[1], "/go.mod")
		key := name + "\x00" + version
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item{ecosystem: "go", name: name, version: version, depth: 0})
	}
	return out, true
}

// parseCargoLock looks for [[package]] / name = / version = triplets without
// dragging in a full TOML parser.
func parseCargoLock(path string) ([]item, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	var out []item
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var name, version string
	flush := func() {
		if name != "" && version != "" {
			out = append(out, item{ecosystem: "cargo", name: name, version: version, depth: 0})
		}
		name, version = "", ""
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "[[package]]":
			flush()
		case strings.HasPrefix(line, "name = "):
			name = strings.Trim(line[len("name = "):], `"`)
		case strings.HasPrefix(line, "version = "):
			version = strings.Trim(line[len("version = "):], `"`)
		}
	}
	flush()
	return out, true
}

// parsePythonRequirements: lines like `requests==2.31.0 # comment`. Skips
// blanks, editable installs, and -r/-e directives.
func parsePythonRequirements(path string) ([]item, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	var out []item
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		var name, version string
		for _, sep := range []string{"==", ">=", "<=", "~=", ">", "<"} {
			if i := strings.Index(line, sep); i >= 0 {
				name = strings.TrimSpace(line[:i])
				version = strings.TrimSpace(line[i+len(sep):])
				break
			}
		}
		if name == "" {
			continue
		}
		out = append(out, item{ecosystem: "pypi", name: name, version: version, depth: 0})
	}
	return out, true
}

// --- HTTP handlers ---

type Handlers struct {
	Pool *pgxpool.Pool
}

// ListByProject returns the parsed dependencies grouped by ecosystem.
// GET /api/projects/:id/dependencies
func (h *Handlers) ListByProject(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}
	if !userOwnsProject(r.Context(), h.Pool, uid, pid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, ecosystem, name, version, depth, parsed_at
		FROM project_dependencies WHERE project_id = $1
		ORDER BY ecosystem, depth, name
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Dependency{}
	for rows.Next() {
		var d Dependency
		if err := rows.Scan(&d.ID, &d.Ecosystem, &d.Name, &d.Version, &d.Depth, &d.ParsedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, d)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func userOwnsProject(ctx context.Context, pool *pgxpool.Pool, userID, projectID uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM projects p
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE p.id = $1 AND m.user_id = $2
	`, projectID, userID).Scan(&n)
	return err == nil
}
