// Package monorepo detects sub-packages inside monorepos and maps them to
// tasks via package_task_scopes.
package monorepo

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// --- types ---

// Package is one detected sub-package inside a monorepo.
type Package struct {
	ID         uuid.UUID `json:"id"`
	ProjectID  uuid.UUID `json:"project_id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	PkgType    string    `json:"pkg_type"`
	Version    string    `json:"version"`
	DetectedAt time.Time `json:"detected_at"`
	TaskCount  int       `json:"task_count,omitempty"`
}

// PackageTaskScope links a package to a task.
type PackageTaskScope struct {
	ID        uuid.UUID `json:"id"`
	PackageID uuid.UUID `json:"package_id"`
	TaskID    uuid.UUID `json:"task_id"`
	TaskName  string    `json:"task_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// RawPackage is a package discovered by scanning the filesystem before
// persisting to the database.
type RawPackage struct {
	Name    string
	Path    string
	PkgType string
	Version string
}

// Handlers holds shared dependencies for monorepo endpoints.
type Handlers struct {
	Pool      *pgxpool.Pool
	ReposRoot string
	Logger    *slog.Logger
}

// --- detection ---

// DetectPackages scans a repository for monorepo sub-packages. It checks npm
// workspaces, Go workspaces, Cargo workspaces, Python multi-package layouts,
// and Gradle multi-project builds.
func DetectPackages(repoPath string) ([]RawPackage, error) {
	var pkgs []RawPackage

	npm, err := detectNpmWorkspaces(repoPath)
	if err == nil {
		pkgs = append(pkgs, npm...)
	}

	goPkgs, err := detectGoWorkspace(repoPath)
	if err == nil {
		pkgs = append(pkgs, goPkgs...)
	}

	cargo, err := detectCargoWorkspace(repoPath)
	if err == nil {
		pkgs = append(pkgs, cargo...)
	}

	python, err := detectPythonPackages(repoPath)
	if err == nil {
		pkgs = append(pkgs, python...)
	}

	gradle, err := detectGradleProjects(repoPath)
	if err == nil {
		pkgs = append(pkgs, gradle...)
	}

	return pkgs, nil
}

// --- npm workspaces ---

func detectNpmWorkspaces(repoPath string) ([]RawPackage, error) {
	data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
	if err != nil {
		return nil, err
	}

	var root struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	if root.Workspaces == nil {
		return nil, nil
	}

	// Workspaces can be an array of globs or an object with "packages" key.
	var globs []string
	if err := json.Unmarshal(root.Workspaces, &globs); err != nil {
		var obj struct {
			Packages []string `json:"packages"`
		}
		if err := json.Unmarshal(root.Workspaces, &obj); err != nil {
			return nil, err
		}
		globs = obj.Packages
	}

	var pkgs []RawPackage
	for _, g := range globs {
		matches, err := filepath.Glob(filepath.Join(repoPath, g))
		if err != nil {
			continue
		}
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil || !info.IsDir() {
				continue
			}
			pkgData, err := os.ReadFile(filepath.Join(m, "package.json"))
			if err != nil {
				continue
			}
			var pkg struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			}
			if err := json.Unmarshal(pkgData, &pkg); err != nil {
				continue
			}
			relPath, _ := filepath.Rel(repoPath, m)
			name := pkg.Name
			if name == "" {
				name = filepath.Base(m)
			}
			pkgs = append(pkgs, RawPackage{
				Name:    name,
				Path:    relPath,
				PkgType: "npm",
				Version: pkg.Version,
			})
		}
	}
	return pkgs, nil
}

// --- Go workspace ---

var goUseRe = regexp.MustCompile(`^\s*(\./\S+|\S+)\s*$`)

