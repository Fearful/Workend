-- +goose Up
CREATE TABLE task_file_inputs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    pattern    TEXT NOT NULL, -- glob pattern like "src/**/*.go", "Dockerfile", "package.json"
    input_type TEXT NOT NULL DEFAULT 'file', -- 'file', 'directory', 'config'
    auto_detected BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, pattern)
);
CREATE INDEX task_file_inputs_task_idx ON task_file_inputs (task_id);

-- +goose Down
DROP TABLE task_file_inputs;
