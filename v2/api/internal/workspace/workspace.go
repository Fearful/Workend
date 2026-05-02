package workspace

import (
	"context"
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

const (
	RoleOwner  = "owner"
	RoleMember = "member"
)

type Workspace struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// MyRole is hydrated from workspace_members for the requesting user.
	MyRole string `json:"my_role,omitempty"`
}

type Member struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	AddedAt     time.Time `json:"added_at"`
}

type Handlers struct {
	Pool *pgxpool.Pool
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := h.Pool.Query(r.Context(), `
		SELECT w.id, w.name, w.description, w.created_at, w.updated_at, m.role
		FROM workspaces w
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE m.user_id = $1
		ORDER BY w.created_at DESC
	`, uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []Workspace{}
	for rows.Next() {
		var ws Workspace
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.Description, &ws.CreatedAt, &ws.UpdatedAt, &ws.MyRole); err != nil {
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

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var ws Workspace
	err = tx.QueryRow(r.Context(), `
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
	// Membership row for the creator.
	if _, err := tx.Exec(r.Context(), `
		INSERT INTO workspace_members (workspace_id, user_id, role, added_by)
		VALUES ($1, $2, $3, $2)
	`, ws.ID, uid, RoleOwner); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	ws.MyRole = RoleOwner

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
		SELECT w.id, w.name, w.description, w.created_at, w.updated_at, m.role
		FROM workspaces w
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE w.id = $1 AND m.user_id = $2
	`, id, uid).Scan(&ws.ID, &ws.Name, &ws.Description, &ws.CreatedAt, &ws.UpdatedAt, &ws.MyRole)
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

	// Only owners can delete.
	role, err := RoleOf(r.Context(), h.Pool, uid, id)
	if err != nil || role != RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `DELETE FROM workspaces WHERE id = $1`, id); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- members ---

// GET /api/workspaces/:id/members
func (h *Handlers) ListMembers(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := RoleOf(r.Context(), h.Pool, uid, id); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rows, err := h.Pool.Query(r.Context(), `
		SELECT u.id, u.email, u.display_name, m.role, m.added_at
		FROM workspace_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = $1
		ORDER BY m.role = 'owner' DESC, m.added_at ASC
	`, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID, &m.Email, &m.DisplayName, &m.Role, &m.AddedAt); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		out = append(out, m)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type addMemberReq struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// POST /api/workspaces/:id/members
func (h *Handlers) AddMember(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	role, err := RoleOf(r.Context(), h.Pool, uid, id)
	if err != nil || role != RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req addMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		http.Error(w, "email required", http.StatusBadRequest)
		return
	}
	switch req.Role {
	case "", RoleMember:
		req.Role = RoleMember
	case RoleOwner:
		// allowed
	default:
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	var addedID uuid.UUID
	if err := h.Pool.QueryRow(r.Context(),
		`SELECT id FROM users WHERE lower(email) = $1`, req.Email).Scan(&addedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "no user with that email", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `
		INSERT INTO workspace_members (workspace_id, user_id, role, added_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, id, addedID, req.Role, uid); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/workspaces/:id/members/:user_id
func (h *Handlers) RemoveMember(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	memberID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	role, err := RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	// Members can remove themselves; owners can remove anyone except the
	// last owner.
	if memberID != uid && role != RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	// Block removing the last owner.
	if memberID == uid && role == RoleOwner {
		var owners int
		_ = h.Pool.QueryRow(r.Context(),
			`SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1 AND role = 'owner'`, wsID).Scan(&owners)
		if owners <= 1 {
			http.Error(w, "cannot remove the last owner; transfer ownership first", http.StatusConflict)
			return
		}
	}
	if _, err := h.Pool.Exec(r.Context(),
		`DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, wsID, memberID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RoleOf returns the user's role on a workspace, or an error if no membership.
func RoleOf(ctx context.Context, pool *pgxpool.Pool, userID, workspaceID uuid.UUID) (string, error) {
	var role string
	err := pool.QueryRow(ctx, `
		SELECT role FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID).Scan(&role)
	return role, err
}
