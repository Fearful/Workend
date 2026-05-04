-- +goose Up
-- Stage 29: per-task timeout enforcement. NULL means "use the global default
-- from WORKEND_DEFAULT_TASK_TIMEOUT_SECONDS"; a positive value overrides it.
ALTER TABLE tasks ADD COLUMN timeout_seconds INT;

-- Marker on runs for "killed because we hit the deadline" so the UI can
-- distinguish a deliberate failure from a timeout without a new status enum.
ALTER TABLE runs ADD COLUMN timed_out BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE runs  DROP COLUMN timed_out;
ALTER TABLE tasks DROP COLUMN timeout_seconds;
