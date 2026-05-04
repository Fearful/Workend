// Package comment owns per-run comment threads and the mention inbox they
// produce as a side effect.
package comment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

type Comment struct {
	ID              uuid.UUID `json:"id"`
	RunID           uuid.UUID `json:"run_id"`
	UserID          uuid.UUID `json:"user_id"`
	UserDisplayName string    `json:"user_display_name"`
	Body            string    `json:"body"`
	CreatedAt       time.Time `json:"created_at"`
}

type Mention struct {
	ID          int64     `json:"id"`
	RunID       uuid.UUID `json:"run_id"`
	CommentID   uuid.UUID `json:"comment_id"`
	ActorID     uuid.UUID `json:"actor_id"`
	ActorName   string    `json:"actor_name"`
	Body        string    `json:"body"`
	TaskName    string    `json:"task_name"`
	ProjectName string    `json:"project_name"`
	CreatedAt   time.Time `json:"created_at"`
	ReadAt      *time.Time `json:"read_at"`
}

type Handlers struct {
	Pool *pgxpool.Pool
}

// mentionRegex matches @display-name (case-insensitive). Deliberately
// permissive: alphanumerics, dashes, dots, underscores; matched against the
// display_name column (which is what users see). Stops at whitespace or
// punctuation that wouldn't appear in a name.
var mentionRegex = regexp.MustCompile(`@([A-Za-z0-9._-]+)`)

// List returns comments for a run, oldest first. Membership-gated.
// GET /api/runs/:id/comments
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}
	if !h.userOwnsRun(r.Context(), uid, runID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rows, err := h.Pool.Query(r.Context(), `
		SELECT c.id, c.run_id, c.user_id, COALESCE(u.display_name, ''), c.body, c.created_at
		FROM run_comments c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.run_id = $1
		ORDER BY c.created_at
	`, runID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Comment{}
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.RunID, &c.UserID, &c.UserDisplayName, &c.Body, &c.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, c)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Create posts a new comment on a run. Parses @mentions and inserts a row
// per matched workspace member into the mentions table. Self-mentions
// don't fire (no point pinging yourself).
// POST /api/runs/:id/comments  body: {"body": "..."}
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}
	if !h.userOwnsRun(r.Context(), uid, runID) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.Body = strings.TrimSpace(body.Body)
	if body.Body == "" {
		http.Error(w, "body required", http.StatusBadRequest)
		return
	}
	if len(body.Body) > 16*1024 {
		http.Error(w, "body too long (max 16 KiB)", http.StatusBadRequest)
		return
	}

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var c Comment
	err = tx.QueryRow(r.Context(), `
		INSERT INTO run_comments (run_id, user_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, run_id, user_id, body, created_at
	`, runID, uid, body.Body).Scan(&c.ID, &c.RunID, &c.UserID, &c.Body, &c.CreatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// @mention parse: look up workspace members whose display_name matches
	// any captured handle (case-insensitive). Skip the author themselves.
	mentioned := extractMentions(body.Body)
	if len(mentioned) > 0 {
		rows, err := tx.Query(r.Context(), `
			SELECT DISTINCT u.id
			FROM run_comments rc
			JOIN runs r       ON r.id = rc.run_id
			JOIN projects p   ON p.id = r.project_id
			JOIN workspace_members m ON m.workspace_id = p.workspace_id
			JOIN users u      ON u.id = m.user_id
			WHERE rc.id = $1
			  AND lower(u.display_name) = ANY($2::text[])
			  AND u.id <> $3
		`, c.ID, mentioned, uid)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var targetID uuid.UUID
				if rows.Scan(&targetID) != nil {
					continue
				}
				_, _ = tx.Exec(r.Context(), `
					INSERT INTO mentions (user_id, actor_id, run_id, comment_id)
					VALUES ($1, $2, $3, $4)
				`, targetID, uid, runID, c.ID)
			}
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Hydrate display name for the response.
	_ = h.Pool.QueryRow(r.Context(), `SELECT display_name FROM users WHERE id = $1`, uid).Scan(&c.UserDisplayName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

// Delete removes a comment. Only the author or the run-owner can delete.
// DELETE /api/runs/:run_id/comments/:id
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	cid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	tag, err := h.Pool.Exec(r.Context(), `
		DELETE FROM run_comments WHERE id = $1 AND user_id = $2
	`, cid, uid)
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

// ListMentions returns the user's mention inbox (unread first by default).
// GET /api/me/mentions?unread_only=true&limit=N
func (h *Handlers) ListMentions(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	whereUnread := ""
	if unreadOnly {
		whereUnread = " AND m.read_at IS NULL"
	}

	rows, err := h.Pool.Query(r.Context(), `
		SELECT m.id, m.run_id, m.comment_id, m.actor_id, COALESCE(au.display_name, ''),
		       c.body, t.name, p.name, m.created_at, m.read_at
		FROM mentions m
		JOIN run_comments c ON c.id = m.comment_id
		JOIN runs r        ON r.id = m.run_id
		JOIN tasks t       ON t.id = r.task_id
		JOIN projects p    ON p.id = r.project_id
		LEFT JOIN users au ON au.id = m.actor_id
		WHERE m.user_id = $1`+whereUnread+`
		ORDER BY m.id DESC
		LIMIT $2
	`, uid, limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Mention{}
	for rows.Next() {
		var m Mention
		if err := rows.Scan(&m.ID, &m.RunID, &m.CommentID, &m.ActorID, &m.ActorName,
			&m.Body, &m.TaskName, &m.ProjectName, &m.CreatedAt, &m.ReadAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, m)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// MentionsCount returns just unread/total counts for a header badge.
// GET /api/me/mentions/count
func (h *Handlers) MentionsCount(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var total, unread int
	if err := h.Pool.QueryRow(r.Context(), `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE read_at IS NULL)
		FROM mentions WHERE user_id = $1
	`, uid).Scan(&total, &unread); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{
		"total":  total,
		"unread": unread,
	})
}

// MarkRead marks one or all of a user's mentions as read. POST body
// {"id": <mention id>} marks one; empty body marks all.
// POST /api/me/mentions/read
func (h *Handlers) MarkRead(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var body struct {
		ID *int64 `json:"id"`
	}
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	var err error
	if body.ID != nil {
		_, err = h.Pool.Exec(r.Context(),
			`UPDATE mentions SET read_at = now() WHERE id = $1 AND user_id = $2 AND read_at IS NULL`,
			*body.ID, uid)
	} else {
		_, err = h.Pool.Exec(r.Context(),
			`UPDATE mentions SET read_at = now() WHERE user_id = $1 AND read_at IS NULL`,
			uid)
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) userOwnsRun(ctx context.Context, uid, runID uuid.UUID) bool {
	var n int
	err := h.Pool.QueryRow(ctx, `
		SELECT 1 FROM runs r
		JOIN projects p   ON p.id = r.project_id
		JOIN workspaces w ON w.id = p.workspace_id
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE r.id = $1 AND m.user_id = $2
	`, runID, uid).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	return err == nil
}

// extractMentions returns lowercased unique handles found in `@name` form.
// Handles trailing punctuation (commas, periods) by anchoring on the regex.
func extractMentions(body string) []string {
	matches := mentionRegex.FindAllStringSubmatch(body, -1)
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		name := strings.ToLower(m[1])
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}
