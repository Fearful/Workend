package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workend/api/internal/health"
	"workend/api/internal/testutil"
)

func TestHealthDBOK(t *testing.T) {
	pool := testutil.SetupDB(t)

	handler := &health.Handler{
		Pool:           pool,
		DaggerSockPath: "/nonexistent/dagger.sock",
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var status health.Status
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&status))
	assert.Equal(t, "ok", status.Checks["db"])
	assert.Contains(t, status.Checks["dagger"], "error")
}

func TestHealthDBDown(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://test:test@localhost:1/nonexistent?sslmode=disable")
	require.NoError(t, err)
	cfg.MaxConns = 1

	deadPool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)
	defer deadPool.Close()

	handler := &health.Handler{
		Pool:           deadPool,
		DaggerSockPath: "/nonexistent/dagger.sock",
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var status health.Status
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&status))
	assert.Equal(t, "degraded", status.Status)
	assert.Contains(t, status.Checks["db"], "error")
}
