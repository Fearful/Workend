-- +goose Up
-- Stage 15: rename user_tokens -> provider_connections, add columns to support
-- multiple instances per provider type (e.g., gitlab.com + a self-hosted
-- gitlab) and lazy token refresh.
ALTER TABLE user_tokens RENAME TO provider_connections;

-- Existing rows: GitHub-only from Stage 9. Set instance_url, drop the old PK,
-- add a real surrogate key + uniqueness on (user_id, provider, instance_url).
ALTER TABLE provider_connections ADD COLUMN id UUID NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE provider_connections ADD COLUMN instance_url TEXT NOT NULL DEFAULT 'https://github.com';
ALTER TABLE provider_connections ADD COLUMN refresh_token BYTEA;
ALTER TABLE provider_connections ADD COLUMN expires_at TIMESTAMPTZ;

ALTER TABLE provider_connections DROP CONSTRAINT user_tokens_pkey;
ALTER TABLE provider_connections ADD PRIMARY KEY (id);
ALTER TABLE provider_connections ADD CONSTRAINT provider_connections_unique UNIQUE (user_id, provider, instance_url);

CREATE INDEX provider_connections_user_idx ON provider_connections (user_id);

-- +goose Down
ALTER TABLE provider_connections DROP CONSTRAINT provider_connections_unique;
ALTER TABLE provider_connections DROP CONSTRAINT provider_connections_pkey;
ALTER TABLE provider_connections DROP COLUMN id;
ALTER TABLE provider_connections DROP COLUMN instance_url;
ALTER TABLE provider_connections DROP COLUMN refresh_token;
ALTER TABLE provider_connections DROP COLUMN expires_at;
ALTER TABLE provider_connections ADD PRIMARY KEY (user_id, provider);
ALTER TABLE provider_connections RENAME TO user_tokens;
