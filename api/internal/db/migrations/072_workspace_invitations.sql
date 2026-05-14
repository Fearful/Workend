-- +goose Up
CREATE TABLE workspace_invitations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email        TEXT NOT NULL,
    invited_by   UUID NOT NULL REFERENCES users(id),
    status       TEXT NOT NULL DEFAULT 'pending', -- pending, accepted, expired
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL DEFAULT now() + interval '7 days',
    UNIQUE (workspace_id, email)
);

-- +goose Down
DROP TABLE workspace_invitations;
