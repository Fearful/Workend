package rbac

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
	"workend/api/internal/workspace"
)

// Valid resource and action sets for permission validation.
var (
	validResources = map[string]bool{
		"workspace": true, "project": true, "task": true, "run": true,
		"secret": true, "sandbox": true, "pipeline": true, "member": true,
	}
	validActions = map[string]bool{
		"view": true, "create": true, "update": true, "delete": true,
		"run": true, "approve": true, "admin": true,
	}
	allResources = []string{
		"workspace", "project", "task", "run",
		"secret", "sandbox", "pipeline", "member",
	}
	allActions = []string{
		"view", "create", "update", "delete",
		"run", "approve", "admin",
	}
)

// CustomRole represents a workspace role with optional permissions.
type CustomRole struct {
	ID          uuid.UUID    `json:"id"`
	WorkspaceID uuid.UUID    `json:"workspace_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	IsSystem    bool         `json:"is_system"`
	CreatedAt   time.Time    `json:"created_at"`
	Permissions []Permission `json:"permissions,omitempty"`
}

// Permission represents a single resource+action grant on a role.
type Permission struct {
	ID       uuid.UUID `json:"id"`
	RoleID   uuid.UUID `json:"role_id"`
	Resource string    `json:"resource"`
	Action   string    `json:"action"`
}

// Handlers groups the HTTP handler methods for the RBAC feature.
type Handlers struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
}

// ---------- system-role definitions ----------

type systemRoleDef struct {
	name        string
	description string
	permissions []permDef
}

type permDef struct {
	resource string
	action   string
}

func ownerPerms() []permDef {
	out := make([]permDef, 0, len(allResources)*len(allActions))
	for _, r := range allResources {
		for _, a := range allActions {
			out = append(out, permDef{r, a})
		}
	}
	return out
}

func adminPerms() []permDef {
	out := make([]permDef, 0, len(allResources)*len(allActions))
	for _, r := range allResources {
		for _, a := range allActions {
			if r == "member" && a == "admin" {
				continue
			}
			out = append(out, permDef{r, a})
		}
	}
	return out
}

func developerPerms() []permDef {
	devResources := map[string][]string{
		"project":  {"view", "create", "run", "update"},
		"task":     {"view", "create", "run", "update"},
		"run":      {"view", "create", "run", "update"},
		"sandbox":  {"view", "create", "run", "update"},
		"pipeline": {"view", "create", "run", "update"},
		"secret":   {"view"},
		"member":   {"view"},
	}
	var out []permDef
	for res, actions := range devResources {
		for _, a := range actions {
			out = append(out, permDef{res, a})
		}
	}
	return out
}

func viewerPerms() []permDef {
	out := make([]permDef, 0, len(allResources))
	for _, r := range allResources {
		out = append(out, permDef{r, "view"})
	}
	return out
}

var systemRoles = []systemRoleDef{
	{name: "owner", description: "Full access to all resources", permissions: ownerPerms()},
	{name: "admin", description: "Full access except member admin", permissions: adminPerms()},
	{name: "developer", description: "Create and run projects, tasks, and pipelines", permissions: developerPerms()},
	{name: "viewer", description: "Read-only access to all resources", permissions: viewerPerms()},
}

// EnsureSystemRoles creates the four built-in system roles and their
// permissions for a workspace. Safe to call multiple times thanks to
// ON CONFLICT DO NOTHING.
func EnsureSystemRoles(ctx context.Context, pool *pgxpool.Pool, wsID uuid.UUID) error {
	for _, def := range systemRoles {
		var roleID uuid.UUID
		err := pool.QueryRow(ctx, `
			INSERT INTO custom_roles (workspace_id, name, description, is_system)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (workspace_id, name) DO UPDATE SET name = custom_roles.name
			RETURNING id
		`, wsID, def.name, def.description).Scan(&roleID)
		if err != nil {
			return err
		}

		for _, p := range def.permissions {
			if _, err := pool.Exec(ctx, `
				INSERT INTO role_permissions (role_id, resource, action)
				VALUES ($1, $2, $3)
				ON CONFLICT (role_id, resource, action) DO NOTHING
			`, roleID, p.resource, p.action); err != nil {
				return err
			}
		}
	}
	return nil
}

// ---------- HTTP handlers ----------

// ListRoles returns all roles (system + custom) with permissions hydrated.
// GET /api/workspaces/{id}/roles
func (h *Handlers) ListRoles(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	roles, err := h.listRolesWithPerms(r.Context(), wsID)
	if err != nil {
		h.Logger.Error("list roles", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(roles)
}

func (h *Handlers) listRolesWithPerms(ctx context.Context, wsID uuid.UUID) ([]CustomRole, error) {
	rows, err := h.Pool.Query(ctx, `
		SELECT id, workspace_id, name, description, is_system, created_at
		FROM custom_roles
		WHERE workspace_id = $1
		ORDER BY is_system DESC, name ASC
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []CustomRole
	roleIDs := []uuid.UUID{}
	roleIdx := map[uuid.UUID]int{}
	for rows.Next() {
		var cr CustomRole
		if err := rows.Scan(&cr.ID, &cr.WorkspaceID, &cr.Name, &cr.Description, &cr.IsSystem, &cr.CreatedAt); err != nil {
			return nil, err
		}
		roleIdx[cr.ID] = len(roles)
		roleIDs = append(roleIDs, cr.ID)
		cr.Permissions = []Permission{}
		roles = append(roles, cr)
	}
	if len(roleIDs) == 0 {
		return []CustomRole{}, nil
	}

	permRows, err := h.Pool.Query(ctx, `
		SELECT id, role_id, resource, action
		FROM role_permissions
		WHERE role_id = ANY($1)
		ORDER BY resource, action
	`, roleIDs)
	if err != nil {
		return nil, err
	}
	defer permRows.Close()

	for permRows.Next() {
		var p Permission
		if err := permRows.Scan(&p.ID, &p.RoleID, &p.Resource, &p.Action); err != nil {
			return nil, err
		}
		if idx, ok := roleIdx[p.RoleID]; ok {
			roles[idx].Permissions = append(roles[idx].Permissions, p)
		}
	}

	return roles, nil
}

