-- +goose Up
ALTER TABLE tasks ADD COLUMN quarantined BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE tasks ADD COLUMN quarantined_at TIMESTAMPTZ;
ALTER TABLE tasks ADD COLUMN quarantine_reason TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE tasks DROP COLUMN quarantine_reason;
ALTER TABLE tasks DROP COLUMN quarantined_at;
ALTER TABLE tasks DROP COLUMN quarantined;
