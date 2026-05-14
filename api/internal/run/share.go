package run

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"workend/api/internal/auth"
)

// RunShare represents a time-limited, token-based public link to a run.
type RunShare struct {
	ID        uuid.UUID `json:"id"`
	RunID     uuid.UUID `json:"run_id"`
	Token     string    `json:"token"`
	ShareURL  string    `json:"share_url"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	shareTokenBytes  = 32
	defaultShareHrs  = 7 * 24 // 7 days
	minShareHrs      = 1
	maxShareHrs      = 720 // 30 days
)

// CreateShare generates a public read-only link for a run.
//
// POST /api/runs/{id}/share
func (h *Handlers) CreateShare(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	if _, err := h.fetchOwned(r.Context(), uid, runID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	expiresHrs := defaultShareHrs
	if r.ContentLength > 0 && r.ContentLength < 4096 {
		var body struct {
			ExpiresHours int `json:"expires_hours"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.ExpiresHours > 0 {
			if body.ExpiresHours < minShareHrs {
				body.ExpiresHours = minShareHrs
			}
			if body.ExpiresHours > maxShareHrs {
				body.ExpiresHours = maxShareHrs
			}
			expiresHrs = body.ExpiresHours
		}
	}

	token, err := generateShareToken()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(time.Duration(expiresHrs) * time.Hour)

	var share RunShare
	err = h.Pool.QueryRow(r.Context(), `
		INSERT INTO run_shares (run_id, token, created_by, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, run_id, token, expires_at, created_at
	`, runID, token, uid, expiresAt).Scan(
		&share.ID, &share.RunID, &share.Token, &share.ExpiresAt, &share.CreatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	share.ShareURL = h.WebURL + "/shared/runs/" + share.Token

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(share)
}

// RevokeShare deletes a share link. Only the creator can revoke.
//
// DELETE /api/run-shares/{id}
func (h *Handlers) RevokeShare(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	shareID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid share id", http.StatusBadRequest)
		return
	}

	tag, err := h.Pool.Exec(r.Context(),
		`DELETE FROM run_shares WHERE id = $1 AND created_by = $2`, shareID, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListShares returns active (non-expired) share links for a run.
//
// GET /api/runs/{id}/shares
func (h *Handlers) ListShares(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	if _, err := h.fetchOwned(r.Context(), uid, runID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, run_id, token, expires_at, created_at
		FROM run_shares
		WHERE run_id = $1 AND expires_at > now()
		ORDER BY created_at DESC
	`, runID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []RunShare{}
	for rows.Next() {
		var s RunShare
		if err := rows.Scan(&s.ID, &s.RunID, &s.Token, &s.ExpiresAt, &s.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		s.ShareURL = h.WebURL + "/shared/runs/" + s.Token
		out = append(out, s)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GetSharedRun is the PUBLIC endpoint (no auth required) that returns a run
// and its masked log via a share token.
//
// NOTE: This handler must be mounted OUTSIDE the RequireUser middleware group.
//
// GET /api/shared/runs/{token}
func (h *Handlers) GetSharedRun(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	var (
		runID       uuid.UUID
		projectName string
		taskName    string
	)
	err := h.Pool.QueryRow(r.Context(), `
		SELECT s.run_id, p.name, t.name
		FROM run_shares s
		JOIN runs r   ON r.id = s.run_id
		JOIN tasks t  ON t.id = r.task_id
		JOIN projects p ON p.id = r.project_id
		WHERE s.token = $1 AND s.expires_at > now()
	`, token).Scan(&runID, &projectName, &taskName)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Load the run directly (no ownership check — the token is the authz).
	var run Run
	var rawParams []byte
	err = h.Pool.QueryRow(r.Context(), `
		SELECT r.id, r.task_id, r.project_id, r.commit_sha, r.status,
		       r.started_at, r.finished_at, r.exit_code, r.timed_out, r.log_path,
		       r.params, r.attempt, r.parent_run_id,
		       r.cpu_ms, r.mem_peak_bytes, r.net_rx_bytes, r.net_tx_bytes,
		       r.created_at, r.updated_at,
		       t.name, t.source
		FROM runs r
		JOIN tasks t ON t.id = r.task_id
		WHERE r.id = $1
	`, runID).Scan(&run.ID, &run.TaskID, &run.ProjectID, &run.CommitSHA, &run.Status,
		&run.StartedAt, &run.FinishedAt, &run.ExitCode, &run.TimedOut, &run.LogPath,
		&rawParams, &run.Attempt, &run.ParentRunID,
		&run.CPUMs, &run.MemPeakBytes, &run.NetRxBytes, &run.NetTxBytes,
		&run.CreatedAt, &run.UpdatedAt,
		&run.TaskName, &run.TaskSource)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	run.Params = decodeParams(rawParams)

	logContent := ""
	if data, err := os.ReadFile(run.LogPath); err == nil {
		logContent = MaskSecrets(string(data))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"run":          run,
		"log":          logContent,
		"task_name":    taskName,
		"project_name": projectName,
	})
}

// generateShareToken produces a URL-safe, 32-byte random token with no padding.
func generateShareToken() (string, error) {
	b := make([]byte, shareTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
