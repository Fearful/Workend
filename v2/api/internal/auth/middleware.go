package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxKey int

const userIDKey ctxKey = 1

// RequireUser is middleware that loads the session from the cookie and rejects
// requests with no valid session. The user ID is stored in the request context.
func RequireUser(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			sess, err := LookupSession(r.Context(), pool, cookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, sess.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserID returns the authenticated user's ID from the request context.
// Panics if no user is set — only call from handlers behind RequireUser.
func UserID(ctx context.Context) uuid.UUID {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		panic("auth.UserID called without RequireUser middleware")
	}
	return id
}
