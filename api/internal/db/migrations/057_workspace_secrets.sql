-- +goose Up
CREATE TABLE workspace_secrets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    key           TEXT NOT NULL,
    value_enc     BYTEA NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    created_by    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX workspace_secrets_ws_key_idx ON workspace_secrets (workspace_id, key);

-- +goose Down
DROP TABLE workspace_secrets;
