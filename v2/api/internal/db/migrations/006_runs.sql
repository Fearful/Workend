-- +goose Up
CREATE TABLE runs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id     UUID        NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    commit_sha  TEXT,
    -- queued | running | succeeded | failed | cancelled
    status      TEXT        NOT NULL DEFAULT 'queued',
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    exit_code   INT,
    log_path    TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX runs_task_id_idx    ON runs (task_id, created_at DESC);
CREATE INDEX runs_project_id_idx ON runs (project_id, created_at DESC);
CREATE INDEX runs_status_idx     ON runs (status);

-- +goose Down
DROP TABLE runs;
