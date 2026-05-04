-- +goose Up
CREATE TABLE pipeline_configs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ci_system   TEXT        NOT NULL,
    file_path   TEXT        NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX pipeline_configs_project_file_idx ON pipeline_configs (project_id, file_path);
CREATE INDEX pipeline_configs_project_idx ON pipeline_configs (project_id);

-- +goose Down
DROP TABLE pipeline_configs;
