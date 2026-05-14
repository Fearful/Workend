package pin_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workend/api/internal/testutil"
)

func TestPinUnpin(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid := testutil.CreateUser(t, pool, "pin@example.com", "securepass123", "Pin")
	cookie := testutil.AuthCookie(t, pool, uid)
	wsID := testutil.CreateWorkspace(t, pool, uid, "pin-ws")
	projID := testutil.CreateProject(t, pool, wsID, "pin-proj", "https://github.com/example/pin.git")
	taskID := testutil.CreateTask(t, pool, projID, "npm", "build", "npm run build")

	// Pin
	resp := testutil.DoJSON(t, router, "POST", fmt.Sprintf("/api/tasks/%s/pin", taskID), nil, cookie)
	assert.Equal(t, http.StatusNoContent, resp.Code)

	// List pinned
	resp2 := testutil.DoJSON(t, router, "GET", "/api/me/pinned-tasks", nil, cookie)
	assert.Equal(t, http.StatusOK, resp2.Code)
	pinned := testutil.DecodeJSON[[]map[string]any](t, resp2)
	require.Len(t, pinned, 1)
	assert.Equal(t, taskID.String(), pinned[0]["task_id"])

	// Unpin
	resp3 := testutil.DoJSON(t, router, "DELETE", fmt.Sprintf("/api/tasks/%s/pin", taskID), nil, cookie)
	assert.Equal(t, http.StatusNoContent, resp3.Code)

	// List should be empty
	resp4 := testutil.DoJSON(t, router, "GET", "/api/me/pinned-tasks", nil, cookie)
	assert.Equal(t, http.StatusOK, resp4.Code)
	pinned2 := testutil.DecodeJSON[[]map[string]any](t, resp4)
	assert.Len(t, pinned2, 0)
}

func TestPinIdempotent(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid := testutil.CreateUser(t, pool, "pin-idem@example.com", "securepass123", "PI")
	cookie := testutil.AuthCookie(t, pool, uid)
	wsID := testutil.CreateWorkspace(t, pool, uid, "pin-idem-ws")
	projID := testutil.CreateProject(t, pool, wsID, "pin-idem-proj", "https://github.com/example/idem.git")
	taskID := testutil.CreateTask(t, pool, projID, "npm", "test", "npm test")

	for i := 0; i < 3; i++ {
		resp := testutil.DoJSON(t, router, "POST", fmt.Sprintf("/api/tasks/%s/pin", taskID), nil, cookie)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	}

	resp := testutil.DoJSON(t, router, "GET", "/api/me/pinned-tasks", nil, cookie)
	pinned := testutil.DecodeJSON[[]map[string]any](t, resp)
	assert.Len(t, pinned, 1)
}

func TestUnpinIdempotent(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid := testutil.CreateUser(t, pool, "unpin-idem@example.com", "securepass123", "UI")
	cookie := testutil.AuthCookie(t, pool, uid)
	wsID := testutil.CreateWorkspace(t, pool, uid, "unpin-idem-ws")
	projID := testutil.CreateProject(t, pool, wsID, "unpin-idem-proj", "https://github.com/example/unpin.git")
	taskID := testutil.CreateTask(t, pool, projID, "npm", "test", "npm test")

	for i := 0; i < 3; i++ {
		resp := testutil.DoJSON(t, router, "DELETE", fmt.Sprintf("/api/tasks/%s/pin", taskID), nil, cookie)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	}
}

func TestPinUnauthorized(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	owner := testutil.CreateUser(t, pool, "pin-own@example.com", "securepass123", "Owner")
	other := testutil.CreateUser(t, pool, "pin-other@example.com", "securepass123", "Other")
	otherCookie := testutil.AuthCookie(t, pool, other)
	wsID := testutil.CreateWorkspace(t, pool, owner, "pin-priv-ws")
	projID := testutil.CreateProject(t, pool, wsID, "pin-priv-proj", "https://github.com/example/priv.git")
	taskID := testutil.CreateTask(t, pool, projID, "npm", "build", "npm run build")

	resp := testutil.DoJSON(t, router, "POST", fmt.Sprintf("/api/tasks/%s/pin", taskID), nil, otherCookie)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}
