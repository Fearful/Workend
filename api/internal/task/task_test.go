package task_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workend/api/internal/testutil"
)

func TestListByProject(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid := testutil.CreateUser(t, pool, "tasks@example.com", "securepass123", "Tasks")
	cookie := testutil.AuthCookie(t, pool, uid)
	wsID := testutil.CreateWorkspace(t, pool, uid, "task-ws")
	projID := testutil.CreateProject(t, pool, wsID, "my-proj", "https://github.com/example/repo.git")

	testutil.CreateTask(t, pool, projID, "npm", "build", "npm run build")
	testutil.CreateTask(t, pool, projID, "npm", "test", "npm run test")

	resp := testutil.DoJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%s/tasks", projID), nil, cookie)
	assert.Equal(t, http.StatusOK, resp.Code)

	tasks := testutil.DecodeJSON[[]map[string]any](t, resp)
	require.Len(t, tasks, 2)
	assert.Equal(t, "npm", tasks[0]["source"])
}

func TestListByProjectUnauthorized(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	owner := testutil.CreateUser(t, pool, "own-task@example.com", "securepass123", "Owner")
	other := testutil.CreateUser(t, pool, "other-task@example.com", "securepass123", "Other")
	otherCookie := testutil.AuthCookie(t, pool, other)
	wsID := testutil.CreateWorkspace(t, pool, owner, "priv-ws")
	projID := testutil.CreateProject(t, pool, wsID, "priv-proj", "https://github.com/example/priv.git")
	testutil.CreateTask(t, pool, projID, "npm", "build", "npm run build")

	resp := testutil.DoJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%s/tasks", projID), nil, otherCookie)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestUpdateTask(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid := testutil.CreateUser(t, pool, "upd@example.com", "securepass123", "Upd")
	cookie := testutil.AuthCookie(t, pool, uid)
	wsID := testutil.CreateWorkspace(t, pool, uid, "upd-ws")
	projID := testutil.CreateProject(t, pool, wsID, "upd-proj", "https://github.com/example/upd.git")
	taskID := testutil.CreateTask(t, pool, projID, "npm", "build", "npm run build")

	timeout := 600
	retryMax := 3
	resp := testutil.DoJSON(t, router, "PATCH",
		fmt.Sprintf("/api/tasks/%s", taskID),
		map[string]any{"timeout_seconds": timeout, "retry_max": retryMax},
		cookie)
	assert.Equal(t, http.StatusOK, resp.Code)

	body := testutil.DecodeJSON[map[string]any](t, resp)
	assert.Equal(t, float64(600), body["timeout_seconds"])
	assert.Equal(t, float64(3), body["retry_max"])
}

func TestUpdateTaskValidation(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	uid := testutil.CreateUser(t, pool, "val@example.com", "securepass123", "Val")
	cookie := testutil.AuthCookie(t, pool, uid)
	wsID := testutil.CreateWorkspace(t, pool, uid, "val-ws")
	projID := testutil.CreateProject(t, pool, wsID, "val-proj", "https://github.com/example/val.git")
	taskID := testutil.CreateTask(t, pool, projID, "npm", "build", "npm run build")

	tests := []struct {
		name string
		body map[string]any
	}{
		{"timeout too low", map[string]any{"timeout_seconds": 0}},
		{"timeout too high", map[string]any{"timeout_seconds": 100000}},
		{"retry_max too high", map[string]any{"retry_max": 11}},
		{"retry_max negative", map[string]any{"retry_max": -1}},
		{"invalid supersede_policy", map[string]any{"supersede_policy": "invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testutil.DoJSON(t, router, "PATCH",
				fmt.Sprintf("/api/tasks/%s", taskID), tt.body, cookie)
			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
	}
}
