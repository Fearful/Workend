package graph

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// --- types ---

// CommitEntry is one row in the commit_history table.
type CommitEntry struct {
	ID           uuid.UUID `json:"id"`
	ProjectID    uuid.UUID `json:"project_id"`
	SHA          string    `json:"sha"`
	AuthorName   string    `json:"author_name"`
	AuthorEmail  string    `json:"author_email"`
	Message      string    `json:"message"`
	CommittedAt  time.Time `json:"committed_at"`
	FilesChanged int       `json:"files_changed"`
	Insertions   int       `json:"insertions"`
	Deletions    int       `json:"deletions"`
}

// BlameTimelineEntry pairs a commit with aggregated run statistics.
type BlameTimelineEntry struct {
	Commit        CommitEntry `json:"commit"`
	RunCount      int         `json:"run_count"`
	PassCount     int         `json:"pass_count"`
	FailCount     int         `json:"fail_count"`
	AvgDurationMs int64       `json:"avg_duration_ms"`
}

// BlameTimeline is the paginated response for a project's commit history
// correlated with run outcomes.
type BlameTimeline struct {
	ProjectID    uuid.UUID            `json:"project_id"`
	ProjectName  string               `json:"project_name"`
	TotalCommits int                  `json:"total_commits"`
	Entries      []BlameTimelineEntry `json:"entries"`
}

