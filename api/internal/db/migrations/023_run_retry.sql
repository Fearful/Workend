-- +goose Up
-- Stage 34: per-task retry policy + per-run attempt tracking.
ALTER TABLE tasks
    ADD COLUMN retry_max         INT NOT NULL DEFAULT 0,
    ADD COLUMN retry_backoff_sec INT NOT NULL DEFAULT 30;

ALTER TABLE runs
    ADD COLUMN attempt        INT  NOT NULL DEFAULT 1,
    ADD COLUMN parent_run_id  UUID REFERENCES runs(id) ON DELETE SET NULL;

CREATE INDEX runs_parent_run_idx ON runs (parent_run_id) WHERE parent_run_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS runs_parent_run_idx;
ALTER TABLE runs  DROP COLUMN parent_run_id;
ALTER TABLE runs  DROP COLUMN attempt;
ALTER TABLE tasks DROP COLUMN retry_backoff_sec;
ALTER TABLE tasks DROP COLUMN retry_max;