func detectGoWorkspace(repoPath string) ([]RawPackage, error) {
	f, err := os.Open(filepath.Join(repoPath, "go.work"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var paths []string
	inUseBlock := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "use (") || line == "use (" {
			inUseBlock = true
			continue
		}
		if inUseBlock {
			if line == ")" {
				inUseBlock = false
				continue
			}
			if m := goUseRe.FindStringSubmatch(line); m != nil {
				paths = append(paths, strings.TrimPrefix(m[1], "./"))
			}
			continue
		}
		// Single-line use directive: use ./path
		if strings.HasPrefix(line, "use ") {
			p := strings.TrimSpace(strings.TrimPrefix(line, "use"))
			p = strings.TrimPrefix(p, "./")
			if p != "" {
				paths = append(paths, p)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var pkgs []RawPackage
	for _, p := range paths {
		absPath := filepath.Join(repoPath, p)
		modData, err := os.ReadFile(filepath.Join(absPath, "go.mod"))
		if err != nil {
			continue
		}
		modName := parseGoModuleName(string(modData))
		if modName == "" {
			modName = filepath.Base(p)
		}
		pkgs = append(pkgs, RawPackage{
			Name:    modName,
			Path:    p,
			PkgType: "go",
		})
	}
	return pkgs, nil
}

func parseGoModuleName(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return ""
}

// --- Cargo workspace ---

var cargoMembersRe = regexp.MustCompile(`"([^"]+)"`)

func detectCargoWorkspace(repoPath string) ([]RawPackage, error) {
	data, err := os.ReadFile(filepath.Join(repoPath, "Cargo.toml"))
	if err != nil {
		return nil, err
	}

	content := string(data)
	// Check for [workspace] section.
	wsIdx := strings.Index(content, "[workspace]")
	if wsIdx < 0 {
		return nil, nil
	}

	// Find members = [...] after [workspace].
	membersIdx := strings.Index(content[wsIdx:], "members")
	if membersIdx < 0 {
		return nil, nil
	}

	// Extract the members array. It may span multiple lines.
	afterMembers := content[wsIdx+membersIdx:]
	eqIdx := strings.Index(afterMembers, "=")
	if eqIdx < 0 {
		return nil, nil
	}
	afterEq := afterMembers[eqIdx+1:]

	openBracket := strings.Index(afterEq, "[")
	if openBracket < 0 {
		return nil, nil
	}
	closeBracket := strings.Index(afterEq[openBracket:], "]")
	if closeBracket < 0 {
		return nil, nil
	}
	membersStr := afterEq[openBracket : openBracket+closeBracket+1]

	// Extract glob patterns from the members array.
	matches := cargoMembersRe.FindAllStringSubmatch(membersStr, -1)
	var patterns []string
	for _, m := range matches {
		patterns = append(patterns, m[1])
	}

	var pkgs []RawPackage
	for _, pattern := range patterns {
		globbed, err := filepath.Glob(filepath.Join(repoPath, pattern))
		if err != nil {
			continue
		}
		for _, dir := range globbed {
			info, err := os.Stat(dir)
			if err != nil || !info.IsDir() {
				continue
			}
			cargoData, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
			if err != nil {
				continue
			}
			name, version := parseCargoPackage(string(cargoData))
			if name == "" {
				name = filepath.Base(dir)
			}
			relPath, _ := filepath.Rel(repoPath, dir)
			pkgs = append(pkgs, RawPackage{
				Name:    name,
				Path:    relPath,
				PkgType: "cargo",
				Version: version,
			})
		}
	}
	return pkgs, nil
}

func parseCargoPackage(content string) (string, string) {
	var name, version string
	inPackage := false
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "[package]" {
			inPackage = true
			continue
		}
		if strings.HasPrefix(line, "[") && line != "[package]" {
			if inPackage {
				break
			}
			continue
		}
		if !inPackage {
			continue
		}
		if strings.HasPrefix(line, "name") {
			name = extractTomlStringValue(line)
		}
		if strings.HasPrefix(line, "version") {
			version = extractTomlStringValue(line)
		}
	}
	return name, version
}

func extractTomlStringValue(line string) string {
	idx := strings.Index(line, "=")
	if idx < 0 {
		return ""
	}
	val := strings.TrimSpace(line[idx+1:])
	val = strings.Trim(val, `"'`)
	return val
}

// --- Python packages ---

func detectPythonPackages(repoPath string) ([]RawPackage, error) {
	var pkgs []RawPackage

	// Strategy 1: pyproject.toml with [tool.poetry.packages].
	pyData, err := os.ReadFile(filepath.Join(repoPath, "pyproject.toml"))
	if err == nil {
		content := string(pyData)
		if strings.Contains(content, "[tool.poetry.packages]") ||
			strings.Contains(content, "packages") {
			// Scan for include = [...] patterns.
			scanner := bufio.NewScanner(strings.NewReader(content))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "include") {
					val := extractTomlStringValue(line)
					if val != "" {
						absPath := filepath.Join(repoPath, val)
						if info, err := os.Stat(absPath); err == nil && info.IsDir() {
							pkgs = append(pkgs, RawPackage{
								Name:    filepath.Base(val),
								Path:    val,
								PkgType: "python",
							})
						}
					}
				}
			}
		}
	}

	// Strategy 2: packages/ directory with multiple setup.py or pyproject.toml.
	packagesDir := filepath.Join(repoPath, "packages")
	entries, err := os.ReadDir(packagesDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			subDir := filepath.Join(packagesDir, entry.Name())
			hasSetupPy := fileExists(filepath.Join(subDir, "setup.py"))
			hasPyproject := fileExists(filepath.Join(subDir, "pyproject.toml"))
			if hasSetupPy || hasPyproject {
				name := entry.Name()
				version := ""
				if hasPyproject {
					name, version = parsePyprojectName(filepath.Join(subDir, "pyproject.toml"), name)
				}
				relPath := filepath.Join("packages", entry.Name())
				pkgs = append(pkgs, RawPackage{
					Name:    name,
					Path:    relPath,
					PkgType: "python",
					Version: version,
				})
			}
		}
	}

	if len(pkgs) == 0 {
		return nil, nil
	}
	return pkgs, nil
}