// createRoleReq is the JSON body for CreateRole.
type createRoleReq struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []permInput  `json:"permissions"`
}

type permInput struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// CreateRole creates a custom (non-system) role with permissions.
// POST /api/workspaces/{id}/roles
func (h *Handlers) CreateRole(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	role, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil || role != workspace.RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req createRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 50 {
		http.Error(w, "name must be 1-50 characters", http.StatusBadRequest)
		return
	}
	if len(req.Permissions) > 50 {
		http.Error(w, "max 50 permissions per role", http.StatusBadRequest)
		return
	}
	for _, p := range req.Permissions {
		if !validResources[p.Resource] {
			http.Error(w, "invalid resource: "+p.Resource, http.StatusBadRequest)
			return
		}
		if !validActions[p.Action] {
			http.Error(w, "invalid action: "+p.Action, http.StatusBadRequest)
			return
		}
	}

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	var cr CustomRole
	err = tx.QueryRow(r.Context(), `
		INSERT INTO custom_roles (workspace_id, name, description, is_system)
		VALUES ($1, $2, $3, false)
		RETURNING id, workspace_id, name, description, is_system, created_at
	`, wsID, req.Name, req.Description).Scan(
		&cr.ID, &cr.WorkspaceID, &cr.Name, &cr.Description, &cr.IsSystem, &cr.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "custom_roles_workspace_id_name_key") {
			http.Error(w, "role name already exists", http.StatusConflict)
			return
		}
		h.Logger.Error("create role", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	cr.Permissions = make([]Permission, 0, len(req.Permissions))
	for _, p := range req.Permissions {
		var perm Permission
		err := tx.QueryRow(r.Context(), `
			INSERT INTO role_permissions (role_id, resource, action)
			VALUES ($1, $2, $3)
			RETURNING id, role_id, resource, action
		`, cr.ID, p.Resource, p.Action).Scan(&perm.ID, &perm.RoleID, &perm.Resource, &perm.Action)
		if err != nil {
			h.Logger.Error("insert permission", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		cr.Permissions = append(cr.Permissions, perm)
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(cr)
}

// updateRoleReq is the JSON body for UpdateRole.
type updateRoleReq struct {
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Permissions *[]permInput `json:"permissions"`
}

// UpdateRole patches a custom role. System roles cannot be modified.
// PATCH /api/roles/{id}
func (h *Handlers) UpdateRole(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid role id", http.StatusBadRequest)
		return
	}

	var existing CustomRole
	err = h.Pool.QueryRow(r.Context(), `
		SELECT id, workspace_id, name, description, is_system, created_at
		FROM custom_roles WHERE id = $1
	`, roleID).Scan(&existing.ID, &existing.WorkspaceID, &existing.Name,
		&existing.Description, &existing.IsSystem, &existing.CreatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if existing.IsSystem {
		http.Error(w, "cannot modify system roles", http.StatusBadRequest)
		return
	}

	role, err := workspace.RoleOf(r.Context(), h.Pool, uid, existing.WorkspaceID)
	if err != nil || role != workspace.RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req updateRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" || len(trimmed) > 50 {
			http.Error(w, "name must be 1-50 characters", http.StatusBadRequest)
			return
		}
		req.Name = &trimmed
	}
	if req.Permissions != nil {
		if len(*req.Permissions) > 50 {
			http.Error(w, "max 50 permissions per role", http.StatusBadRequest)
			return
		}
		for _, p := range *req.Permissions {
			if !validResources[p.Resource] {
				http.Error(w, "invalid resource: "+p.Resource, http.StatusBadRequest)
				return
			}
			if !validActions[p.Action] {
				http.Error(w, "invalid action: "+p.Action, http.StatusBadRequest)
				return
			}
		}
	}

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	_, err = tx.Exec(r.Context(), `
		UPDATE custom_roles
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description)
		WHERE id = $3
	`, req.Name, req.Description, roleID)
	if err != nil {
		if strings.Contains(err.Error(), "custom_roles_workspace_id_name_key") {
			http.Error(w, "role name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if req.Permissions != nil {
		if _, err := tx.Exec(r.Context(), `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		for _, p := range *req.Permissions {
			if _, err := tx.Exec(r.Context(), `
				INSERT INTO role_permissions (role_id, resource, action)
				VALUES ($1, $2, $3)
			`, roleID, p.Resource, p.Action); err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	updated, err := h.fetchRoleWithPerms(r.Context(), roleID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func (h *Handlers) fetchRoleWithPerms(ctx context.Context, roleID uuid.UUID) (*CustomRole, error) {
	var cr CustomRole
	err := h.Pool.QueryRow(ctx, `
		SELECT id, workspace_id, name, description, is_system, created_at
		FROM custom_roles WHERE id = $1
	`, roleID).Scan(&cr.ID, &cr.WorkspaceID, &cr.Name, &cr.Description, &cr.IsSystem, &cr.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := h.Pool.Query(ctx, `
		SELECT id, role_id, resource, action
		FROM role_permissions
		WHERE role_id = $1
		ORDER BY resource, action
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cr.Permissions = []Permission{}
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.RoleID, &p.Resource, &p.Action); err != nil {
			return nil, err
		}
		cr.Permissions = append(cr.Permissions, p)
	}
	return &cr, nil
}

// DeleteRole deletes a custom role. System roles cannot be deleted.
// DELETE /api/roles/{id}
func (h *Handlers) DeleteRole(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid role id", http.StatusBadRequest)
		return
	}

	var wsID uuid.UUID
	var isSystem bool
	err = h.Pool.QueryRow(r.Context(), `
		SELECT workspace_id, is_system FROM custom_roles WHERE id = $1
	`, roleID).Scan(&wsID, &isSystem)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if isSystem {
		http.Error(w, "cannot delete system roles", http.StatusBadRequest)
		return
	}

	role, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil || role != workspace.RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `DELETE FROM custom_roles WHERE id = $1`, roleID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// assignRoleReq is the JSON body for AssignRole.
type assignRoleReq struct {
	RoleID uuid.UUID `json:"role_id"`
}

// AssignRole assigns a custom role to a workspace member.
// POST /api/workspaces/{id}/members/{user_id}/role
func (h *Handlers) AssignRole(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	targetUID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	role, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID)
	if err != nil || role != workspace.RoleOwner {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Verify target user is a workspace member.
	if _, err := workspace.RoleOf(r.Context(), h.Pool, targetUID, wsID); err != nil {
		http.Error(w, "user is not a workspace member", http.StatusBadRequest)
		return
	}

	var req assignRoleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validate the role belongs to this workspace.
	var roleWSID uuid.UUID
	err = h.Pool.QueryRow(r.Context(), `
		SELECT workspace_id FROM custom_roles WHERE id = $1
	`, req.RoleID).Scan(&roleWSID)
	if err != nil {
		http.Error(w, "role not found", http.StatusBadRequest)
		return
	}
	if roleWSID != wsID {
		http.Error(w, "role does not belong to this workspace", http.StatusBadRequest)
		return
	}

	if _, err := h.Pool.Exec(r.Context(), `
		INSERT INTO member_role_assignments (workspace_id, user_id, role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET role_id = EXCLUDED.role_id, assigned_at = now()
	`, wsID, targetUID, req.RoleID); err != nil {
		h.Logger.Error("assign role", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMemberRole returns the assigned custom role (with permissions) for a
// workspace member, falling back to the system role implied by the legacy
// workspace_members.role column.
// GET /api/workspaces/{id}/members/{user_id}/role
func (h *Handlers) GetMemberRole(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid workspace id", http.StatusBadRequest)
		return
	}
	targetUID, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if _, err := workspace.RoleOf(r.Context(), h.Pool, uid, wsID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	cr, err := h.resolveEffectiveRole(r.Context(), wsID, targetUID)
	if err != nil {
		h.Logger.Error("get member role", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if cr == nil {
		http.Error(w, "user is not a workspace member", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cr)
}

// resolveEffectiveRole returns the explicitly assigned custom role or falls
// back to the legacy workspace_members.role mapping (owner->"owner",
// member->"developer").
func (h *Handlers) resolveEffectiveRole(ctx context.Context, wsID, userID uuid.UUID) (*CustomRole, error) {
	// Try explicit assignment first.
	var assignedRoleID uuid.UUID
	err := h.Pool.QueryRow(ctx, `
		SELECT role_id FROM member_role_assignments
		WHERE workspace_id = $1 AND user_id = $2
	`, wsID, userID).Scan(&assignedRoleID)
	if err == nil {
		return h.fetchRoleWithPerms(ctx, assignedRoleID)
	}

	// Fall back to legacy workspace_members.role.
	legacyRole, err := workspace.RoleOf(ctx, h.Pool, userID, wsID)
	if err != nil {
		return nil, nil
	}
	systemName := "developer"
	if legacyRole == workspace.RoleOwner {
		systemName = "owner"
	}

	var cr CustomRole
	err = h.Pool.QueryRow(ctx, `
		SELECT id, workspace_id, name, description, is_system, created_at
		FROM custom_roles
		WHERE workspace_id = $1 AND name = $2 AND is_system = true
	`, wsID, systemName).Scan(&cr.ID, &cr.WorkspaceID, &cr.Name, &cr.Description, &cr.IsSystem, &cr.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := h.Pool.Query(ctx, `
		SELECT id, role_id, resource, action
		FROM role_permissions
		WHERE role_id = $1
		ORDER BY resource, action
	`, cr.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cr.Permissions = []Permission{}
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.RoleID, &p.Resource, &p.Action); err != nil {
			return nil, err
		}
		cr.Permissions = append(cr.Permissions, p)
	}
	return &cr, nil
}

// CheckPermission is a utility function (not an HTTP handler) that checks
// whether a user has a specific permission on a workspace. It resolves the
// user's effective role and scans its permissions.
func CheckPermission(ctx context.Context, pool *pgxpool.Pool, uid, wsID uuid.UUID, resource, action string) bool {
	// Try explicit assignment first.
	var assignedRoleID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT role_id FROM member_role_assignments
		WHERE workspace_id = $1 AND user_id = $2
	`, wsID, uid).Scan(&assignedRoleID)

	var roleID uuid.UUID
	if err == nil {
		roleID = assignedRoleID
	} else {
		// Fall back to legacy role mapping.
		legacyRole, lerr := workspace.RoleOf(ctx, pool, uid, wsID)
		if lerr != nil {
			return false
		}
		systemName := "developer"
		if legacyRole == workspace.RoleOwner {
			systemName = "owner"
		}
		err = pool.QueryRow(ctx, `
			SELECT id FROM custom_roles
			WHERE workspace_id = $1 AND name = $2 AND is_system = true
		`, wsID, systemName).Scan(&roleID)
		if err != nil {
			return false
		}
	}

	var found bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM role_permissions
			WHERE role_id = $1 AND resource = $2 AND action = $3
		)
	`, roleID, resource, action).Scan(&found)
	if err != nil {
		return false
	}
	return found
}
