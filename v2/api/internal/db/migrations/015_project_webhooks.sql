-- +goose Up
-- Stage 19: per-project webhook tokens for push-triggered syncs.
-- Generate one-shot at project creation; stored in plaintext (token in
-- URL is the auth — same model as e.g., Discord webhooks).
ALTER TABLE projects ADD COLUMN webhook_token TEXT;

-- Backfill existing projects with random tokens.
UPDATE projects SET webhook_token = encode(gen_random_bytes(24), 'base64')
WHERE webhook_token IS NULL;

ALTER TABLE projects ALTER COLUMN webhook_token SET NOT NULL;
ALTER TABLE projects ADD CONSTRAINT projects_webhook_token_unique UNIQUE (webhook_token);

-- +goose Down
ALTER TABLE projects DROP CONSTRAINT projects_webhook_token_unique;
ALTER TABLE projects DROP COLUMN webhook_token;
