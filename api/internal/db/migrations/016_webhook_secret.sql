-- +goose Up
-- Stage 24: optional shared secret per project for HMAC webhook validation.
-- When NULL, the URL-token-only auth from Stage 19 still works (back-compat).
ALTER TABLE projects ADD COLUMN webhook_secret TEXT;

-- +goose Down
ALTER TABLE projects DROP COLUMN webhook_secret;
