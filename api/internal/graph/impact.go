package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

// --- types ---

// TaskFileInput maps a glob pattern to a task, indicating what file changes
// should trigger that task.
type TaskFileInput struct {
	ID           uuid.UUID `json:"id"`
	TaskID       uuid.UUID `json:"task_id"`
	Pattern      string    `json:"pattern"`
	InputType    string    `json:"input_type"`
	AutoDetected bool      `json:"auto_detected"`
	CreatedAt    time.Time `json:"created_at"`
}

// ImpactResult describes which tasks are affected by a set of file changes.
type ImpactResult struct {
	ChangedFiles  []string       `json:"changed_files"`
	AffectedTasks []AffectedTask `json:"affected_tasks"`
	TotalAffected int            `json:"total_affected"`
}

// AffectedTask is one task that matched changed files.
type AffectedTask struct {
	TaskID         uuid.UUID `json:"task_id"`
	TaskName       string    `json:"task_name"`
	ProjectID      uuid.UUID `json:"project_id"`
	ProjectName    string    `json:"project_name"`
	MatchedPattern string    `json:"matched_pattern"`
	MatchedFiles   []string  `json:"matched_files"`
}

// --- glob matcher ---

// matchGlob checks whether filePath matches a glob pattern. Supports standard
// filepath.Match patterns plus "**" for recursive directory matching.
func matchGlob(pattern, filePath string) bool {
	// Handle ** by splitting on it and matching segments.
	if strings.Contains(pattern, "**") {
		return matchDoublestar(pattern, filePath)
	}
	matched, _ := filepath.Match(pattern, filePath)
	if matched {
		return true
	}
	// Also try matching against just the basename for simple patterns
	// like "Dockerfile" or "package.json".
	if !strings.Contains(pattern, "/") && !strings.Contains(pattern, string(filepath.Separator)) {
		matched, _ = filepath.Match(pattern, filepath.Base(filePath))
		return matched
	}
	return false
}

// matchDoublestar splits a pattern on "**" and checks if the file path
// satisfies the prefix and suffix constraints with any directory nesting.
func matchDoublestar(pattern, filePath string) bool {
	parts := strings.SplitN(pattern, "**", 2)
	prefix := parts[0]
	suffix := ""
	if len(parts) > 1 {
		suffix = parts[1]
	}

	// Remove trailing/leading separators from prefix/suffix.
	prefix = strings.TrimRight(prefix, "/")
	suffix = strings.TrimLeft(suffix, "/")

	// If prefix is non-empty, the file must live under that prefix.
	if prefix != "" {
		if !strings.HasPrefix(filePath, prefix+"/") && filePath != prefix {
			return false
		}
	}

	// If suffix is empty, any file under prefix matches.
	if suffix == "" {
		return true
	}

	// The suffix may itself be a glob, so check it against the remaining path
	// at every directory level.
	remaining := filePath
	if prefix != "" {
		remaining = strings.TrimPrefix(filePath, prefix+"/")
	}

	// Try matching suffix against every possible tail of the remaining path.
	segments := strings.Split(remaining, "/")
	for i := range segments {
		tail := strings.Join(segments[i:], "/")
		matched, _ := filepath.Match(suffix, tail)
		if matched {
			return true
		}
		// Also try matching just the filename for patterns like "**/*.go".
		if i == len(segments)-1 {
			matched, _ = filepath.Match(suffix, segments[i])
			if matched {
				return true
			}
		}
	}
	return false
}

// --- handlers ---

