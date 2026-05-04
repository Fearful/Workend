-- +goose Up
ALTER TABLE users
    ADD COLUMN oidc_sub    TEXT,
    ADD COLUMN oidc_issuer TEXT;
CREATE UNIQUE INDEX users_oidc_sub_issuer_idx ON users (oidc_sub, oidc_issuer)
    WHERE oidc_sub IS NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
DROP INDEX IF EXISTS users_oidc_sub_issuer_idx;
ALTER TABLE users DROP COLUMN IF EXISTS oidc_issuer;
ALTER TABLE users DROP COLUMN IF EXISTS oidc_sub;
