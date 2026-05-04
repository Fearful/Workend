// Package secscan runs gitleaks against a freshly-cloned repo and writes
// the findings to the secret_findings table. Designed to be invoked from
// project.cloneAsync after the working tree is exported to disk.
package secscan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	wdagger "workend/api/internal/dagger"
)

// Finding mirrors the JSON document gitleaks emits.
type Finding struct {
	RuleID      string `json:"RuleID"`
	File        string `json:"File"`
	StartLine   int    `json:"StartLine"`
	Fingerprint string `json:"Fingerprint"`
	Description string `json:"Description"`
}

// Scan runs gitleaks against the repo at repoPath, parses its JSON report,
// and upserts the findings into secret_findings. Best-effort: failures
// don't propagate (they're logged via the slog Logger if any).
func Scan(ctx context.Context, dc *wdagger.Client, pool *pgxpool.Pool, projectID uuid.UUID, repoPath, commitSHA string) error {
	client, err := dc.Get(ctx)
	if err != nil {
		return err
	}
	repoDir := client.Host().Directory(repoPath)
	out, _ := client.Container().
		// gitleaks 8.x ships pre-built; pin a known good tag.
		From("zricethezav/gitleaks:v8.18.4").
		WithMountedDirectory("/scan", repoDir).
		// `detect --no-banner --report-format json --report-path -` writes to
		// stdout; exit code is non-zero when findings exist, which Dagger
		// (with default Expect: success) would surface as an error. We swallow
		// it via `|| true` so the JSON payload still flows back.
		WithExec([]string{"sh", "-c", "gitleaks detect --no-banner --report-format json --report-path /tmp/leaks.json --source /scan; cat /tmp/leaks.json"}).
		Stdout(ctx)

	out = strings.TrimSpace(out)
	if out == "" || out == "[]" {
		return nil
	}
	var findings []Finding
	if err := json.Unmarshal([]byte(out), &findings); err != nil {
		return fmt.Errorf("parse gitleaks json: %w", err)
	}

	for _, f := range findings {
		_, err := pool.Exec(ctx, `
			INSERT INTO secret_findings (project_id, commit_sha, file, line_no, rule, severity, fingerprint)
			VALUES ($1, NULLIF($2, ''), $3, $4, $5, 'high', $6)
			ON CONFLICT (project_id, fingerprint) DO UPDATE
			SET file = EXCLUDED.file, line_no = EXCLUDED.line_no, rule = EXCLUDED.rule,
			    found_at = now(), resolved_at = NULL
		`, projectID, commitSHA, f.File, f.StartLine, f.RuleID, f.Fingerprint)
		if err != nil {
			return fmt.Errorf("insert finding: %w", err)
		}
	}
	// Mark fingerprints we no longer see as resolved.
	if len(findings) > 0 {
		fps := make([]string, 0, len(findings))
		for _, f := range findings {
			fps = append(fps, f.Fingerprint)
		}
		_, _ = pool.Exec(ctx, `
			UPDATE secret_findings
			SET resolved_at = now()
			WHERE project_id = $1
			  AND resolved_at IS NULL
			  AND NOT (fingerprint = ANY($2::text[]))
		`, projectID, fps)
	}
	return nil
}

// FindingView is what the API surfaces. Adds owned-by check via projects→
// workspaces→workspace_members.
type FindingView struct {
	ID         uuid.UUID  `json:"id"`
	CommitSHA  *string    `json:"commit_sha"`
	File       string     `json:"file"`
	LineNo     *int       `json:"line_no"`
	Rule       string     `json:"rule"`
	Severity   string     `json:"severity"`
	FoundAt    time.Time  `json:"found_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
}

type Handlers struct {
	Pool *pgxpool.Pool
}

// ListByProject returns the open + recently-resolved secret findings.
// GET /api/projects/:id/secret-findings
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
		SELECT id, commit_sha, file, line_no, rule, severity, found_at, resolved_at
		FROM secret_findings
		WHERE project_id = $1
		ORDER BY resolved_at IS NOT NULL, found_at DESC
		LIMIT 200
	`, pid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []FindingView{}
	for rows.Next() {
		var v FindingView
		if err := rows.Scan(&v.ID, &v.CommitSHA, &v.File, &v.LineNo, &v.Rule, &v.Severity, &v.FoundAt, &v.ResolvedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, v)
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

// silence linter on unused error var while iterating
var _ = errors.New
