-- +goose Up
-- Stage 30: capture optional per-run inputs (env vars + extra args). Stored
-- alongside the run so the history shows what was actually executed.
ALTER TABLE runs ADD COLUMN params JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE runs DROP COLUMN params;
