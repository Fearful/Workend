-- +goose Up
-- Stage 43: per-user personal access tokens for HTTPS git clones. Lets users
-- store a PAT (GitHub/GitLab/Gitea/etc.) on instances where the admin hasn't
-- configured OAuth, or for hosts no provider is registered for. Token bytes
-- encrypted at rest with the same NaCl box used for OAuth tokens and SSH
-- keys (requires WORKEND_TOKEN_KEY).
CREATE TABLE user_pat_credentials (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    host            TEXT        NOT NULL,
    label           TEXT        NOT NULL,
    encrypted_token BYTEA       NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX user_pat_credentials_user_host_idx ON user_pat_credentials (user_id, host);

-- +goose Down
DROP TABLE user_pat_credentials;
