package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

// ActivityEntry represents one event in the team activity feed.
type ActivityEntry struct {
	ID          int64          `json:"id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	UserID      uuid.UUID      `json:"user_id"`
	UserName    string         `json:"user_name"`
	EventType   string         `json:"event_type"`
	EntityType  string         `json:"entity_type"`
	EntityID    *uuid.UUID     `json:"entity_id"`
	Summary     string         `json:"summary"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
}

// ActivityFeedHandlers serves the team activity feed for a workspace.
type ActivityFeedHandlers struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
}

const (
	feedDefaultLimit = 50
	feedMaxLimit     = 200
	sseHeartbeatSec  = 15
	ssePollSec       = 2
)

// GetFeed returns recent activity entries for a workspace with cursor-based
// pagination.
//
// GET /api/workspaces/{id}/feed?limit=50&before=<id>
func (h *ActivityFeedHandlers) GetFeed(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}

	if _, err := RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	limit := feedDefaultLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= feedMaxLimit {
			limit = n
		}
	}

	args := []any{wsID, limit}
	whereBefore := ""
	if v := r.URL.Query().Get("before"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			args = append(args, n)
			whereBefore = " AND f.id < $3"
		}
	}

	q := `
		SELECT f.id, f.workspace_id, f.user_id,
		       COALESCE(u.display_name, '') AS user_name,
		       f.event_type, f.entity_type, f.entity_id,
		       f.summary, f.metadata, f.created_at
		FROM activity_feed f
		JOIN users u ON u.id = f.user_id
		WHERE f.workspace_id = $1` + whereBefore + `
		ORDER BY f.id DESC
		LIMIT $2
	`
	rows, err := h.Pool.Query(r.Context(), q, args...)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []ActivityEntry{}
	for rows.Next() {
		var e ActivityEntry
		var rawMeta []byte
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.UserID, &e.UserName,
			&e.EventType, &e.EntityType, &e.EntityID,
			&e.Summary, &rawMeta, &e.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		e.Metadata = decodeMetadata(rawMeta)
		out = append(out, e)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// GetFeedSSE streams real-time activity events via Server-Sent Events.
// Polls the activity_feed table every 2 seconds for new entries.
//
// GET /api/workspaces/{id}/feed/stream
func (h *ActivityFeedHandlers) GetFeedSSE(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}

	if _, err := RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Start from the latest known ID so we only stream new events.
	var lastSeenID int64
	_ = h.Pool.QueryRow(r.Context(),
		`SELECT COALESCE(MAX(id), 0) FROM activity_feed WHERE workspace_id = $1`, wsID,
	).Scan(&lastSeenID)

	poll := time.NewTicker(time.Duration(ssePollSec) * time.Second)
	defer poll.Stop()
	heartbeat := time.NewTicker(time.Duration(sseHeartbeatSec) * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-poll.C:
			rows, err := h.Pool.Query(r.Context(), `
				SELECT f.id, f.workspace_id, f.user_id,
				       COALESCE(u.display_name, '') AS user_name,
				       f.event_type, f.entity_type, f.entity_id,
				       f.summary, f.metadata, f.created_at
				FROM activity_feed f
				JOIN users u ON u.id = f.user_id
				WHERE f.workspace_id = $1 AND f.id > $2
				ORDER BY f.id ASC
				LIMIT 50
			`, wsID, lastSeenID)
			if err != nil {
				continue
			}

			for rows.Next() {
				var e ActivityEntry
				var rawMeta []byte
				if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.UserID, &e.UserName,
					&e.EventType, &e.EntityType, &e.EntityID,
					&e.Summary, &rawMeta, &e.CreatedAt); err != nil {
					continue
				}
				e.Metadata = decodeMetadata(rawMeta)
				lastSeenID = e.ID

				data, err := json.Marshal(e)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "event: activity\ndata: %s\n\n", data)
			}
			rows.Close()
			flusher.Flush()
		}
	}
}

// RecordActivity inserts an activity entry into the feed. It is a fire-and-
// forget helper intended to be called from other handlers (run complete,
// project created, member added, etc.). Failures are logged but never
// propagated to the caller.
func RecordActivity(ctx context.Context, pool *pgxpool.Pool, workspaceID, userID uuid.UUID, eventType, entityType string, entityID *uuid.UUID, summary string, metadata map[string]any) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO activity_feed (workspace_id, user_id, event_type, entity_type, entity_id, summary, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
	`, workspaceID, userID, eventType, entityType, entityID, summary, string(metaJSON))
	if err != nil {
		// Fire-and-forget: log but do not fail the caller.
		slog.Warn("activity_feed insert failed",
			"workspace_id", workspaceID,
			"event_type", eventType,
			"err", err)
	}
}

// decodeMetadata unmarshals a JSONB column into a map. Returns an empty map
// on any decode error so callers can always iterate safely.
func decodeMetadata(raw []byte) map[string]any {
	m := map[string]any{}
	if len(raw) == 0 {
		return m
	}
	_ = json.Unmarshal(raw, &m)
	return m
}
