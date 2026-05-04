-- +goose Up
CREATE TABLE schedules (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id      UUID        NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    project_id   UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    cron_expr    TEXT        NOT NULL,    -- standard 5-field cron: m h dom mon dow
    enabled      BOOLEAN     NOT NULL DEFAULT true,
    last_run_at  TIMESTAMPTZ,
    next_run_at  TIMESTAMPTZ,
    created_by   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX schedules_due_idx       ON schedules (next_run_at) WHERE enabled = true;
CREATE INDEX schedules_project_idx   ON schedules (project_id);
CREATE INDEX schedules_task_idx      ON schedules (task_id);

-- +goose Down
DROP TABLE schedules;
