-- +goose Up
CREATE TABLE user_tokens (
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider     TEXT        NOT NULL,    -- 'github' (extensible to gitlab, bitbucket)
    access_token BYTEA       NOT NULL,    -- secretbox-encrypted with WORKEND_TOKEN_KEY
    scopes       TEXT,
    handle       TEXT,                    -- the provider-side username/handle
    connected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, provider)
);

-- +goose Down
DROP TABLE user_tokens;
