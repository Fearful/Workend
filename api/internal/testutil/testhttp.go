package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/config"
	"workend/api/internal/server"
)

func NewRouter(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()
	cfg := &config.Config{
		ListenAddr:            ":0",
		ReposRoot:             t.TempDir(),
		LogsRoot:              t.TempDir(),
		ArtifactsRoot:         t.TempDir(),
		DefaultTaskTimeoutSec: 300,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := server.New(cfg, pool, nil, nil, nil, logger, nil)
	return srv.Router()
}

func DoJSON(t *testing.T, router http.Handler, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func DecodeJSON[T any](t *testing.T, resp *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, resp.Body.String())
	}
	return v
}
