-- +goose Up
CREATE TABLE remote_pipeline_runs (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider_run_id TEXT        NOT NULL,
    status          TEXT        NOT NULL,
    branch          TEXT,
    commit_sha      TEXT,
    workflow_name   TEXT,
    html_url        TEXT,
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX remote_pipeline_runs_project_idx
    ON remote_pipeline_runs (project_id, fetched_at DESC);
CREATE UNIQUE INDEX remote_pipeline_runs_provider_idx
    ON remote_pipeline_runs (project_id, provider_run_id);

ALTER TABLE projects ADD COLUMN last_pipeline_fetch_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE projects DROP COLUMN last_pipeline_fetch_at;
DROP TABLE remote_pipeline_runs;
