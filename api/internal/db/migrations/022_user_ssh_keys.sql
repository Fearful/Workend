-- +goose Up
-- Stage 33: SSH-key git clone support. Private keys encrypted at rest with
-- the same NaCl box used for OAuth tokens (requires WORKEND_TOKEN_KEY).
CREATE TABLE user_ssh_keys (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                  TEXT        NOT NULL,
    encrypted_private_key BYTEA       NOT NULL,
    public_key            TEXT        NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX user_ssh_keys_user_name_idx ON user_ssh_keys (user_id, name);

-- +goose Down
DROP TABLE user_ssh_keys;
