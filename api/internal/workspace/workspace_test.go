package workspace_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workend/api/internal/testutil"
)

func TestCreateWorkspace(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)
	uid := testutil.CreateUser(t, pool, "ws@example.com", "securepass123", "WS User")
	cookie := testutil.AuthCookie(t, pool, uid)

	resp := testutil.DoJSON(t, router, "POST", "/api/workspaces", map[string]string{
		"name": "my-workspace", "description": "test workspace",
	}, cookie)

	assert.Equal(t, http.StatusCreated, resp.Code)

	body := testutil.DecodeJSON[map[string]any](t, resp)
	assert.Equal(t, "my-workspace", body["name"])
	assert.Equal(t, "test workspace", body["description"])
	assert.Equal(t, "owner", body["my_role"])
	assert.NotEmpty(t, body["id"])
}

func TestCreateWorkspaceDuplicateName(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)
	uid := testutil.CreateUser(t, pool, "dup-ws@example.com", "securepass123", "Dup")
	cookie := testutil.AuthCookie(t, pool, uid)

	body := map[string]string{"name": "duplicate"}
	resp1 := testutil.DoJSON(t, router, "POST", "/api/workspaces", body, cookie)
	assert.Equal(t, http.StatusCreated, resp1.Code)

	resp2 := testutil.DoJSON(t, router, "POST", "/api/workspaces", body, cookie)
	assert.Equal(t, http.StatusConflict, resp2.Code)
}

func TestCreateWorkspaceEmptyName(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)
	uid := testutil.CreateUser(t, pool, "empty-ws@example.com", "securepass123", "E")
	cookie := testutil.AuthCookie(t, pool, uid)

	resp := testutil.DoJSON(t, router, "POST", "/api/workspaces", map[string]string{"name": ""}, cookie)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestListWorkspaces(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid1 := testutil.CreateUser(t, pool, "list1@example.com", "securepass123", "U1")
	uid2 := testutil.CreateUser(t, pool, "list2@example.com", "securepass123", "U2")
	cookie1 := testutil.AuthCookie(t, pool, uid1)
	cookie2 := testutil.AuthCookie(t, pool, uid2)

	testutil.CreateWorkspace(t, pool, uid1, "ws-a")
	testutil.CreateWorkspace(t, pool, uid1, "ws-b")
	testutil.CreateWorkspace(t, pool, uid2, "ws-c")

	resp := testutil.DoJSON(t, router, "GET", "/api/workspaces", nil, cookie1)
	assert.Equal(t, http.StatusOK, resp.Code)
	list1 := testutil.DecodeJSON[[]map[string]any](t, resp)
	assert.Len(t, list1, 2)

	resp2 := testutil.DoJSON(t, router, "GET", "/api/workspaces", nil, cookie2)
	assert.Equal(t, http.StatusOK, resp2.Code)
	list2 := testutil.DecodeJSON[[]map[string]any](t, resp2)
	assert.Len(t, list2, 1)
}

func TestGetWorkspace(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid := testutil.CreateUser(t, pool, "get-ws@example.com", "securepass123", "Get")
	otherUID := testutil.CreateUser(t, pool, "other@example.com", "securepass123", "Other")
	cookie := testutil.AuthCookie(t, pool, uid)
	otherCookie := testutil.AuthCookie(t, pool, otherUID)

	wsID := testutil.CreateWorkspace(t, pool, uid, "my-ws")

	resp := testutil.DoJSON(t, router, "GET", fmt.Sprintf("/api/workspaces/%s", wsID), nil, cookie)
	assert.Equal(t, http.StatusOK, resp.Code)
	body := testutil.DecodeJSON[map[string]any](t, resp)
	assert.Equal(t, "my-ws", body["name"])

	resp2 := testutil.DoJSON(t, router, "GET", fmt.Sprintf("/api/workspaces/%s", wsID), nil, otherCookie)
	assert.Equal(t, http.StatusNotFound, resp2.Code)
}

func TestDeleteWorkspace(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid := testutil.CreateUser(t, pool, "del-ws@example.com", "securepass123", "Del")
	cookie := testutil.AuthCookie(t, pool, uid)
	wsID := testutil.CreateWorkspace(t, pool, uid, "delete-me")

	resp := testutil.DoJSON(t, router, "DELETE", fmt.Sprintf("/api/workspaces/%s", wsID), nil, cookie)
	assert.Equal(t, http.StatusNoContent, resp.Code)

	resp2 := testutil.DoJSON(t, router, "GET", fmt.Sprintf("/api/workspaces/%s", wsID), nil, cookie)
	assert.Equal(t, http.StatusNotFound, resp2.Code)
}