// AutoDetectInputs analyzes tasks in a project and auto-detects file input
// patterns based on each task's source type.
//
// POST /api/projects/{id}/detect-inputs
func (h *Handlers) AutoDetectInputs(w http.ResponseWriter, r *http.Request) {
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

	// Fetch tasks for this project.
	taskRows, err := h.Pool.Query(r.Context(), `
		SELECT id, source FROM tasks WHERE project_id = $1
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer taskRows.Close()

	type taskInfo struct {
		id     uuid.UUID
		source string
	}
	var tasks []taskInfo
	for taskRows.Next() {
		var t taskInfo
		if err := taskRows.Scan(&t.id, &t.source); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}

	out := []TaskFileInput{}
	for _, t := range tasks {
		patterns := patternsForSource(t.source)
		for _, pat := range patterns {
			inputType := classifyPattern(pat)
			var fi TaskFileInput
			err := h.Pool.QueryRow(r.Context(), `
				INSERT INTO task_file_inputs (task_id, pattern, input_type, auto_detected)
				VALUES ($1, $2, $3, true)
				ON CONFLICT (task_id, pattern) DO UPDATE
					SET input_type = EXCLUDED.input_type, auto_detected = true
				RETURNING id, task_id, pattern, input_type, auto_detected, created_at
			`, t.id, pat, inputType).Scan(
				&fi.ID, &fi.TaskID, &fi.Pattern, &fi.InputType, &fi.AutoDetected, &fi.CreatedAt)
			if err != nil {
				h.Logger.Error("insert task_file_input failed", "task", t.id, "pattern", pat, "err", err)
				continue
			}
			out = append(out, fi)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// patternsForSource returns the default file-input patterns for a given task
// source type.
func patternsForSource(source string) []string {
	switch {
	case source == "npm" || source == "pnpm" || source == "yarn" || source == "bun" || source == "node":
		return []string{"package.json", "package-lock.json", "src/**"}
	case source == "gomod" || source == "go":
		return []string{"go.mod", "go.sum", "**/*.go"}
	case source == "dockerfile" || source == "docker":
		return []string{"Dockerfile", "docker-compose.yml"}
	case source == "cargo":
		return []string{"Cargo.toml", "Cargo.lock", "src/**"}
	case source == "just":
		return []string{"justfile", "Makefile"}
	case source == "makefile" || source == "make":
		return []string{"Makefile", "justfile"}
	default:
		return []string{"*"}
	}
}

// classifyPattern guesses the input type from the pattern.
func classifyPattern(pattern string) string {
	lower := strings.ToLower(pattern)
	switch {
	case strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".yml") ||
		strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".toml") ||
		strings.HasSuffix(lower, ".lock"):
		return "config"
	case strings.HasSuffix(lower, "/**") || lower == "*":
		return "directory"
	default:
		return "file"
	}
}

// ListInputs returns all file-input patterns for a task.
//
// GET /api/tasks/{id}/file-inputs
func (h *Handlers) ListInputs(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if !userCanAccessTask(r.Context(), h.Pool, uid, tid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, task_id, pattern, input_type, auto_detected, created_at
		FROM task_file_inputs
		WHERE task_id = $1
		ORDER BY created_at
	`, tid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []TaskFileInput{}
	for rows.Next() {
		var fi TaskFileInput
		if err := rows.Scan(&fi.ID, &fi.TaskID, &fi.Pattern, &fi.InputType,
			&fi.AutoDetected, &fi.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, fi)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// AddInput creates a new file-input pattern for a task.
//
// POST /api/tasks/{id}/file-inputs
func (h *Handlers) AddInput(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	tid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if !userCanAccessTask(r.Context(), h.Pool, uid, tid) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var body struct {
		Pattern   string `json:"pattern"`
		InputType string `json:"input_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	body.Pattern = strings.TrimSpace(body.Pattern)
	if body.Pattern == "" {
		http.Error(w, "pattern required", http.StatusBadRequest)
		return
	}
	if len(body.Pattern) > 200 {
		http.Error(w, "pattern max 200 chars", http.StatusBadRequest)
		return
	}

	if body.InputType == "" {
		body.InputType = "file"
	}
	if body.InputType != "file" && body.InputType != "directory" && body.InputType != "config" {
		http.Error(w, "input_type must be file|directory|config", http.StatusBadRequest)
		return
	}

	var fi TaskFileInput
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO task_file_inputs (task_id, pattern, input_type, auto_detected)
		VALUES ($1, $2, $3, false)
		RETURNING id, task_id, pattern, input_type, auto_detected, created_at
	`, tid, body.Pattern, body.InputType).Scan(
		&fi.ID, &fi.TaskID, &fi.Pattern, &fi.InputType, &fi.AutoDetected, &fi.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "task_file_inputs_task_id_pattern_key") {
			http.Error(w, "pattern already exists for this task", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(fi)
}

// RemoveInput deletes a file-input pattern.
//
// DELETE /api/file-inputs/{id}
func (h *Handlers) RemoveInput(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	fiID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Auth: file-input -> task -> project -> workspace membership.
	var n int
	err = h.Pool.QueryRow(r.Context(), `
		SELECT 1 FROM task_file_inputs fi
		JOIN tasks t ON t.id = fi.task_id
		JOIN projects p ON p.id = t.project_id
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE fi.id = $1 AND m.user_id = $2
	`, fiID, uid).Scan(&n)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := h.Pool.Exec(r.Context(),
		`DELETE FROM task_file_inputs WHERE id = $1`, fiID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PredictImpact determines which tasks are affected by a list of changed files.
//
// POST /api/projects/{id}/predict-impact
func (h *Handlers) PredictImpact(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		ChangedFiles []string `json:"changed_files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(body.ChangedFiles) == 0 {
		http.Error(w, "changed_files required", http.StatusBadRequest)
		return
	}

	result, err := h.computeImpact(r.Context(), pid, body.ChangedFiles)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// PredictImpactFromDiff reads a git diff between branches and predicts impact.
//
// POST /api/projects/{id}/predict-impact/diff
func (h *Handlers) PredictImpactFromDiff(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		BaseBranch string `json:"base_branch"`
		HeadBranch string `json:"head_branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.BaseBranch = strings.TrimSpace(body.BaseBranch)
	body.HeadBranch = strings.TrimSpace(body.HeadBranch)
	if body.BaseBranch == "" || body.HeadBranch == "" {
		http.Error(w, "base_branch and head_branch required", http.StatusBadRequest)
		return
	}

	// Validate branch names: reject shell injection characters.
	if !isSafeBranchName(body.BaseBranch) || !isSafeBranchName(body.HeadBranch) {
		http.Error(w, "invalid branch name", http.StatusBadRequest)
		return
	}

	cmd := exec.CommandContext(r.Context(),
		"git", "diff", "--name-only", body.BaseBranch+".."+body.HeadBranch)
	cmd.Dir = *clonePath
	out, err := cmd.Output()
	if err != nil {
		http.Error(w, "git diff failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	changedFiles := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			changedFiles = append(changedFiles, line)
		}
	}

	if len(changedFiles) == 0 {
		result := ImpactResult{
			ChangedFiles:  changedFiles,
			AffectedTasks: []AffectedTask{},
			TotalAffected: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	result, err := h.computeImpact(r.Context(), pid, changedFiles)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// computeImpact loads task file inputs for a project and matches them against
// the changed files list.
func (h *Handlers) computeImpact(ctx context.Context, pid uuid.UUID, changedFiles []string) (*ImpactResult, error) {
	rows, err := h.Pool.Query(ctx, `
		SELECT fi.task_id, fi.pattern, t.name AS task_name, p.name AS project_name, t.project_id
		FROM task_file_inputs fi
		JOIN tasks t ON t.id = fi.task_id
		JOIN projects p ON p.id = t.project_id
		WHERE t.project_id = $1
		ORDER BY t.name, fi.pattern
	`, pid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type inputRow struct {
		taskID      uuid.UUID
		pattern     string
		taskName    string
		projectName string
		projectID   uuid.UUID
	}
	var inputs []inputRow
	for rows.Next() {
		var ir inputRow
		if err := rows.Scan(&ir.taskID, &ir.pattern, &ir.taskName, &ir.projectName, &ir.projectID); err != nil {
			return nil, err
		}
		inputs = append(inputs, ir)
	}

	// Group by task and find matches.
	type taskMatch struct {
		taskID      uuid.UUID
		taskName    string
		projectID   uuid.UUID
		projectName string
		pattern     string
		files       []string
	}
	matches := map[uuid.UUID]*taskMatch{}

	for _, input := range inputs {
		var matchedFiles []string
		for _, f := range changedFiles {
			if matchGlob(input.pattern, f) {
				matchedFiles = append(matchedFiles, f)
			}
		}
		if len(matchedFiles) == 0 {
			continue
		}

		existing, ok := matches[input.taskID]
		if !ok {
			matches[input.taskID] = &taskMatch{
				taskID:      input.taskID,
				taskName:    input.taskName,
				projectID:   input.projectID,
				projectName: input.projectName,
				pattern:     input.pattern,
				files:       matchedFiles,
			}
		} else {
			// Merge additional matched files, keep first pattern.
			seen := map[string]struct{}{}
			for _, f := range existing.files {
				seen[f] = struct{}{}
			}
			for _, f := range matchedFiles {
				if _, ok := seen[f]; !ok {
					existing.files = append(existing.files, f)
				}
			}
		}
	}

	affected := make([]AffectedTask, 0, len(matches))
	for _, m := range matches {
		affected = append(affected, AffectedTask{
			TaskID:         m.taskID,
			TaskName:       m.taskName,
			ProjectID:      m.projectID,
			ProjectName:    m.projectName,
			MatchedPattern: m.pattern,
			MatchedFiles:   m.files,
		})
	}

	return &ImpactResult{
		ChangedFiles:  changedFiles,
		AffectedTasks: affected,
		TotalAffected: len(affected),
	}, nil
}

// isSafeBranchName rejects branch names that could cause shell injection.
func isSafeBranchName(name string) bool {
	for _, c := range name {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '-' || c == '_' || c == '/' || c == '.':
		default:
			return false
		}
	}
	return len(name) > 0 && len(name) <= 256
}

// userCanAccessTask checks workspace membership through the task chain.
func userCanAccessTask(ctx context.Context, pool *pgxpool.Pool, uid, tid uuid.UUID) bool {
	var n int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM tasks t
		JOIN projects p ON p.id = t.project_id
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE t.id = $1 AND m.user_id = $2
	`, tid, uid).Scan(&n)
	return err == nil
}
