package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workend/api/internal/auth"
	"workend/api/internal/testutil"
)

func setupMiddlewareTest(t *testing.T) (*pgxpool.Pool, http.Handler) {
	t.Helper()
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := auth.UserID(r.Context())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(uid.String()))
	})

	handler := auth.RequireUser(pool)(inner)
	return pool, handler
}

func TestRequireUserValidSession(t *testing.T) {
	pool, handler := setupMiddlewareTest(t)
	uid := testutil.CreateUser(t, pool, "valid@example.com", "securepass123", "Valid")
	cookie := testutil.AuthCookie(t, pool, uid)

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uid.String(), rec.Body.String())
}

func TestRequireUserNoCookie(t *testing.T) {
	_, handler := setupMiddlewareTest(t)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireUserInvalidToken(t *testing.T) {
	_, handler := setupMiddlewareTest(t)

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "invalid-token-value"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireUserExpiredSession(t *testing.T) {
	pool, handler := setupMiddlewareTest(t)
	uid := testutil.CreateUser(t, pool, "expired@example.com", "securepass123", "Expired")

	raw, _, err := auth.CreateSession(context.Background(), pool, uid, "test", "127.0.0.1")
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(),
		`UPDATE sessions SET expires_at = $1 WHERE user_id = $2`,
		time.Now().Add(-time.Hour), uid)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: raw})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
