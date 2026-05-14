package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workend/api/internal/auth"
	"workend/api/internal/testutil"
)

func TestSignup(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	resp := testutil.DoJSON(t, router, "POST", "/api/auth/signup", map[string]string{
		"email":        "alice@example.com",
		"password":     "securepass123",
		"display_name": "Alice",
	}, nil)

	assert.Equal(t, http.StatusCreated, resp.Code)

	body := testutil.DecodeJSON[map[string]any](t, resp)
	assert.Equal(t, "alice@example.com", body["email"])
	assert.Equal(t, "Alice", body["display_name"])
	assert.NotEmpty(t, body["id"])

	cookies := resp.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == auth.SessionCookieName {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie, "session cookie should be set")
}

func TestSignupDuplicateEmail(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	body := map[string]string{
		"email": "dup@example.com", "password": "securepass123", "display_name": "Dup",
	}
	resp1 := testutil.DoJSON(t, router, "POST", "/api/auth/signup", body, nil)
	assert.Equal(t, http.StatusCreated, resp1.Code)

	resp2 := testutil.DoJSON(t, router, "POST", "/api/auth/signup", body, nil)
	assert.Equal(t, http.StatusConflict, resp2.Code)
}

func TestSignupValidation(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	tests := []struct {
		name string
		body map[string]string
	}{
		{"missing email", map[string]string{"password": "12345678", "display_name": "X"}},
		{"missing password", map[string]string{"email": "a@b.com", "display_name": "X"}},
		{"missing display_name", map[string]string{"email": "a@b.com", "password": "12345678"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testutil.DoJSON(t, router, "POST", "/api/auth/signup", tt.body, nil)
			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
	}
}

func TestSignupShortPassword(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	resp := testutil.DoJSON(t, router, "POST", "/api/auth/signup", map[string]string{
		"email": "short@example.com", "password": "1234567", "display_name": "Short",
	}, nil)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestLogin(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)
	testutil.CreateUser(t, pool, "login@example.com", "securepass123", "Login User")

	resp := testutil.DoJSON(t, router, "POST", "/api/auth/login", map[string]string{
		"email": "login@example.com", "password": "securepass123",
	}, nil)

	assert.Equal(t, http.StatusOK, resp.Code)

	body := testutil.DecodeJSON[map[string]any](t, resp)
	assert.Equal(t, "login@example.com", body["email"])
	assert.Equal(t, "Login User", body["display_name"])

	cookies := resp.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == auth.SessionCookieName {
			found = true
			break
		}
	}
	assert.True(t, found, "session cookie should be set on login")
}

func TestLoginWrongPassword(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)
	testutil.CreateUser(t, pool, "wrong@example.com", "correctpass", "Wrong")

	resp := testutil.DoJSON(t, router, "POST", "/api/auth/login", map[string]string{
		"email": "wrong@example.com", "password": "wrongpass",
	}, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestLoginNonexistentUser(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	resp := testutil.DoJSON(t, router, "POST", "/api/auth/login", map[string]string{
		"email": "ghost@example.com", "password": "somepass123",
	}, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestLogout(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)
	uid := testutil.CreateUser(t, pool, "logout@example.com", "securepass123", "Logout")
	cookie := testutil.AuthCookie(t, pool, uid)

	resp := testutil.DoJSON(t, router, "POST", "/api/auth/logout", nil, cookie)
	assert.Equal(t, http.StatusNoContent, resp.Code)

	cleared := resp.Result().Cookies()
	for _, c := range cleared {
		if c.Name == auth.SessionCookieName {
			assert.True(t, c.MaxAge < 0, "cookie should be cleared")
		}
	}
}

func TestMe(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)
	uid := testutil.CreateUser(t, pool, "me@example.com", "securepass123", "Me User")
	cookie := testutil.AuthCookie(t, pool, uid)

	resp := testutil.DoJSON(t, router, "GET", "/api/me", nil, cookie)
	assert.Equal(t, http.StatusOK, resp.Code)

	body := testutil.DecodeJSON[map[string]any](t, resp)
	assert.Equal(t, "me@example.com", body["email"])
	assert.Equal(t, "Me User", body["display_name"])
	assert.Equal(t, uid.String(), body["id"])
}

func TestMeUnauthenticated(t *testing.T) {
	pool := testutil.SetupDB(t)
	testutil.TruncateAll(t, pool)
	router := testutil.NewRouter(t, pool)

	resp := testutil.DoJSON(t, router, "GET", "/api/me", nil, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
