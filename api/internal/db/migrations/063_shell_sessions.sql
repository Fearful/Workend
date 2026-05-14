-- +goose Up
CREATE TABLE shell_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sandbox_id  UUID NOT NULL REFERENCES sandboxes(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    shell       TEXT NOT NULL DEFAULT '/bin/sh',
    status      TEXT NOT NULL DEFAULT 'active', -- active, closed
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at    TIMESTAMPTZ
);
CREATE INDEX shell_sessions_sandbox_idx ON shell_sessions (sandbox_id);

-- +goose Down
DROP TABLE shell_sessions;
