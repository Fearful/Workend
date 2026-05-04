// Package stats computes per-project code statistics by running tokei in a
// Dagger-managed container. Triggered after every project sync.
package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	wdagger "workend/api/internal/dagger"
)

// Lang is the per-language stats stored in the DB and returned to the frontend.
type Lang struct {
	Files    int `json:"files"`
	Lines    int `json:"lines"`
	Code     int `json:"code"`
	Comments int `json:"comments"`
	Blanks   int `json:"blanks"`
}

// tokeiLang matches tokei's actual JSON output per language entry.
type tokeiLang struct {
	Blanks   int           `json:"blanks"`
	Code     int           `json:"code"`
	Comments int           `json:"comments"`
	Reports  []tokeiReport `json:"reports"`
}

type tokeiReport struct {
	Name  string    `json:"name"`
	Stats tokeiStat `json:"stats"`
}

type tokeiStat struct {
	Blanks   int `json:"blanks"`
	Code     int `json:"code"`
	Comments int `json:"comments"`
}

// Result is what Compute returns.
type Result struct {
	TotalFiles int             `json:"total_files"`
	TotalLines int             `json:"total_lines"`
	TotalCode  int             `json:"total_code"`
	Languages  map[string]Lang `json:"languages"`
}

// Compute runs `tokei -o json` in an alpine container against repoPath via
// Dagger. Parses output and returns aggregated counts.
func Compute(ctx context.Context, dc *wdagger.Client, repoPath string) (*Result, error) {
	client, err := dc.Get(ctx)
	if err != nil {
		return nil, err
	}

	out, err := client.Container().
		From("alpine:3.20").
		WithExec([]string{"apk", "add", "--no-cache", "tokei"}).
		WithMountedDirectory("/repo", client.Host().Directory(repoPath)).
		WithWorkdir("/repo").
		WithExec([]string{"tokei", "-o", "json"}).
		Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("tokei exec: %w", err)
	}

	raw := map[string]tokeiLang{}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("parse tokei output: %w", err)
	}

	res := &Result{Languages: map[string]Lang{}}
	for name, tl := range raw {
		lang := Lang{
			Files:    len(tl.Reports),
			Lines:    tl.Code + tl.Comments + tl.Blanks,
			Code:     tl.Code,
			Comments: tl.Comments,
			Blanks:   tl.Blanks,
		}
		if name == "Total" {
			res.TotalLines = lang.Lines
			res.TotalCode = lang.Code
			continue
		}
		res.Languages[name] = lang
		res.TotalFiles += lang.Files
	}
	return res, nil
}

// Run computes and persists stats for a project. Failures are logged but
// don't fail the caller's flow.
func Run(ctx context.Context, dc *wdagger.Client, pool *pgxpool.Pool, logger *slog.Logger, projectID uuid.UUID, repoPath string) {
	res, err := Compute(ctx, dc, repoPath)
	if err != nil {
		logger.Warn("stats compute failed", "project", projectID, "err", err)
		return
	}

	langsJSON, err := json.Marshal(res.Languages)
	if err != nil {
		logger.Warn("stats marshal failed", "project", projectID, "err", err)
		return
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO project_stats (project_id, total_files, total_lines, total_code, languages)
		VALUES ($1, $2, $3, $4, $5)
	`, projectID, res.TotalFiles, res.TotalLines, res.TotalCode, langsJSON); err != nil {
		logger.Warn("stats insert failed", "project", projectID, "err", err)
		return
	}

	logger.Info("stats computed", "project", projectID, "files", res.TotalFiles, "lines", res.TotalLines)
}

// Latest returns the most recent stats record for a project, or nil if none
// computed yet.
func Latest(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) (*Record, error) {
	var rec Record
	var langsJSON []byte
	err := pool.QueryRow(ctx, `
		SELECT id, project_id, computed_at, total_files, total_lines, total_code, languages
		FROM project_stats
		WHERE project_id = $1
		ORDER BY computed_at DESC
		LIMIT 1
	`, projectID).Scan(&rec.ID, &rec.ProjectID, &rec.ComputedAt,
		&rec.TotalFiles, &rec.TotalLines, &rec.TotalCode, &langsJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(langsJSON, &rec.Languages); err != nil {
		return nil, err
	}
	return &rec, nil
}