// RunSummary is a lightweight run representation returned with commit details.
type RunSummary struct {
	ID         uuid.UUID  `json:"id"`
	TaskName   string     `json:"task_name"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	DurationMs int64      `json:"duration_ms"`
}

// CommitDetail is the response for GET /api/commits/{id}.
type CommitDetail struct {
	Commit CommitEntry  `json:"commit"`
	Runs   []RunSummary `json:"runs"`
}

// --- git log parser ---

var shortstatRe = regexp.MustCompile(
	`(\d+)\s+files?\s+changed(?:,\s+(\d+)\s+insertions?\(\+\))?(?:,\s+(\d+)\s+deletions?\(-\))?`,
)

// parseGitLog runs git log with --shortstat and parses each commit entry.
// Returns up to limit commits. The repoPath must be an absolute path to the
// cloned repository.
func parseGitLog(repoPath string, limit int) ([]CommitEntry, error) {
	if limit <= 0 {
		limit = 200
	}

	cmd := exec.Command(
		"git", "-C", repoPath, "log",
		fmt.Sprintf("--format=%%H|%%an|%%ae|%%aI|%%s"),
		"--shortstat",
		"-n", strconv.Itoa(limit),
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	var commits []CommitEntry
	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	var current *CommitEntry
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Try to parse as a commit format line (SHA|author|email|date|message).
		parts := strings.SplitN(line, "|", 5)
		if len(parts) == 5 && len(parts[0]) == 40 && isHexString(parts[0]) {
			// Flush previous commit if it exists.
			if current != nil {
				commits = append(commits, *current)
			}

			committedAt, parseErr := time.Parse(time.RFC3339, parts[3])
			if parseErr != nil {
				committedAt = time.Now()
			}

			current = &CommitEntry{
				SHA:         parts[0],
				AuthorName:  parts[1],
				AuthorEmail: parts[2],
				CommittedAt: committedAt,
				Message:     parts[4],
			}
			continue
		}

		// Try to parse as a shortstat line.
		if current != nil {
			if m := shortstatRe.FindStringSubmatch(line); m != nil {
				current.FilesChanged, _ = strconv.Atoi(m[1])
				if m[2] != "" {
					current.Insertions, _ = strconv.Atoi(m[2])
				}
				if m[3] != "" {
					current.Deletions, _ = strconv.Atoi(m[3])
				}
			}
		}
	}

	// Flush last commit.
	if current != nil {
		commits = append(commits, *current)
	}

	return commits, scanner.Err()
}

// isHexString returns true if every character is a hex digit.
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return len(s) > 0
}

// --- handlers ---

// SyncCommitHistory parses the git log of a project's cloned repo and upserts
// commits into the commit_history table.
//
// POST /api/projects/{id}/sync-commits
func (h *Handlers) SyncCommitHistory(w http.ResponseWriter, r *http.Request) {
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

	commits, err := parseGitLog(*localPath, 200)
	if err != nil {
		h.Logger.Error("parse git log failed", "project", pid, "err", err)
		http.Error(w, "git log failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	inserted := 0
	for _, c := range commits {
		tag, err := h.Pool.Exec(r.Context(), `
			INSERT INTO commit_history
				(project_id, sha, author_name, author_email, message, committed_at,
				 files_changed, insertions, deletions)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (project_id, sha) DO NOTHING
		`, pid, c.SHA, c.AuthorName, c.AuthorEmail, c.Message, c.CommittedAt,
			c.FilesChanged, c.Insertions, c.Deletions)
		if err != nil {
			h.Logger.Error("insert commit failed", "sha", c.SHA, "err", err)
			continue
		}
		if tag.RowsAffected() > 0 {
			inserted++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{
		"synced":   len(commits),
		"inserted": inserted,
	})
}

// GetBlameTimeline returns commit history for a project correlated with run
// statistics. Each entry shows a commit plus how many runs targeted that
// commit SHA, how many passed/failed, and the average duration.
//
// GET /api/projects/{id}/blame-timeline?limit=50
func (h *Handlers) GetBlameTimeline(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var wsID uuid.UUID
	var projectName string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT p.workspace_id, p.name
		FROM projects p
		JOIN workspace_members m ON m.workspace_id = p.workspace_id
		WHERE p.id = $1 AND m.user_id = $2
	`, pid, uid).Scan(&wsID, &projectName)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 200 {
		limit = 200
	}

	// Count total commits for this project.
	var totalCommits int
	err = h.Pool.QueryRow(r.Context(),
		`SELECT COUNT(*) FROM commit_history WHERE project_id = $1`, pid,
	).Scan(&totalCommits)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Query commits with LEFT JOIN to runs for aggregation.
	rows, err := h.Pool.Query(r.Context(), `
		SELECT
			ch.id, ch.project_id, ch.sha, ch.author_name, ch.author_email,
			ch.message, ch.committed_at, ch.files_changed, ch.insertions, ch.deletions,
			COUNT(r.id)                                             AS run_count,
			COUNT(r.id) FILTER (WHERE r.status = 'succeeded')      AS pass_count,
			COUNT(r.id) FILTER (WHERE r.status = 'failed')         AS fail_count,
			COALESCE(
				AVG(EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000)
				FILTER (WHERE r.finished_at IS NOT NULL AND r.started_at IS NOT NULL),
				0
			)::BIGINT                                               AS avg_duration_ms
		FROM commit_history ch
		LEFT JOIN runs r ON r.commit_sha = ch.sha AND r.project_id = ch.project_id
		WHERE ch.project_id = $1
		GROUP BY ch.id
		ORDER BY ch.committed_at DESC
		LIMIT $2
	`, pid, limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	entries := []BlameTimelineEntry{}
	for rows.Next() {
		var e BlameTimelineEntry
		if err := rows.Scan(
			&e.Commit.ID, &e.Commit.ProjectID, &e.Commit.SHA,
			&e.Commit.AuthorName, &e.Commit.AuthorEmail, &e.Commit.Message,
			&e.Commit.CommittedAt, &e.Commit.FilesChanged,
			&e.Commit.Insertions, &e.Commit.Deletions,
			&e.RunCount, &e.PassCount, &e.FailCount, &e.AvgDurationMs,
		); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	timeline := BlameTimeline{
		ProjectID:    pid,
		ProjectName:  projectName,
		TotalCommits: totalCommits,
		Entries:      entries,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(timeline)
}

// GetCommitDetail returns a single commit entry together with all runs that
// used the same commit SHA.
//
// GET /api/commits/{id}
func (h *Handlers) GetCommitDetail(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	commitID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Load the commit and verify workspace membership.
	var commit CommitEntry
	var wsID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT ch.id, ch.project_id, ch.sha, ch.author_name, ch.author_email,
		       ch.message, ch.committed_at, ch.files_changed, ch.insertions, ch.deletions,
		       p.workspace_id
		FROM commit_history ch
		JOIN projects p ON p.id = ch.project_id
		WHERE ch.id = $1
	`, commitID).Scan(
		&commit.ID, &commit.ProjectID, &commit.SHA,
		&commit.AuthorName, &commit.AuthorEmail, &commit.Message,
		&commit.CommittedAt, &commit.FilesChanged, &commit.Insertions, &commit.Deletions,
		&wsID,
	)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Fetch runs that match this commit SHA within the same project.
	runRows, err := h.Pool.Query(r.Context(), `
		SELECT r.id, t.name, r.status, r.started_at, r.finished_at,
		       COALESCE(
		           EXTRACT(EPOCH FROM (r.finished_at - r.started_at)) * 1000,
		           0
		       )::BIGINT AS duration_ms
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.commit_sha = $1 AND r.project_id = $2
		ORDER BY r.started_at DESC
	`, commit.SHA, commit.ProjectID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer runRows.Close()

	runs := []RunSummary{}
	for runRows.Next() {
		var rs RunSummary
		if err := runRows.Scan(&rs.ID, &rs.TaskName, &rs.Status,
			&rs.StartedAt, &rs.FinishedAt, &rs.DurationMs); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		runs = append(runs, rs)
	}
	if err := runRows.Err(); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	detail := CommitDetail{
		Commit: commit,
		Runs:   runs,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(detail)
}
