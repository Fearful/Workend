-- +goose Up
ALTER TABLE pipeline_steps ADD COLUMN condition_expr TEXT NOT NULL DEFAULT '';
ALTER TABLE pipeline_steps ADD COLUMN on_failure TEXT NOT NULL DEFAULT 'stop';

-- +goose Down
ALTER TABLE pipeline_steps DROP COLUMN on_failure;
ALTER TABLE pipeline_steps DROP COLUMN condition_expr;
