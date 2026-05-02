-- +goose Up
CREATE TABLE project_stats (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    computed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    total_files  INT         NOT NULL DEFAULT 0,
    total_lines  INT         NOT NULL DEFAULT 0,
    total_code   INT         NOT NULL DEFAULT 0,
    languages    JSONB       NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX project_stats_project_idx ON project_stats (project_id, computed_at DESC);

-- +goose Down
DROP TABLE project_stats;
