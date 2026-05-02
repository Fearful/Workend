// Package admin owns endpoints scoped to users with is_admin=true.
package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/audit"
	"workend/api/internal/auth"
)

type Handlers struct {
	Pool  *pgxpool.Pool
	Audit *audit.Logger
}

// RequireAdmin is middleware that 403s any non-admin user.
// Wrap inside auth.RequireUser.
func RequireAdmin(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := auth.UserID(r.Context())
			var isAdmin bool
			err := pool.QueryRow(r.Context(), `SELECT is_admin FROM users WHERE id = $1`, uid).Scan(&isAdmin)
			if err != nil || !isAdmin {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type userSummary struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	IsAdmin       bool      `json:"is_admin"`
	CreatedAt     time.Time `json:"created_at"`
	WorkspaceCount int      `json:"workspace_count"`
	ProjectCount   int      `json:"project_count"`
	RunCount       int      `json:"run_count"`
}

// GET /api/admin/users
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	h.Audit.Record(r.Context(), uid, audit.AdminUserList, "", "", r.RemoteAddr, nil)

	rows, err := h.Pool.Query(r.Context(), `
		SELECT u.id, u.email, u.display_name, u.is_admin, u.created_at,
		       COALESCE(ws.cnt, 0) AS workspace_count,
		       COALESCE(p.cnt, 0)  AS project_count,
		       COALESCE(r.cnt, 0)  AS run_count
		FROM users u
		LEFT JOIN (SELECT user_id, COUNT(*) AS cnt FROM workspaces GROUP BY user_id) ws ON ws.user_id = u.id
		LEFT JOIN (SELECT w.user_id, COUNT(*) AS cnt
		           FROM projects pr JOIN workspaces w ON w.id = pr.workspace_id
		           GROUP BY w.user_id) p ON p.user_id = u.id
		LEFT JOIN (SELECT w.user_id, COUNT(*) AS cnt
		           FROM runs ru
		           JOIN projects pr ON pr.id = ru.project_id
		           JOIN workspaces w ON w.id = pr.workspace_id
		           GROUP BY w.user_id) r ON r.user_id = u.id
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []userSummary{}
	for rows.Next() {
		var u userSummary
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.IsAdmin, &u.CreatedAt,
			&u.WorkspaceCount, &u.ProjectCount, &u.RunCount); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, u)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type auditEntry struct {
	ID         int64           `json:"id"`
	OccurredAt time.Time       `json:"occurred_at"`
	ActorID    *uuid.UUID      `json:"actor_id"`
	ActorEmail *string         `json:"actor_email"`
	Action     string          `json:"action"`
	TargetKind *string         `json:"target_kind"`
	TargetID   *string         `json:"target_id"`
	IP         *string         `json:"ip"`
	Metadata   json.RawMessage `json:"metadata"`
}

// GET /api/admin/audit-log?limit=N
func (h *Handlers) AuditLog(w http.ResponseWriter, r *http.Request) {
	limit := 200
	rows, err := h.Pool.Query(r.Context(), `
		SELECT a.id, a.occurred_at, a.actor_id, u.email,
		       a.action, a.target_kind, a.target_id, a.ip, a.metadata
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.actor_id
		ORDER BY a.occurred_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []auditEntry{}
	for rows.Next() {
		var e auditEntry
		var meta []byte
		if err := rows.Scan(&e.ID, &e.OccurredAt, &e.ActorID, &e.ActorEmail,
			&e.Action, &e.TargetKind, &e.TargetID, &e.IP, &meta); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if len(meta) > 0 {
			e.Metadata = meta
		} else {
			e.Metadata = json.RawMessage("null")
		}
		out = append(out, e)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
