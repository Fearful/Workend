// Package audit records security-relevant events. Best-effort: failures
// don't propagate to the caller (we'd rather lose an audit record than fail
// a user action).
package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	UserSignup     = "user.signup"
	UserLogin      = "user.login"
	UserLogout     = "user.logout"
	WorkspaceCreate = "workspace.create"
	WorkspaceDelete = "workspace.delete"
	ProjectCreate  = "project.create"
	ProjectDelete  = "project.delete"
	ProjectSync    = "project.sync"
	RunStart       = "run.start"
	RunCancel      = "run.cancel"
	RunComplete    = "run.complete"
	GitHubConnect  = "user.github.connect"
	GitHubDisconnect = "user.github.disconnect"
	AdminUserList  = "admin.user.list"
)

type Logger struct {
	Pool *pgxpool.Pool
	Log  *slog.Logger
}

// Record writes an audit event. actorID may be uuid.Nil for unauthenticated
// or system events. metadata is optional structured context.
func (l *Logger) Record(ctx context.Context, actorID uuid.UUID, action, targetKind, targetID, ip string, metadata map[string]any) {
	var actor any = actorID
	if actorID == uuid.Nil {
		actor = nil
	}
	var meta any = nil
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			meta = b
		}
	}
	if _, err := l.Pool.Exec(ctx, `
		INSERT INTO audit_log (actor_id, action, target_kind, target_id, ip, metadata)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6)
	`, actor, action, targetKind, targetID, ip, meta); err != nil {
		l.Log.Warn("audit insert failed", "action", action, "err", err)
	}
}
