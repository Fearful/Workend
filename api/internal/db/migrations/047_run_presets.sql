-- +goose Up
CREATE TABLE run_presets (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    env_vars   JSONB NOT NULL DEFAULT '{}'::jsonb,
    extra_args TEXT[] NOT NULL DEFAULT '{}',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX run_presets_task_name_idx ON run_presets (task_id, name);

-- +goose Down
DROP TABLE run_presets;
