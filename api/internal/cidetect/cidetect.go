package cidetect

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

type CIConfig struct {
	ID         uuid.UUID `json:"id"`
	ProjectID  uuid.UUID `json:"project_id"`
	CISystem   string    `json:"ci_system"`
	FilePath   string    `json:"file_path"`
	DetectedAt time.Time `json:"detected_at"`
}

type Handlers struct {
	Pool *pgxpool.Pool
}

func DetectCIConfigs(repoPath string) []CIConfig {
	var configs []CIConfig

	wfDir := filepath.Join(repoPath, ".github", "workflows")
	if entries, err := os.ReadDir(wfDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := filepath.Ext(e.Name())
			if ext == ".yml" || ext == ".yaml" {
				configs = append(configs, CIConfig{
					CISystem: "github-actions",
					FilePath: ".github/workflows/" + e.Name(),
				})
			}
		}
	}

	for _, check := range []struct {
		path     string
		ciSystem string
	}{
		{".gitlab-ci.yml", "gitlab-ci"},
		{"Jenkinsfile", "jenkins"},
		{filepath.Join(".circleci", "config.yml"), "circleci"},
		{"bitbucket-pipelines.yml", "bitbucket"},
	} {
		p := filepath.Join(repoPath, check.path)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			configs = append(configs, CIConfig{
				CISystem: check.ciSystem,
				FilePath: check.path,
			})
		}
	}

	return configs
}

func PersistCIConfigs(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger, projectID uuid.UUID, repoPath string) {
	configs := DetectCIConfigs(repoPath)

	tx, err := pool.Begin(ctx)
	if err != nil {
		logger.Warn("cidetect begin tx", "err", err)
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM pipeline_configs WHERE project_id = $1`, projectID); err != nil {
		logger.Warn("cidetect delete", "err", err)
		return
	}

	for _, c := range configs {
		_, err := tx.Exec(ctx, `
			INSERT INTO pipeline_configs (project_id, ci_system, file_path)
			VALUES ($1, $2, $3)
		`, projectID, c.CISystem, c.FilePath)
		if err != nil {
			logger.Warn("cidetect insert", "err", err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Warn("cidetect commit", "err", err)
		return
	}
	logger.Info("ci configs detected", "project", projectID, "count", len(configs))
}

// GET /api/projects/{id}/pipeline-configs
func (h *Handlers) ListByProject(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if !userOwnsProject(r.Context(), h.Pool, uid, projectID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, project_id, ci_system, file_path, detected_at
		FROM pipeline_configs
		WHERE project_id = $1
		ORDER BY ci_system, file_path
	`, projectID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []CIConfig{}
	for rows.Next() {
		var c CIConfig
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.CISystem, &c.FilePath, &c.DetectedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, c)
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
