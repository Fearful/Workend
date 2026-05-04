// Package quota tracks per-user repo-volume usage and enforces quotas
// before clones.
//
// Storage attribution: a workspace's bytes are charged to the workspace
// creator (workspaces.user_id). With sharing (Stage 18), members can use a
// workspace without consuming their own quota — the creator pays.
package quota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

type Usage struct {
	UserID     uuid.UUID `json:"user_id"`
	UsedBytes  int64     `json:"used_bytes"`
	QuotaBytes int64     `json:"quota_bytes"`
	Workspaces []WSUsage `json:"workspaces"`
}

type WSUsage struct {
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	Bytes         int64     `json:"bytes"`
}

type Handlers struct {
	Pool      *pgxpool.Pool
	ReposRoot string
}

// GET /api/me/usage
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	var quotaBytes int64
	if err := h.Pool.QueryRow(r.Context(),
		`SELECT quota_bytes FROM users WHERE id = $1`, uid).Scan(&quotaBytes); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, name FROM workspaces WHERE user_id = $1
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	usage := Usage{UserID: uid, QuotaBytes: quotaBytes, Workspaces: []WSUsage{}}
	for rows.Next() {
		var w WSUsage
		if err := rows.Scan(&w.WorkspaceID, &w.WorkspaceName); err != nil {
			continue
		}
		w.Bytes = dirSize(filepath.Join(h.ReposRoot, w.WorkspaceID.String()))
		usage.UsedBytes += w.Bytes
		usage.Workspaces = append(usage.Workspaces, w)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(usage)
}

// dirSize sums regular-file sizes recursively. Returns 0 on missing dir.
func dirSize(root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

// CheckOrErr returns ErrQuotaExceeded if the workspace creator has used at
// or above their quota. Cheap-ish — does a filesystem walk per call. Run
// this from the clone path before kicking off cloneAsync.
func CheckOrErr(ctx context.Context, pool *pgxpool.Pool, reposRoot string, workspaceID uuid.UUID) error {
	var creatorID uuid.UUID
	var quotaBytes int64
	err := pool.QueryRow(ctx, `
		SELECT w.user_id, u.quota_bytes
		FROM workspaces w
		JOIN users u ON u.id = w.user_id
		WHERE w.id = $1
	`, workspaceID).Scan(&creatorID, &quotaBytes)
	if err != nil {
		return fmt.Errorf("load quota: %w", err)
	}
	// Sum across all workspaces this user owns.
	rows, err := pool.Query(ctx, `SELECT id FROM workspaces WHERE user_id = $1`, creatorID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var used int64
	for rows.Next() {
		var wsID uuid.UUID
		if err := rows.Scan(&wsID); err == nil {
			used += dirSize(filepath.Join(reposRoot, wsID.String()))
		}
	}
	if used >= quotaBytes {
		return ErrQuotaExceeded
	}
	return nil
}

var ErrQuotaExceeded = errors.New("user storage quota exceeded")
