package notify

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"workend/api/internal/auth"
)

type Handlers struct {
	D *Dispatcher
}

// GET /api/me/notifications
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.D.Pool.Query(r.Context(), `
		SELECT id, user_id, kind, target, trigger, enabled, created_at
		FROM notification_configs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Config{}
	for rows.Next() {
		var c Config
		if err := rows.Scan(&c.ID, &c.UserID, &c.Kind, &c.Target, &c.Trigger, &c.Enabled, &c.CreatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, c)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type createReq struct {
	Kind    string `json:"kind"`
	Target  string `json:"target"`
	Trigger string `json:"trigger"`
}

// POST /api/me/notifications
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Kind = strings.TrimSpace(req.Kind)
	req.Target = strings.TrimSpace(req.Target)
	req.Trigger = strings.TrimSpace(req.Trigger)
	if req.Kind == "" || req.Target == "" {
		http.Error(w, "kind and target required", http.StatusBadRequest)
		return
	}
	if req.Trigger == "" {
		req.Trigger = TriggerOnFailure
	}
	switch req.Kind {
	case KindWebhook, KindSlack, KindEmail, KindDiscord, KindTeams:
	default:
		http.Error(w, "invalid kind", http.StatusBadRequest)
		return
	}

	var c Config
	err := h.D.Pool.QueryRow(r.Context(), `
		INSERT INTO notification_configs (user_id, kind, target, trigger)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, kind, target, trigger, enabled, created_at
	`, uid, req.Kind, req.Target, req.Trigger).Scan(
		&c.ID, &c.UserID, &c.Kind, &c.Target, &c.Trigger, &c.Enabled, &c.CreatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// Auto-create a sensible default subscription so a brand-new config
	// actually fires for something. Users can adjust scopes later.
	_, _ = h.D.Pool.Exec(r.Context(), `
		INSERT INTO notification_subscriptions (user_id, config_id, scope_type, severity)
		VALUES ($1, $2, 'global', 'failures')
	`, uid, c.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

// ----- Subscriptions -----

type Subscription struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	ConfigID  uuid.UUID  `json:"config_id"`
	ScopeType string     `json:"scope_type"`
	ScopeID   *uuid.UUID `json:"scope_id"`
	Severity  string     `json:"severity"`
	CreatedAt string     `json:"created_at"`

	// Hydrated for the UI:
	ScopeName  string `json:"scope_name,omitempty"`
	ConfigKind string `json:"config_kind,omitempty"`
}

// GET /api/me/notification-subscriptions
func (h *Handlers) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.D.Pool.Query(r.Context(), `
		SELECT s.id, s.user_id, s.config_id, s.scope_type, s.scope_id, s.severity, s.created_at,
		       c.kind,
		       CASE s.scope_type
		           WHEN 'workspace' THEN (SELECT name FROM workspaces WHERE id = s.scope_id)
		           WHEN 'project'   THEN (SELECT name FROM projects   WHERE id = s.scope_id)
		           WHEN 'task'      THEN (SELECT name FROM tasks      WHERE id = s.scope_id)
		           ELSE 'all runs'
		       END AS scope_name
		FROM notification_subscriptions s
		JOIN notification_configs c ON c.id = s.config_id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Subscription{}
	for rows.Next() {
		var s Subscription
		var scopeName *string
		if err := rows.Scan(&s.ID, &s.UserID, &s.ConfigID, &s.ScopeType, &s.ScopeID, &s.Severity, &s.CreatedAt, &s.ConfigKind, &scopeName); err != nil {
			continue
		}
		if scopeName != nil {
			s.ScopeName = *scopeName
		}
		out = append(out, s)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// POST /api/me/notification-subscriptions
//   {"config_id":"…","scope_type":"project","scope_id":"…","severity":"failures"}
func (h *Handlers) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var req struct {
		ConfigID  string `json:"config_id"`
		ScopeType string `json:"scope_type"`
		ScopeID   string `json:"scope_id"`
		Severity  string `json:"severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	configID, err := uuid.Parse(strings.TrimSpace(req.ConfigID))
	if err != nil {
		http.Error(w, "invalid config_id", http.StatusBadRequest)
		return
	}
	switch req.ScopeType {
	case "global", "workspace", "project", "task":
	default:
		http.Error(w, "scope_type must be global|workspace|project|task", http.StatusBadRequest)
		return
	}
	switch req.Severity {
	case "all", "failures", "off":
	default:
		http.Error(w, "severity must be all|failures|off", http.StatusBadRequest)
		return
	}
	var scopeID *uuid.UUID
	if req.ScopeType != "global" {
		id, err := uuid.Parse(strings.TrimSpace(req.ScopeID))
		if err != nil {
			http.Error(w, "scope_id required for non-global scope", http.StatusBadRequest)
			return
		}
		scopeID = &id
	}

	// Confirm the config belongs to the user.
	var owns bool
	_ = h.D.Pool.QueryRow(r.Context(),
		`SELECT EXISTS (SELECT 1 FROM notification_configs WHERE id = $1 AND user_id = $2)`,
		configID, uid).Scan(&owns)
	if !owns {
		http.Error(w, "config not found", http.StatusNotFound)
		return
	}

	var s Subscription
	err = h.D.Pool.QueryRow(r.Context(), `
		INSERT INTO notification_subscriptions (user_id, config_id, scope_type, scope_id, severity)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, config_id, scope_type, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'::uuid))
		DO UPDATE SET severity = EXCLUDED.severity
		RETURNING id, user_id, config_id, scope_type, scope_id, severity, created_at
	`, uid, configID, req.ScopeType, scopeID, req.Severity).Scan(
		&s.ID, &s.UserID, &s.ConfigID, &s.ScopeType, &s.ScopeID, &s.Severity, &s.CreatedAt)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(s)
}

// DELETE /api/me/notification-subscriptions/:id
func (h *Handlers) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	tag, err := h.D.Pool.Exec(r.Context(),
		`DELETE FROM notification_subscriptions WHERE id = $1 AND user_id = $2`, id, uid)
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

// DELETE /api/me/notifications/:id
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	tag, err := h.D.Pool.Exec(r.Context(),
		`DELETE FROM notification_configs WHERE id = $1 AND user_id = $2`, id, uid)
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

// POST /api/me/notifications/:id/test  (synchronous fire to verify config)
func (h *Handlers) Test(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var c Config
	err = h.D.Pool.QueryRow(r.Context(), `
		SELECT id, user_id, kind, target, trigger, enabled, created_at
		FROM notification_configs
		WHERE id = $1 AND user_id = $2
	`, id, uid).Scan(&c.ID, &c.UserID, &c.Kind, &c.Target, &c.Trigger, &c.Enabled, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	e := Event{
		RunID:       uuid.New(),
		UserID:      uid,
		ProjectName: "test-project",
		TaskName:    "test-task",
		TaskSource:  "test",
		Status:      "succeeded",
		ExitCode:    0,
		DurationSec: 0,
		URL:         h.D.WebURL + "/",
	}
	if err := h.D.send(r.Context(), c, e); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
