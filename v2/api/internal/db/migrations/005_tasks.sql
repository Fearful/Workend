-- +goose Up
CREATE TABLE tasks (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source      TEXT        NOT NULL,    -- 'npm' | 'just' (extensible)
    name        TEXT        NOT NULL,
    raw_command TEXT        NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX tasks_project_source_name_idx ON tasks (project_id, source, name);
CREATE INDEX tasks_project_idx ON tasks (project_id);

-- +goose Down
DROP TABLE tasks;
