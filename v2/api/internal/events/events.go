// Package events dispatches structured events to user-configured outbound
// webhooks. Distinct from package notify (which is "tell *me* when a run
// finishes"); this one is "let *my automation* react to anything".
//
// Payloads are HMAC-SHA256 signed with the per-target secret. The signature
// rides in the X-Workend-Signature-256 header as `sha256=<hex>`, mirroring
// GitHub's webhook scheme so existing tooling drops in.
package events

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"

	"workend/api/internal/auth"
)

const (
	EventRunCreated    = "run.created"
	EventRunStarted    = "run.started"
	EventRunFinished   = "run.finished"
	EventProjectSynced = "project.synced"
	EventImageBuilt    = "image.built"
)

// AllEvents is the set the UI offers as subscription checkboxes.
var AllEvents = []string{
	EventRunCreated, EventRunStarted, EventRunFinished,
	EventProjectSynced, EventImageBuilt,
}

type Target struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Name         string     `json:"name"`
	URL          string     `json:"url"`
	Secret       string     `json:"secret,omitempty"`
	Events       []string   `json:"events"`
	Enabled      bool       `json:"enabled"`
	CreatedAt    time.Time  `json:"created_at"`
	LastFiredAt  *time.Time `json:"last_fired_at"`
	LastStatus   *int       `json:"last_status"`
}

// Dispatcher fans out an event to every matching target for the user that
// owns the resource. Best-effort: failures are recorded on the target row
// for visibility but don't propagate.
type Dispatcher struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
	HTTP   *http.Client
}

func New(pool *pgxpool.Pool, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		Pool:   pool,
		Logger: logger,
		HTTP:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Emit enqueues an event for fan-out to every target owned by `userID`
// that subscribes to `event` (either explicitly or via "*").
func (d *Dispatcher) Emit(userID uuid.UUID, event string, payload map[string]any) {
	if d == nil {
		return
	}
	go d.dispatch(userID, event, payload)
}

func (d *Dispatcher) dispatch(userID uuid.UUID, event string, payload map[string]any) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := d.Pool.Query(ctx, `
		SELECT id, name, url, secret, events FROM outbound_webhook_targets
		WHERE user_id = $1 AND enabled = true
	`, userID)
	if err != nil {
		d.Logger.Warn("events: load targets failed", "err", err)
		return
	}
	defer rows.Close()

	type row struct {
		ID     uuid.UUID
		Name   string
		URL    string
		Secret string
		Events []string
	}
	var targets []row
	for rows.Next() {
		var t row
		var rawEvents []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.URL, &t.Secret, &rawEvents); err != nil {
			continue
		}
		_ = json.Unmarshal(rawEvents, &t.Events)
		targets = append(targets, t)
	}

	body := map[string]any{
		"event":     event,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      payload,
	}
	buf, _ := json.Marshal(body)

	for _, t := range targets {
		if !subscribes(t.Events, event) {
			continue
		}
		status, err := d.send(ctx, t.URL, t.Secret, buf)
		if err != nil {
			d.Logger.Warn("events: send failed", "target", t.ID, "url", t.URL, "err", err)
		}
		_, _ = d.Pool.Exec(ctx, `
			UPDATE outbound_webhook_targets
			SET last_fired_at = now(), last_status = $1
			WHERE id = $2
		`, status, t.ID)
	}
}

func (d *Dispatcher) send(ctx context.Context, url, secret string, body []byte) (int, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Workend-Signature-256", sig)
	req.Header.Set("User-Agent", "Workend-Webhook/1")
	resp, err := d.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("status %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func subscribes(subscribed []string, event string) bool {
	for _, s := range subscribed {
		if s == "*" || s == event {
			return true
		}
	}
	return false
}

// --- HTTP handlers ---

type Handlers struct {
	Pool *pgxpool.Pool
	D    *Dispatcher
}

// List returns all of a user's targets. Secrets are masked.
// GET /api/me/outbound-webhooks
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, user_id, name, url, events, enabled, created_at, last_fired_at, last_status
		FROM outbound_webhook_targets WHERE user_id = $1 ORDER BY created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Target{}
	for rows.Next() {
		var t Target
		var rawEvents []byte
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.URL, &rawEvents,
			&t.Enabled, &t.CreatedAt, &t.LastFiredAt, &t.LastStatus); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_ = json.Unmarshal(rawEvents, &t.Events)
		out = append(out, t)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Create registers a new outbound webhook target. The server generates the
// signing secret; it's returned exactly once in the response and never
// echoed by List.
//
// POST /api/me/outbound-webhooks  body: {"name":"...", "url":"...", "events":["*"]}
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var req struct {
		Name   string   `json:"name"`
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.URL = strings.TrimSpace(req.URL)
	if req.Name == "" || !strings.HasPrefix(req.URL, "http") {
		http.Error(w, "name and http(s) url required", http.StatusBadRequest)
		return
	}
	if len(req.Events) == 0 {
		req.Events = []string{"*"}
	}
	for _, e := range req.Events {
		if e == "*" {
			continue
		}
		ok := false
		for _, valid := range AllEvents {
			if e == valid {
				ok = true
				break
			}
		}
		if !ok {
			http.Error(w, "unknown event: "+e, http.StatusBadRequest)
			return
		}
	}

	// Generate a 32-byte URL-safe secret. Stored plaintext intentionally —
	// it's a shared secret for HMAC, equivalent to a password the user picks
	// for their webhook receiver. Encrypting it would just push the problem.
	var raw [32]byte
	_, _ = rand.Read(raw[:])
	secret := base64.RawURLEncoding.EncodeToString(raw[:])

	eventsJSON, _ := json.Marshal(req.Events)
	var t Target
	err := h.Pool.QueryRow(r.Context(), `
		INSERT INTO outbound_webhook_targets (user_id, name, url, secret, events)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		RETURNING id, user_id, name, url, secret, events, enabled, created_at
	`, uid, req.Name, req.URL, secret, string(eventsJSON)).Scan(
		&t.ID, &t.UserID, &t.Name, &t.URL, &t.Secret, &eventsJSON,
		&t.Enabled, &t.CreatedAt)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = json.Unmarshal(eventsJSON, &t.Events)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

// Delete removes a target.
// DELETE /api/me/outbound-webhooks/:id
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	tag, err := h.Pool.Exec(r.Context(),
		`DELETE FROM outbound_webhook_targets WHERE id = $1 AND user_id = $2`, id, uid)
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

// Test fires a synthetic event to one target.
// POST /api/me/outbound-webhooks/:id/test
func (h *Handlers) Test(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var url, secret string
	err = h.Pool.QueryRow(r.Context(), `
		SELECT url, secret FROM outbound_webhook_targets WHERE id = $1 AND user_id = $2
	`, id, uid).Scan(&url, &secret)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	body := map[string]any{
		"event":     "test",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      map[string]any{"hello": "from workend"},
	}
	buf, _ := json.Marshal(body)
	status, err := h.D.send(r.Context(), url, secret, buf)
	if err != nil {
		http.Error(w, fmt.Sprintf("delivery failed: HTTP %d %s", status, err), http.StatusBadGateway)
		return
	}
	_, _ = h.Pool.Exec(r.Context(), `
		UPDATE outbound_webhook_targets SET last_fired_at = now(), last_status = $1 WHERE id = $2
	`, status, id)
	w.WriteHeader(http.StatusNoContent)
}
