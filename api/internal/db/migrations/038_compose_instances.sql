-- +goose Up
CREATE TABLE compose_instances (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID        NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    compose_file        TEXT        NOT NULL,
    status              TEXT        NOT NULL DEFAULT 'stopped',
    env_vars_encrypted  BYTEA,
    started_at          TIMESTAMPTZ,
    stopped_at          TIMESTAMPTZ,
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE compose_instances;
