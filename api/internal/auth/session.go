package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	SessionCookieName = "workend_session"
	tokenBytes        = 32
	SessionTTL        = 30 * 24 * time.Hour
)

var ErrSessionNotFound = errors.New("session not found or expired")

type Session struct {
	UserID     uuid.UUID
	ExpiresAt  time.Time
	LastUsedAt time.Time
}

// GenerateToken returns (rawToken, hash). Store the hash; give the raw token
// to the client.
func GenerateToken() (string, []byte, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	hash := hashToken(raw)
	return raw, hash, nil
}

func hashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func CreateSession(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, userAgent, ip string) (rawToken string, expiresAt time.Time, err error) {
	raw, hash, err := GenerateToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt = time.Now().Add(SessionTTL)

	_, err = pool.Exec(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at, user_agent, ip)
		VALUES ($1, $2, $3, $4, $5)
	`, hash, userID, expiresAt, nullable(userAgent), nullable(ip))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("insert session: %w", err)
	}
	return raw, expiresAt, nil
}

func LookupSession(ctx context.Context, pool *pgxpool.Pool, rawToken string) (*Session, error) {
	hash := hashToken(rawToken)
	var s Session
	err := pool.QueryRow(ctx, `
		SELECT user_id, expires_at, last_used_at
		FROM sessions
		WHERE token_hash = $1 AND expires_at > now()
	`, hash).Scan(&s.UserID, &s.ExpiresAt, &s.LastUsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lookup session: %w", err)
	}

	// Touch last_used_at; ignore error (best-effort)
	_, _ = pool.Exec(ctx, `UPDATE sessions SET last_used_at = now() WHERE token_hash = $1`, hash)
	return &s, nil
}

func DeleteSession(ctx context.Context, pool *pgxpool.Pool, rawToken string) error {
	hash := hashToken(rawToken)
	_, err := pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, hash)
	return err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
