// Package graph provides knowledge-graph features: service dependency
// mapping, incident timelines, and change-impact prediction.
package graph

import (
	"bufio"
	"encoding/json"
	"fmt"
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

// ServiceDependency is one discovered dependency between projects.
type ServiceDependency struct {
	ID              uuid.UUID  `json:"id"`
	SourceProjectID uuid.UUID  `json:"source_project_id"`
	TargetProjectID *uuid.UUID `json:"target_project_id"`
	DepType         string     `json:"dep_type"`
	Reference       string     `json:"reference"`
	FilePath        string     `json:"file_path"`
	DiscoveredAt    time.Time  `json:"discovered_at"`
}

// DependencyGraph is the full node/edge representation of a workspace or
// single-project dependency view.
type DependencyGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode is a project participating in the dependency graph.
type GraphNode struct {
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectName string    `json:"project_name"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
}

// GraphEdge is a directed dependency from one project to another.
type GraphEdge struct {
	Source    uuid.UUID `json:"source"`
	Target   uuid.UUID `json:"target"`
	DepType  string    `json:"dep_type"`
	Reference string   `json:"reference"`
}

// Handlers holds shared dependencies for all graph endpoints.
type Handlers struct {
	Pool      *pgxpool.Pool
	ReposRoot string
	Logger    *slog.Logger
}

// --- scanner ---

// RawDep is a single dependency discovered by scanning files in a repo.
type RawDep struct {
	DepType   string
	Reference string
	FilePath  string
}

var dockerFromRe = regexp.MustCompile(`(?i)^FROM\s+(\S+)`)

// ScanProject walks a repo directory and discovers dependencies from
// Dockerfiles, docker-compose files, and dagger.json.
func ScanProject(repoPath string) ([]RawDep, error) {
	var deps []RawDep

	dockerfileDeps, err := scanDockerfiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("scan dockerfiles: %w", err)
	}
	deps = append(deps, dockerfileDeps...)

	composeDeps, err := scanComposeFiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("scan compose: %w", err)
	}
	deps = append(deps, composeDeps...)

	daggerDeps, err := scanDaggerJSON(repoPath)
	if err != nil {
		return nil, fmt.Errorf("scan dagger: %w", err)
	}
	deps = append(deps, daggerDeps...)

	return deps, nil
}

func scanDockerfiles(repoPath string) ([]RawDep, error) {
	var deps []RawDep
	candidates := []string{"Dockerfile"}

	// Also walk for Dockerfile.* variants at top level.
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), "Dockerfile.") {
			candidates = append(candidates, e.Name())
		}
	}

	for _, name := range candidates {
		filePath := filepath.Join(repoPath, name)
		found, err := parseDockerfile(filePath, name)
		if err != nil {
			continue // skip unreadable files
		}
		deps = append(deps, found...)
	}
	return deps, nil
}

func parseDockerfile(absPath, relPath string) ([]RawDep, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var deps []RawDep
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		m := dockerFromRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		ref := m[1]
		// Skip build-arg placeholders like ${BASE_IMAGE}.
		if strings.HasPrefix(ref, "$") {
			continue
		}
		deps = append(deps, RawDep{
			DepType:   "dockerfile_from",
			Reference: ref,
			FilePath:  relPath,
		})
	}
	return deps, scanner.Err()
}

var (
	composeImageRe   = regexp.MustCompile(`^\s+image:\s*['"]?(\S+?)['"]?\s*$`)
	composeBuildRe   = regexp.MustCompile(`^\s+context:\s*['"]?(\S+?)['"]?\s*$`)
	composeBuildShortRe = regexp.MustCompile(`^\s+build:\s*['"]?(\S+?)['"]?\s*$`)
)

func scanComposeFiles(repoPath string) ([]RawDep, error) {
	var deps []RawDep
	names := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, name := range names {
		absPath := filepath.Join(repoPath, name)
		if _, err := os.Stat(absPath); err != nil {
			continue
		}
		found, err := parseComposeFile(absPath, name)
		if err != nil {
			continue // skip unreadable
		}
		deps = append(deps, found...)
	}
	return deps, nil
}

func parseComposeFile(absPath, relPath string) ([]RawDep, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var deps []RawDep
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := composeImageRe.FindStringSubmatch(line); m != nil {
			deps = append(deps, RawDep{
				DepType:   "compose_image",
				Reference: m[1],
				FilePath:  relPath,
			})
			continue
		}
		if m := composeBuildRe.FindStringSubmatch(line); m != nil {
			ref := m[1]
			if ref != "." && ref != "./" {
				deps = append(deps, RawDep{
					DepType:   "compose_service",
					Reference: ref,
					FilePath:  relPath,
				})
			}
			continue
		}
		if m := composeBuildShortRe.FindStringSubmatch(line); m != nil {
			ref := m[1]
			if ref != "." && ref != "./" {
				deps = append(deps, RawDep{
					DepType:   "compose_service",
					Reference: ref,
					FilePath:  relPath,
				})
			}
		}
	}
	return deps, scanner.Err()
}

func scanDaggerJSON(repoPath string) ([]RawDep, error) {
	absPath := filepath.Join(repoPath, "dagger.json")
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var doc struct {
		Dependencies []json.RawMessage `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse dagger.json: %w", err)
	}

	var deps []RawDep
	for _, raw := range doc.Dependencies {
		// Dependencies can be strings or objects with a "name" field.
		var s string
		if json.Unmarshal(raw, &s) == nil {
			deps = append(deps, RawDep{
				DepType:   "dagger_module",
				Reference: s,
				FilePath:  "dagger.json",
			})
			continue
		}
		var obj struct {
			Name   string `json:"name"`
			Source string `json:"source"`
		}
		if json.Unmarshal(raw, &obj) == nil {
			ref := obj.Source
			if ref == "" {
				ref = obj.Name
			}
			if ref != "" {
				deps = append(deps, RawDep{
					DepType:   "dagger_module",
					Reference: ref,
					FilePath:  "dagger.json",
				})
			}
		}
	}
	return deps, nil
}

// --- handlers ---

// ScanProjectDeps discovers dependencies in a project's cloned repo and
// persists them to the service_dependencies table.
//
// POST /api/projects/{id}/scan-dependencies
func (h *Handlers) ScanProjectDeps(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var wsID uuid.UUID
	var clonePath *string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT p.workspace_id, p.local_path
		FROM projects p
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&wsID, &clonePath)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if clonePath == nil || *clonePath == "" {
		http.Error(w, "project not yet cloned", http.StatusConflict)
		return
	}

	rawDeps, err := ScanProject(*clonePath)
	if err != nil {
		h.Logger.Error("scan dependencies failed", "project", pid, "err", err)
		http.Error(w, "scan failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	if _, err := tx.Exec(r.Context(),
		`DELETE FROM service_dependencies WHERE source_project_id = $1`, pid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out := make([]ServiceDependency, 0, len(rawDeps))
	for _, rd := range rawDeps {
		var dep ServiceDependency
		err := tx.QueryRow(r.Context(), `
			INSERT INTO service_dependencies (source_project_id, dep_type, reference, file_path)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (source_project_id, dep_type, reference) DO UPDATE
				SET file_path = EXCLUDED.file_path, discovered_at = now()
			RETURNING id, source_project_id, target_project_id, dep_type, reference, file_path, discovered_at
		`, pid, rd.DepType, rd.Reference, rd.FilePath).Scan(
			&dep.ID, &dep.SourceProjectID, &dep.TargetProjectID,
			&dep.DepType, &dep.Reference, &dep.FilePath, &dep.DiscoveredAt)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, dep)
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GetDependencyMap builds a full dependency graph for all projects in a
// workspace, resolving references to target projects where possible.
//
// GET /api/workspaces/{id}/dependency-map
func (h *Handlers) GetDependencyMap(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Fetch all projects in the workspace (nodes).
	projRows, err := h.Pool.Query(r.Context(), `
		SELECT id, name, workspace_id FROM projects WHERE workspace_id = $1
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer projRows.Close()

	nodes := []GraphNode{}
	projectsByName := map[string]uuid.UUID{}
	projectsByURL := map[string]uuid.UUID{}
	for projRows.Next() {
		var n GraphNode
		if err := projRows.Scan(&n.ProjectID, &n.ProjectName, &n.WorkspaceID); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		nodes = append(nodes, n)
		projectsByName[strings.ToLower(n.ProjectName)] = n.ProjectID
	}
	if err := projRows.Err(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Also index by repo_url for resolution.
	urlRows, err := h.Pool.Query(r.Context(), `
		SELECT id, git_url FROM projects WHERE workspace_id = $1 AND git_url != ''
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer urlRows.Close()
	for urlRows.Next() {
		var id uuid.UUID
		var url string
		if err := urlRows.Scan(&id, &url); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		projectsByURL[strings.ToLower(url)] = id
	}

	// Fetch all service_dependencies for workspace projects.
	depRows, err := h.Pool.Query(r.Context(), `
		SELECT sd.source_project_id, sd.target_project_id, sd.dep_type, sd.reference
		FROM service_dependencies sd
		JOIN projects p ON p.id = sd.source_project_id
		WHERE p.workspace_id = $1
	`, wsID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer depRows.Close()

	edges := []GraphEdge{}
	for depRows.Next() {
		var e GraphEdge
		var targetID *uuid.UUID
		if err := depRows.Scan(&e.Source, &targetID, &e.DepType, &e.Reference); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Attempt resolution if target is not already set.
		if targetID != nil {
			e.Target = *targetID
		} else {
			resolved := resolveTarget(e.Reference, projectsByName, projectsByURL)
			if resolved != nil {
				e.Target = *resolved
				// Best-effort update in DB (fire-and-forget).
				_, _ = h.Pool.Exec(r.Context(), `
					UPDATE service_dependencies SET target_project_id = $1
					WHERE source_project_id = $2 AND dep_type = $3 AND reference = $4
				`, *resolved, e.Source, e.DepType, e.Reference)
			}
		}

		edges = append(edges, e)
	}

	graph := DependencyGraph{Nodes: nodes, Edges: edges}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(graph)
}

// resolveTarget attempts to match a reference string against known project
// names and repo URLs. Returns nil if no match is found.
func resolveTarget(reference string, byName, byURL map[string]uuid.UUID) *uuid.UUID {
	lower := strings.ToLower(reference)

	// Strip tag/digest from image references: "myapp:latest" -> "myapp".
	imageName := lower
	if idx := strings.LastIndex(imageName, ":"); idx > 0 {
		imageName = imageName[:idx]
	}
	// Also strip registry prefix: "registry.example.com/myapp" -> "myapp".
	if idx := strings.LastIndex(imageName, "/"); idx >= 0 {
		imageName = imageName[idx+1:]
	}

	// Try exact name match.
	if id, ok := byName[imageName]; ok {
		return &id
	}
	// Try full reference as name.
	if id, ok := byName[lower]; ok {
		return &id
	}
	// Try URL match.
	if id, ok := byURL[lower]; ok {
		return &id
	}

	// Strip path prefix for compose_service references: "./other-service" -> "other-service".
	cleaned := strings.TrimPrefix(lower, "./")
	cleaned = strings.TrimPrefix(cleaned, "../")
	cleaned = strings.TrimSuffix(cleaned, "/")
	if id, ok := byName[cleaned]; ok {
		return &id
	}

	return nil
}

// GetProjectDeps returns dependencies for a single project: both outbound
// (this project depends on X) and inbound (X depends on this project).
//
// GET /api/projects/{id}/dependencies-graph
func (h *Handlers) GetProjectDeps(w http.ResponseWriter, r *http.Request) {
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

	// Collect all project IDs involved in dependencies.
	depRows, err := h.Pool.Query(r.Context(), `
		SELECT sd.source_project_id, sd.target_project_id, sd.dep_type, sd.reference
		FROM service_dependencies sd
		WHERE sd.source_project_id = $1
		   OR sd.target_project_id = $1
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer depRows.Close()

	projectIDs := map[uuid.UUID]struct{}{pid: {}}
	edges := []GraphEdge{}
	for depRows.Next() {
		var e GraphEdge
		var targetID *uuid.UUID
		if err := depRows.Scan(&e.Source, &targetID, &e.DepType, &e.Reference); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if targetID != nil {
			e.Target = *targetID
			projectIDs[*targetID] = struct{}{}
		}
		projectIDs[e.Source] = struct{}{}
		edges = append(edges, e)
	}

	// Fetch nodes for all involved projects.
	ids := make([]uuid.UUID, 0, len(projectIDs))
	for id := range projectIDs {
		ids = append(ids, id)
	}

	nodes := []GraphNode{}
	if len(ids) > 0 {
		nodeRows, err := h.Pool.Query(r.Context(), `
			SELECT id, name, workspace_id FROM projects WHERE id = ANY($1)
		`, ids)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer nodeRows.Close()
		for nodeRows.Next() {
			var n GraphNode
			if err := nodeRows.Scan(&n.ProjectID, &n.ProjectName, &n.WorkspaceID); err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			nodes = append(nodes, n)
		}
	}

	graph := DependencyGraph{Nodes: nodes, Edges: edges}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(graph)
}
