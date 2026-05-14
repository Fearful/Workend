-- +goose Up
ALTER TABLE tasks ADD COLUMN env_vars JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE tasks DROP COLUMN env_vars;
