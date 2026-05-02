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
	case KindWebhook, KindSlack, KindEmail:
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
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
