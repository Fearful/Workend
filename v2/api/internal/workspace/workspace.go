package workspace

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

type Workspace struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Handlers struct {
	Pool *pgxpool.Pool
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.Pool.Query(r.Context(), `
		SELECT id, name, description, created_at, updated_at
		FROM workspaces
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Workspace{}
	for rows.Next() {
		var ws Workspace
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.Description, &ws.CreatedAt, &ws.UpdatedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, ws)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type createReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())

	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	var ws Workspace
	err := h.Pool.QueryRow(r.Context(), `
		INSERT INTO workspaces (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, created_at, updated_at
	`, uid, req.Name, req.Description).Scan(&ws.ID, &ws.Name, &ws.Description, &ws.CreatedAt, &ws.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "workspaces_user_name_idx") {
			http.Error(w, "workspace name already in use", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ws)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var ws Workspace
	err = h.Pool.QueryRow(r.Context(), `
		SELECT id, name, description, created_at, updated_at
		FROM workspaces
		WHERE id = $1 AND user_id = $2
	`, id, uid).Scan(&ws.ID, &ws.Name, &ws.Description, &ws.CreatedAt, &ws.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ws)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	tag, err := h.Pool.Exec(r.Context(), `
		DELETE FROM workspaces WHERE id = $1 AND user_id = $2
	`, id, uid)
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
