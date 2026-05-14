package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// ShellSession represents a row in the shell_sessions table.
type ShellSession struct {
	ID        uuid.UUID  `json:"id"`
	SandboxID uuid.UUID  `json:"sandbox_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Shell     string     `json:"shell"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	ExecURL   string     `json:"exec_url,omitempty"` // hydrated, not a DB column
}

const (
	ShellStatusActive = "active"
	ShellStatusClosed = "closed"
)

// ShellHandlers holds dependencies for shell session HTTP handlers.
type ShellHandlers struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
	WebURL string // base URL for generating exec URLs
}

type createSessionReq struct {
	Shell string `json:"shell"`
}

type execCommandReq struct {
	Command string `json:"command"`
}

type execCommandResp struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

// CreateSession creates a new shell session attached to a sandbox.
// POST /api/sandboxes/{id}/shell
func (sh *ShellHandlers) CreateSession(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	sbID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid sandbox id", http.StatusBadRequest)
		return
	}

	// Check workspace membership via sandbox.
	wsID, err := sandboxWorkspace(r.Context(), sh.Pool, sbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "sandbox not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), sh.Pool, uid, wsID); err != nil {
		http.Error(w, "not a workspace member", http.StatusForbidden)
		return
	}

	// Sandbox must be running.
	status, err := sandboxStatus(r.Context(), sh.Pool, sbID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if status != StatusRunning {
		http.Error(w, "sandbox is not running", http.StatusConflict)
		return
	}

	var req createSessionReq
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
	}
	if req.Shell == "" {
		req.Shell = "/bin/sh"
	}

	var sess ShellSession
	err = sh.Pool.QueryRow(r.Context(), `
		INSERT INTO shell_sessions (sandbox_id, user_id, shell)
		VALUES ($1, $2, $3)
		RETURNING id, sandbox_id, user_id, shell, status, created_at, ended_at
	`, sbID, uid, req.Shell).
		Scan(&sess.ID, &sess.SandboxID, &sess.UserID, &sess.Shell,
			&sess.Status, &sess.CreatedAt, &sess.EndedAt)
	if err != nil {
		sh.Logger.Error("shell session insert failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sess.ExecURL = fmt.Sprintf("%s/api/shell-sessions/%s/exec", sh.WebURL, sess.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sess)
}

// ExecCommand simulates command execution in a sandbox shell session.
// POST /api/shell-sessions/{id}/exec
func (sh *ShellHandlers) ExecCommand(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	sessID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	sess, err := sh.fetchSession(r.Context(), sessID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if sess.UserID != uid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if sess.Status != ShellStatusActive {
		http.Error(w, "session is closed", http.StatusConflict)
		return
	}

	// Verify the sandbox is still running.
	status, err := sandboxStatus(r.Context(), sh.Pool, sess.SandboxID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if status != StatusRunning {
		http.Error(w, "sandbox is no longer running", http.StatusConflict)
		return
	}

	var req execCommandReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Command == "" {
		http.Error(w, "command required", http.StatusBadRequest)
		return
	}

	// Stub: In production this would pipe through Dagger Container.WithExec().
	// For now return a placeholder acknowledging the command.
	resp := execCommandResp{
		Output:   fmt.Sprintf("[stub] command received: %s", req.Command),
		ExitCode: 0,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// CloseSession marks a shell session as closed.
// POST /api/shell-sessions/{id}/close
func (sh *ShellHandlers) CloseSession(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	sessID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	sess, err := sh.fetchSession(r.Context(), sessID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if sess.UserID != uid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	now := time.Now()
	if _, err := sh.Pool.Exec(r.Context(), `
		UPDATE shell_sessions SET status = $1, ended_at = $2
		WHERE id = $3
	`, ShellStatusClosed, now, sessID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sess.Status = ShellStatusClosed
	sess.EndedAt = &now
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sess)
}

// ListSessions returns all shell sessions for a sandbox.
// GET /api/sandboxes/{id}/shell-sessions
func (sh *ShellHandlers) ListSessions(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	sbID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid sandbox id", http.StatusBadRequest)
		return
	}

	wsID, err := sandboxWorkspace(r.Context(), sh.Pool, sbID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "sandbox not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), sh.Pool, uid, wsID); err != nil {
		http.Error(w, "not a workspace member", http.StatusForbidden)
		return
	}

	rows, err := sh.Pool.Query(r.Context(), `
		SELECT id, sandbox_id, user_id, shell, status, created_at, ended_at
		FROM shell_sessions
		WHERE sandbox_id = $1
		ORDER BY created_at DESC
	`, sbID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []ShellSession{}
	for rows.Next() {
		var s ShellSession
		if err := rows.Scan(&s.ID, &s.SandboxID, &s.UserID, &s.Shell,
			&s.Status, &s.CreatedAt, &s.EndedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, s)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// --- helpers ---

// sandboxWorkspace returns the workspace_id for a given sandbox.
func sandboxWorkspace(ctx context.Context, pool *pgxpool.Pool, sbID uuid.UUID) (uuid.UUID, error) {
	var wsID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT workspace_id FROM sandboxes WHERE id = $1
	`, sbID).Scan(&wsID)
	return wsID, err
}

// fetchSession loads a shell session by ID.
func (sh *ShellHandlers) fetchSession(ctx context.Context, sessID uuid.UUID) (*ShellSession, error) {
	var s ShellSession
	err := sh.Pool.QueryRow(ctx, `
		SELECT id, sandbox_id, user_id, shell, status, created_at, ended_at
		FROM shell_sessions
		WHERE id = $1
	`, sessID).Scan(&s.ID, &s.SandboxID, &s.UserID, &s.Shell,
		&s.Status, &s.CreatedAt, &s.EndedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