func parsePyprojectName(path, fallback string) (string, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback, ""
	}
	var name, version string
	inProject := false
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "[project]" || line == "[tool.poetry]" {
			inProject = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			if inProject {
				break
			}
			continue
		}
		if !inProject {
			continue
		}
		if strings.HasPrefix(line, "name") {
			name = extractTomlStringValue(line)
		}
		if strings.HasPrefix(line, "version") {
			version = extractTomlStringValue(line)
		}
	}
	if name == "" {
		name = fallback
	}
	return name, version
}

// --- Gradle multi-project ---

var gradleIncludeRe = regexp.MustCompile(`include\s*\(?\s*['":]([^'")]+)`)

func detectGradleProjects(repoPath string) ([]RawPackage, error) {
	var content string
	for _, name := range []string{"settings.gradle.kts", "settings.gradle"} {
		data, err := os.ReadFile(filepath.Join(repoPath, name))
		if err == nil {
			content = string(data)
			break
		}
	}
	if content == "" {
		return nil, nil
	}

	var pkgs []RawPackage
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, "include") {
			continue
		}
		matches := gradleIncludeRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			proj := strings.TrimSpace(m[1])
			proj = strings.TrimPrefix(proj, ":")
			// Also handle comma-separated includes in a single call.
			for _, part := range strings.Split(proj, ",") {
				part = strings.TrimSpace(part)
				part = strings.Trim(part, `'"`)
				part = strings.TrimPrefix(part, ":")
				if part == "" {
					continue
				}
				// Gradle convention: ":foo:bar" maps to "foo/bar" on disk.
				dirPath := strings.ReplaceAll(part, ":", string(filepath.Separator))
				pkgs = append(pkgs, RawPackage{
					Name:    part,
					Path:    dirPath,
					PkgType: "gradle",
				})
			}
		}
	}

	if len(pkgs) == 0 {
		return nil, nil
	}
	return pkgs, nil
}

// --- helpers ---

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// --- handlers ---

