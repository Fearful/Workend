-- +goose Up
-- Stage 43: ordered task chains. A pipeline runs its steps sequentially and
-- short-circuits on first failure. Steps reference tasks; deleting a task
-- cascades and removes the step.
CREATE TABLE pipelines (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX pipelines_project_name_idx ON pipelines (project_id, name);

CREATE TABLE pipeline_steps (
    pipeline_id  UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    position     INT  NOT NULL,
    task_id      UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    PRIMARY KEY (pipeline_id, position)
);

-- A pipeline_run is the parent of N child runs. status mirrors the runs
-- vocabulary plus 'partial' (some children still going).
CREATE TABLE pipeline_runs (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id  UUID        NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    status       TEXT        NOT NULL DEFAULT 'queued',
    started_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX pipeline_runs_pipeline_idx ON pipeline_runs (pipeline_id, created_at DESC);

-- Link child runs to their parent pipeline run + step position.
ALTER TABLE runs
    ADD COLUMN pipeline_run_id UUID REFERENCES pipeline_runs(id) ON DELETE SET NULL,
    ADD COLUMN pipeline_step   INT;
CREATE INDEX runs_pipeline_run_idx ON runs (pipeline_run_id) WHERE pipeline_run_id IS NOT NULL;

-- +goose Down
ALTER TABLE runs DROP COLUMN pipeline_step;
ALTER TABLE runs DROP COLUMN pipeline_run_id;
DROP TABLE pipeline_runs;
DROP TABLE pipeline_steps;
DROP TABLE pipelines;