func TestDeleteWorkspaceForbiddenForMember(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	owner := testutil.CreateUser(t, pool, "owner@example.com", "securepass123", "Owner")
	member := testutil.CreateUser(t, pool, "member@example.com", "securepass123", "Member")
	memberCookie := testutil.AuthCookie(t, pool, member)
	wsID := testutil.CreateWorkspace(t, pool, owner, "owned-ws")

	addMember(t, pool, wsID, member)

	resp := testutil.DoJSON(t, router, "DELETE", fmt.Sprintf("/api/workspaces/%s", wsID), nil, memberCookie)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestAddMember(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	owner := testutil.CreateUser(t, pool, "owner-add@example.com", "securepass123", "Owner")
	newMember := testutil.CreateUser(t, pool, "new-member@example.com", "securepass123", "New")
	ownerCookie := testutil.AuthCookie(t, pool, owner)
	wsID := testutil.CreateWorkspace(t, pool, owner, "add-member-ws")

	resp := testutil.DoJSON(t, router, "POST", fmt.Sprintf("/api/workspaces/%s/members", wsID), map[string]string{
		"email": "new-member@example.com",
	}, ownerCookie)
	assert.Equal(t, http.StatusNoContent, resp.Code)

	newCookie := testutil.AuthCookie(t, pool, newMember)
	resp2 := testutil.DoJSON(t, router, "GET", fmt.Sprintf("/api/workspaces/%s", wsID), nil, newCookie)
	assert.Equal(t, http.StatusOK, resp2.Code)
}

func TestAddMemberNotOwner(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	owner := testutil.CreateUser(t, pool, "owner-no@example.com", "securepass123", "Owner")
	member := testutil.CreateUser(t, pool, "mem-no@example.com", "securepass123", "Mem")
	testutil.CreateUser(t, pool, "target@example.com", "securepass123", "Target")
	memberCookie := testutil.AuthCookie(t, pool, member)
	wsID := testutil.CreateWorkspace(t, pool, owner, "no-add-ws")

	addMember(t, pool, wsID, member)

	resp := testutil.DoJSON(t, router, "POST", fmt.Sprintf("/api/workspaces/%s/members", wsID), map[string]string{
		"email": "target@example.com",
	}, memberCookie)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestRemoveMember(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	owner := testutil.CreateUser(t, pool, "own-rm@example.com", "securepass123", "Owner")
	member := testutil.CreateUser(t, pool, "mem-rm@example.com", "securepass123", "Member")
	ownerCookie := testutil.AuthCookie(t, pool, owner)
	wsID := testutil.CreateWorkspace(t, pool, owner, "rm-member-ws")
	addMember(t, pool, wsID, member)

	resp := testutil.DoJSON(t, router, "DELETE",
		fmt.Sprintf("/api/workspaces/%s/members/%s", wsID, member), nil, ownerCookie)
	assert.Equal(t, http.StatusNoContent, resp.Code)

	memberCookie := testutil.AuthCookie(t, pool, member)
	resp2 := testutil.DoJSON(t, router, "GET", fmt.Sprintf("/api/workspaces/%s", wsID), nil, memberCookie)
	assert.Equal(t, http.StatusNotFound, resp2.Code)
}

func TestRemoveLastOwner(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	owner := testutil.CreateUser(t, pool, "solo@example.com", "securepass123", "Solo")
	ownerCookie := testutil.AuthCookie(t, pool, owner)
	wsID := testutil.CreateWorkspace(t, pool, owner, "solo-ws")

	resp := testutil.DoJSON(t, router, "DELETE",
		fmt.Sprintf("/api/workspaces/%s/members/%s", wsID, owner), nil, ownerCookie)
	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestListMembers(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	owner := testutil.CreateUser(t, pool, "own-list@example.com", "securepass123", "Owner")
	member := testutil.CreateUser(t, pool, "mem-list@example.com", "securepass123", "Member")
	ownerCookie := testutil.AuthCookie(t, pool, owner)
	wsID := testutil.CreateWorkspace(t, pool, owner, "list-members-ws")
	addMember(t, pool, wsID, member)

	resp := testutil.DoJSON(t, router, "GET", fmt.Sprintf("/api/workspaces/%s/members", wsID), nil, ownerCookie)
	assert.Equal(t, http.StatusOK, resp.Code)

	members := testutil.DecodeJSON[[]map[string]any](t, resp)
	require.Len(t, members, 2)
	assert.Equal(t, "owner", members[0]["role"])
}

func addMember(t *testing.T, pool *pgxpool.Pool, wsID, userID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO workspace_members (workspace_id, user_id, role, added_by)
		VALUES ($1, $2, 'member', $2)
		ON CONFLICT DO NOTHING
	`, wsID, userID)
	require.NoError(t, err)
}
