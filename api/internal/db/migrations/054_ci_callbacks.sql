-- +goose Up
ALTER TABLE remote_pipeline_runs ADD COLUMN callback_url TEXT NOT NULL DEFAULT '';
ALTER TABLE remote_pipeline_runs ADD COLUMN callback_sent_at TIMESTAMPTZ;

CREATE TABLE ci_trigger_rules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    trigger_on    TEXT NOT NULL,
    remote_pipeline TEXT NOT NULL,
    enabled       BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ci_trigger_rules_project_idx ON ci_trigger_rules (project_id);

-- +goose Down
DROP TABLE ci_trigger_rules;
ALTER TABLE remote_pipeline_runs DROP COLUMN callback_sent_at;
ALTER TABLE remote_pipeline_runs DROP COLUMN callback_url;
