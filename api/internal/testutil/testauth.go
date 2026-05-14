package testutil

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"workend/api/internal/auth"
)

func CreateUser(t *testing.T, pool *pgxpool.Pool, email, password, displayName string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var id uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ($1, $2, $3) RETURNING id
	`, email, hash, displayName).Scan(&id)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}

func AuthCookie(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) *http.Cookie {
	t.Helper()
	ctx := context.Background()

	raw, expiresAt, err := auth.CreateSession(ctx, pool, userID, "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	return &http.Cookie{
		Name:    auth.SessionCookieName,
		Value:   raw,
		Expires: expiresAt,
	}
}

func CreateWorkspace(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO workspaces (user_id, name) VALUES ($1, $2) RETURNING id
	`, userID, name).Scan(&id)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role, added_by)
		VALUES ($1, $2, 'owner', $2)
	`, id, userID)
	if err != nil {
		t.Fatalf("create workspace membership: %v", err)
	}
	return id
}

func CreateProject(t *testing.T, pool *pgxpool.Pool, workspaceID uuid.UUID, name, gitURL string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO projects (workspace_id, name, git_url, status, webhook_token)
		VALUES ($1, $2, $3, 'ready', encode(gen_random_bytes(24), 'base64')) RETURNING id
	`, workspaceID, name, gitURL).Scan(&id)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return id
}

func CreateTask(t *testing.T, pool *pgxpool.Pool, projectID uuid.UUID, source, name, rawCmd string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO tasks (project_id, source, name, raw_command)
		VALUES ($1, $2, $3, $4) RETURNING id
	`, projectID, source, name, rawCmd).Scan(&id)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	return id
}
