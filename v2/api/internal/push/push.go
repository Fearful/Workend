// Package push delivers Web Push notifications to subscribed browser
// devices. Pairs with the service worker registered by the SvelteKit app.
//
// VAPID keys live in env (WORKEND_VAPID_PUBLIC, WORKEND_VAPID_PRIVATE,
// WORKEND_VAPID_SUBJECT). When unset, /me/push/vapid returns 503 and the
// frontend falls back to its existing per-tab Notification API.
package push

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"

	"workend/api/internal/auth"
)

type Config struct {
	PublicKey  string
	PrivateKey string
	Subject    string // mailto:admin@example.com
}

type Sender struct {
	Pool   *pgxpool.Pool
	Cfg    Config
	Logger *slog.Logger
}

func New(pool *pgxpool.Pool, cfg Config, logger *slog.Logger) *Sender {
	return &Sender{Pool: pool, Cfg: cfg, Logger: logger}
}

func (s *Sender) Enabled() bool {
	return s != nil && s.Cfg.PublicKey != "" && s.Cfg.PrivateKey != ""
}

// Send fans out to every active subscription for `userID`.
// Best-effort: failed subscriptions (410 Gone) are pruned.
func (s *Sender) Send(userID uuid.UUID, payload map[string]any) {
	if s == nil || !s.Enabled() {
		return
	}
	go s.sendNow(userID, payload)
}

func (s *Sender) sendNow(userID uuid.UUID, payload map[string]any) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rows, err := s.Pool.Query(ctx, `
		SELECT id, endpoint, auth, p256dh
		FROM push_subscriptions WHERE user_id = $1
	`, userID)
	if err != nil {
		if s.Logger != nil {
			s.Logger.Warn("push: load subs failed", "err", err)
		}
		return
	}
	defer rows.Close()
	type row struct {
		ID                 uuid.UUID
		Endpoint, Auth, P  string
	}
	var subs []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.Endpoint, &r.Auth, &r.P); err != nil {
			continue
		}
		subs = append(subs, r)
	}

	body, _ := json.Marshal(payload)
	for _, sub := range subs {
		ws := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys:     webpush.Keys{Auth: sub.Auth, P256dh: sub.P},
		}
		resp, err := webpush.SendNotification(body, ws, &webpush.Options{
			Subscriber:      s.Cfg.Subject,
			VAPIDPublicKey:  s.Cfg.PublicKey,
			VAPIDPrivateKey: s.Cfg.PrivateKey,
			TTL:             60,
		})
		if err != nil {
			if s.Logger != nil {
				s.Logger.Warn("push: send failed", "sub", sub.ID, "err", err)
			}
			continue
		}
		// 410 Gone or 404 mean the subscription is dead — drop it so we don't
		// keep retrying on the next event.
		if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
			_, _ = s.Pool.Exec(ctx, `DELETE FROM push_subscriptions WHERE id = $1`, sub.ID)
		}
		_ = resp.Body.Close()
	}
}

// --- HTTP handlers ---

type Handlers struct {
	Pool *pgxpool.Pool
	Cfg  Config
}

// VAPID returns the public key the browser needs to subscribe.
// GET /api/me/push/vapid
func (h *Handlers) VAPID(w http.ResponseWriter, r *http.Request) {
	if h.Cfg.PublicKey == "" {
		http.Error(w, "web push not configured", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"public_key": h.Cfg.PublicKey})
}

// Subscribe stores a new subscription. Idempotent on (endpoint).
// POST /api/me/push/subscribe
//   body: {"endpoint":"...","keys":{"auth":"...","p256dh":"..."}}
func (h *Handlers) Subscribe(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var body struct {
		Endpoint string `json:"endpoint"`
		Keys     struct {
			Auth   string `json:"auth"`
			P256dh string `json:"p256dh"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	body.Endpoint = strings.TrimSpace(body.Endpoint)
	if body.Endpoint == "" || body.Keys.Auth == "" || body.Keys.P256dh == "" {
		http.Error(w, "endpoint and keys required", http.StatusBadRequest)
		return
	}
	ua := r.Header.Get("User-Agent")
	_, err := h.Pool.Exec(r.Context(), `
		INSERT INTO push_subscriptions (user_id, endpoint, auth, p256dh, user_agent)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''))
		ON CONFLICT (endpoint) DO UPDATE
		SET user_id = EXCLUDED.user_id, auth = EXCLUDED.auth, p256dh = EXCLUDED.p256dh,
		    user_agent = EXCLUDED.user_agent
	`, uid, body.Endpoint, body.Keys.Auth, body.Keys.P256dh, ua)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unsubscribe drops a subscription by its endpoint.
// DELETE /api/me/push/subscribe?endpoint=<url-encoded>
func (h *Handlers) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		http.Error(w, "endpoint required", http.StatusBadRequest)
		return
	}
	_, _ = h.Pool.Exec(r.Context(),
		`DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`, uid, endpoint)
	w.WriteHeader(http.StatusNoContent)
}

// List returns the user's subscriptions (for the settings UI).
// GET /api/me/push/subscriptions
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, endpoint, COALESCE(user_agent, ''), created_at
		FROM push_subscriptions WHERE user_id = $1 ORDER BY created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type sub struct {
		ID        uuid.UUID `json:"id"`
		Endpoint  string    `json:"endpoint"`
		UserAgent string    `json:"user_agent"`
		CreatedAt time.Time `json:"created_at"`
	}
	out := []sub{}
	for rows.Next() {
		var s sub
		if err := rows.Scan(&s.ID, &s.Endpoint, &s.UserAgent, &s.CreatedAt); err != nil {
			continue
		}
		out = append(out, s)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// DeleteByID is the settings-page "remove this device" affordance.
// DELETE /api/me/push/subscriptions/:id
func (h *Handlers) DeleteByID(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	tag, err := h.Pool.Exec(r.Context(),
		`DELETE FROM push_subscriptions WHERE id = $1 AND user_id = $2`, id, uid)
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

// silence linter
var _ = errors.New
