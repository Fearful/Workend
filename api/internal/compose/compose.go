package compose

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	wdagger "workend/api/internal/dagger"
	"workend/api/internal/secret"

	"dagger.io/dagger"
	"gopkg.in/yaml.v3"
)

type Handlers struct {
	Pool   *pgxpool.Pool
	Dagger *wdagger.Client
	Box    *secret.Box
	Logger *slog.Logger
}

type ComposeConfig struct {
	ComposeFile string     `json:"compose_file"`
	Services    []string   `json:"services"`
	EnvVars     []EnvVar   `json:"env_vars"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at"`
}

type EnvVar struct {
	Key          string `json:"key"`
	DefaultValue string `json:"default_value"`
}

type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Ports  string `json:"ports"`
}

func DetectComposeFile(repoPath string) (string, bool) {
	for _, name := range []string{"compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml"} {
		p := filepath.Join(repoPath, name)
		if _, err := os.Stat(p); err == nil {
			return name, true
		}
	}
	return "", false
}

func DetectEnvExample(repoPath string) []EnvVar {
	for _, name := range []string{".env.example", ".env.sample"} {
		p := filepath.Join(repoPath, name)
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		var vars []EnvVar
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := ""
			if len(parts) == 2 {
				val = strings.TrimSpace(parts[1])
			}
			if key != "" {
				vars = append(vars, EnvVar{Key: key, DefaultValue: val})
			}
		}
		return vars
	}
	return nil
}

func ParseServices(repoPath, composeFile string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(repoPath, composeFile))
	if err != nil {
		return nil, err
	}

	var doc struct {
		Services map[string]any `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(doc.Services))
	for name := range doc.Services {
		names = append(names, name)
	}
	return names, nil
}

// GET /api/projects/{id}/compose
func (h *Handlers) GetConfig(w http.ResponseWriter, r *http.Request) {
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

	var localPath string
	if err := h.Pool.QueryRow(r.Context(), `SELECT local_path FROM projects WHERE id = $1`, projectID).Scan(&localPath); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	composeFile, found := DetectComposeFile(localPath)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"compose_file": nil})
		return
	}

	services, _ := ParseServices(localPath, composeFile)
	envVars := DetectEnvExample(localPath)

	status := "stopped"
	var startedAt *time.Time
	err = h.Pool.QueryRow(r.Context(), `
		SELECT status, started_at FROM compose_instances WHERE project_id = $1
	`, projectID).Scan(&status, &startedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	cfg := ComposeConfig{
		ComposeFile: composeFile,
		Services:    services,
		EnvVars:     envVars,
		Status:      status,
		StartedAt:   startedAt,
	}
	if cfg.Services == nil {
		cfg.Services = []string{}
	}
	if cfg.EnvVars == nil {
		cfg.EnvVars = []EnvVar{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cfg)
}

// POST /api/projects/{id}/compose/start
func (h *Handlers) Start(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		EnvVars map[string]string `json:"env_vars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	var localPath string
	if err := h.Pool.QueryRow(r.Context(), `SELECT local_path FROM projects WHERE id = $1`, projectID).Scan(&localPath); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	composeFile, found := DetectComposeFile(localPath)
	if !found {
		http.Error(w, "no compose file", http.StatusBadRequest)
		return
	}

	var envEncrypted []byte
	if len(body.EnvVars) > 0 && h.Box != nil {
		envJSON, _ := json.Marshal(body.EnvVars)
		envEncrypted, _ = h.Box.Seal(envJSON)
	}

	projectShort := projectID.String()[:8]

	go func(projID uuid.UUID, repoPath, cFile, projShort string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		_, _ = h.Pool.Exec(ctx, `
			INSERT INTO compose_instances (project_id, compose_file, status, env_vars_encrypted, started_at)
			VALUES ($1, $2, 'starting', $3, now())
			ON CONFLICT (project_id) DO UPDATE SET
				status = 'starting', env_vars_encrypted = EXCLUDED.env_vars_encrypted, started_at = now(), error_message = NULL
		`, projID, cFile, envEncrypted)

		client, err := h.Dagger.Get(ctx)
		if err != nil {
			h.setComposeError(ctx, projID, err)
			return
		}

		ctr := client.Container().From("docker:27-cli").
			WithMountedDirectory("/project", client.Host().Directory(repoPath)).
			WithWorkdir("/project")

		if len(body.EnvVars) > 0 {
			for k, v := range body.EnvVars {
				ctr = ctr.WithEnvVariable(k, v)
			}
		}

		_, err = ctr.WithExec([]string{
			"docker", "compose", "-f", cFile,
			"--project-name", "workend-" + projShort,
			"up", "-d", "--build",
		}).Stdout(ctx)
		if err != nil {
			var execErr *dagger.ExecError
			if errors.As(err, &execErr) {
				msg := strings.TrimSpace(execErr.Stderr)
				if msg == "" {
					msg = strings.TrimSpace(execErr.Stdout)
				}
				if msg != "" {
					err = errors.New(msg)
				}
			}
			h.setComposeError(ctx, projID, err)
			return
		}

		_, _ = h.Pool.Exec(ctx, `
			UPDATE compose_instances SET status = 'running' WHERE project_id = $1
		`, projID)
	}(projectID, localPath, composeFile, projectShort)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "starting"})
}

// POST /api/projects/{id}/compose/stop
func (h *Handlers) Stop(w http.ResponseWriter, r *http.Request) {
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

	var localPath, composeFile string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT p.local_path, ci.compose_file
		FROM projects p JOIN compose_instances ci ON ci.project_id = p.id
		WHERE p.id = $1
	`, projectID).Scan(&localPath, &composeFile)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	projectShort := projectID.String()[:8]

	go func(projID uuid.UUID, repoPath, cFile, projShort string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		client, err := h.Dagger.Get(ctx)
		if err != nil {
			h.Logger.Warn("compose stop dagger", "err", err)
			return
		}

		_, _ = client.Container().From("docker:27-cli").
			WithMountedDirectory("/project", client.Host().Directory(repoPath)).
			WithWorkdir("/project").
			WithExec([]string{
				"docker", "compose", "-f", cFile,
				"--project-name", "workend-" + projShort,
				"down",
			}).Stdout(ctx)

		_, _ = h.Pool.Exec(ctx, `
			UPDATE compose_instances SET status = 'stopped', stopped_at = now() WHERE project_id = $1
		`, projID)
	}(projectID, localPath, composeFile, projectShort)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "stopping"})
}

// GET /api/projects/{id}/compose/status
func (h *Handlers) GetStatus(w http.ResponseWriter, r *http.Request) {
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

	var status string
	var errMsg *string
	var startedAt, stoppedAt *time.Time
	err = h.Pool.QueryRow(r.Context(), `
		SELECT status, error_message, started_at, stopped_at
		FROM compose_instances WHERE project_id = $1
	`, projectID).Scan(&status, &errMsg, &startedAt, &stoppedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out := map[string]any{
		"status":     status,
		"started_at": startedAt,
		"stopped_at": stoppedAt,
	}
	if errMsg != nil {
		out["error"] = *errMsg
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handlers) setComposeError(ctx context.Context, projectID uuid.UUID, err error) {
	msg := err.Error()
	if len(msg) > 500 {
		msg = msg[:500]
	}
	_, _ = h.Pool.Exec(ctx, `
		UPDATE compose_instances SET status = 'error', error_message = $2 WHERE project_id = $1
	`, projectID, msg)
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
