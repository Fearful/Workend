-- +goose Up
CREATE TABLE task_templates (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id   UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    source         TEXT NOT NULL,
    raw_command    TEXT NOT NULL,
    base_image     TEXT NOT NULL DEFAULT '',
    env_vars       JSONB NOT NULL DEFAULT '{}'::jsonb,
    timeout_seconds INT,
    retry_max      INT NOT NULL DEFAULT 0,
    created_by     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX task_templates_ws_name_idx ON task_templates (workspace_id, name);

-- +goose Down
DROP TABLE task_templates;