// DetectAndSync discovers sub-packages in a project's repo and persists them
// to monorepo_packages. Old entries for the project are removed first.
//
// POST /api/projects/{id}/detect-packages
func (h *Handlers) DetectAndSync(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var wsID uuid.UUID
	var localPath *string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT p.workspace_id, p.local_path
		FROM projects p
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&wsID, &localPath)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if localPath == nil || *localPath == "" {
		http.Error(w, "project not yet cloned", http.StatusConflict)
		return
	}

	detected, err := DetectPackages(*localPath)
	if err != nil {
		h.Logger.Error("detect packages failed", "project", pid, "err", err)
		http.Error(w, "detection failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Remove old entries for this project.
	if _, err := tx.Exec(r.Context(),
		`DELETE FROM monorepo_packages WHERE project_id = $1`, pid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out := make([]Package, 0, len(detected))
	for _, raw := range detected {
		var pkg Package
		err := tx.QueryRow(r.Context(), `
			INSERT INTO monorepo_packages (project_id, name, path, pkg_type, version)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, project_id, name, path, pkg_type, version, detected_at
		`, pid, raw.Name, raw.Path, raw.PkgType, raw.Version).Scan(
			&pkg.ID, &pkg.ProjectID, &pkg.Name, &pkg.Path,
			&pkg.PkgType, &pkg.Version, &pkg.DetectedAt)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, pkg)
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// ListPackages returns all packages for a project, with the number of tasks
// mapped to each package.
//
// GET /api/projects/{id}/packages
func (h *Handlers) ListPackages(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT p.workspace_id
		FROM projects p
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT mp.id, mp.project_id, mp.name, mp.path, mp.pkg_type,
		       mp.version, mp.detected_at,
		       COUNT(pts.id) AS task_count
		FROM monorepo_packages mp
		LEFT JOIN package_task_scopes pts ON pts.package_id = mp.id
		WHERE mp.project_id = $1
		GROUP BY mp.id
		ORDER BY mp.path
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Package{}
	for rows.Next() {
		var pkg Package
		if err := rows.Scan(&pkg.ID, &pkg.ProjectID, &pkg.Name, &pkg.Path,
			&pkg.PkgType, &pkg.Version, &pkg.DetectedAt, &pkg.TaskCount); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, pkg)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GetPackage returns a single package with its task scope mappings.
//
// GET /api/packages/{id}
func (h *Handlers) GetPackage(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pkgID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Load package and verify workspace membership.
	var pkg Package
	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT mp.id, mp.project_id, mp.name, mp.path, mp.pkg_type,
		       mp.version, mp.detected_at, p.workspace_id
		FROM monorepo_packages mp
		JOIN projects p ON p.id = mp.project_id
		WHERE mp.id = $1
	`, pkgID).Scan(&pkg.ID, &pkg.ProjectID, &pkg.Name, &pkg.Path,
		&pkg.PkgType, &pkg.Version, &pkg.DetectedAt, &wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Fetch task scopes.
	scopeRows, err := h.Pool.Query(r.Context(), `
		SELECT pts.id, pts.package_id, pts.task_id, t.name, pts.created_at
		FROM package_task_scopes pts
		JOIN tasks t ON t.id = pts.task_id
		WHERE pts.package_id = $1
		ORDER BY t.name
	`, pkgID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer scopeRows.Close()

	scopes := []PackageTaskScope{}
	for scopeRows.Next() {
		var s PackageTaskScope
		if err := scopeRows.Scan(&s.ID, &s.PackageID, &s.TaskID,
			&s.TaskName, &s.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		scopes = append(scopes, s)
	}
	if err := scopeRows.Err(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Package Package            `json:"package"`
		Scopes  []PackageTaskScope `json:"scopes"`
	}{
		Package: pkg,
		Scopes:  scopes,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// AddTaskScope creates a mapping between a package and a task.
//
// POST /api/packages/{id}/tasks
func (h *Handlers) AddTaskScope(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pkgID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Verify package exists and user has access.
	var pkgProjectID uuid.UUID
	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT mp.project_id, p.workspace_id
		FROM monorepo_packages mp
		JOIN projects p ON p.id = mp.project_id
		WHERE mp.id = $1
	`, pkgID).Scan(&pkgProjectID, &wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body struct {
		TaskID uuid.UUID `json:"task_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.TaskID == uuid.Nil {
		http.Error(w, "task_id required", http.StatusBadRequest)
		return
	}

	// Validate task belongs to the same project.
	var taskProjectID uuid.UUID
	err = h.Pool.QueryRow(r.Context(),
		`SELECT project_id FROM tasks WHERE id = $1`, body.TaskID,
	).Scan(&taskProjectID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if taskProjectID != pkgProjectID {
		http.Error(w, "task does not belong to the same project", http.StatusBadRequest)
		return
	}

	var scope PackageTaskScope
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO package_task_scopes (package_id, task_id)
		VALUES ($1, $2)
		RETURNING id, package_id, task_id, created_at
	`, pkgID, body.TaskID).Scan(&scope.ID, &scope.PackageID, &scope.TaskID, &scope.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "package_task_scopes_package_id_task_id_key") {
			http.Error(w, "scope already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(scope)
}

// RemoveTaskScope deletes a package-to-task mapping.
//
// DELETE /api/package-task-scopes/{id}
func (h *Handlers) RemoveTaskScope(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	scopeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Auth: scope -> package -> project -> workspace membership.
	var n int
	err = h.Pool.QueryRow(r.Context(), `
		SELECT 1 FROM package_task_scopes pts
		JOIN monorepo_packages mp ON mp.id = pts.package_id
		JOIN projects p ON p.id = mp.project_id
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE pts.id = $1 AND m.user_id = $2
	`, scopeID, uid).Scan(&n)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := h.Pool.Exec(r.Context(),
		`DELETE FROM package_task_scopes WHERE id = $1`, scopeID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AutoMapTasks examines each package in a project and looks for tasks whose
// raw_command or source references the package path or name. Matching pairs
// are inserted as package_task_scopes.
//
// POST /api/projects/{id}/auto-map-tasks
func (h *Handlers) AutoMapTasks(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT p.workspace_id
		FROM projects p
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Load packages for this project.
	pkgRows, err := h.Pool.Query(r.Context(), `
		SELECT id, name, path FROM monorepo_packages WHERE project_id = $1
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer pkgRows.Close()

	type pkgInfo struct {
		id   uuid.UUID
		name string
		path string
	}
	var packages []pkgInfo
	for pkgRows.Next() {
		var p pkgInfo
		if err := pkgRows.Scan(&p.id, &p.name, &p.path); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		packages = append(packages, p)
	}
	if err := pkgRows.Err(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Load tasks for this project.
	taskRows, err := h.Pool.Query(r.Context(), `
		SELECT id, source, name, raw_command FROM tasks WHERE project_id = $1
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer taskRows.Close()

	type taskInfo struct {
		id         uuid.UUID
		source     string
		name       string
		rawCommand string
	}
	var tasks []taskInfo
	for taskRows.Next() {
		var t taskInfo
		if err := taskRows.Scan(&t.id, &t.source, &t.name, &t.rawCommand); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}
	if err := taskRows.Err(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Match packages to tasks by checking if the task's raw_command or source
	// contains the package path or name.
	mapped := 0
	for _, pkg := range packages {
		for _, task := range tasks {
			searchText := strings.ToLower(task.rawCommand + " " + task.source + " " + task.name)
			pkgPath := strings.ToLower(pkg.path)
			pkgName := strings.ToLower(pkg.name)

			if !strings.Contains(searchText, pkgPath) &&
				!strings.Contains(searchText, pkgName) {
				continue
			}

			tag, err := h.Pool.Exec(r.Context(), `
				INSERT INTO package_task_scopes (package_id, task_id)
				VALUES ($1, $2)
				ON CONFLICT (package_id, task_id) DO NOTHING
			`, pkg.id, task.id)
			if err != nil {
				h.Logger.Error("auto-map insert failed",
					"package", pkg.id, "task", task.id, "err", err)
				continue
			}
			if tag.RowsAffected() > 0 {
				mapped++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{
		"mapped": mapped,
	})
}
