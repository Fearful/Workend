-- +goose Up
-- Per-task glob patterns describe what files the runner should snapshot
-- after a successful (or failed) run. Stored on tasks so they're declarative
-- and re-applied across re-runs.
ALTER TABLE tasks
    ADD COLUMN artifact_patterns TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[];

-- Captured artifacts. Path is relative to the run's working directory; the
-- on-disk file lives under <ARTIFACTS_ROOT>/<run_id>/<sanitized path> with
-- a stable name so the API can serve it. We don't blob-store contents in
-- Postgres — too big — only metadata.
CREATE TABLE artifacts (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id        UUID        NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    relative_path TEXT        NOT NULL,
    storage_path  TEXT        NOT NULL,
    size_bytes    BIGINT      NOT NULL,
    mime_type     TEXT,
    captured_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX artifacts_run_idx ON artifacts (run_id);
CREATE UNIQUE INDEX artifacts_run_path_idx ON artifacts (run_id, relative_path);

-- +goose Down
DROP INDEX IF EXISTS artifacts_run_path_idx;
DROP INDEX IF EXISTS artifacts_run_idx;
DROP TABLE IF EXISTS artifacts;
ALTER TABLE tasks DROP COLUMN IF EXISTS artifact_patterns;
