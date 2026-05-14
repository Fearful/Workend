-- +goose Up
CREATE TABLE alert_rules (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id        UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    rule_type      TEXT NOT NULL,
    threshold      INT NOT NULL,
    window_minutes INT NOT NULL DEFAULT 60,
    enabled        BOOLEAN NOT NULL DEFAULT true,
    last_triggered_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX alert_rules_task_idx ON alert_rules (task_id) WHERE enabled;
CREATE UNIQUE INDEX alert_rules_user_task_type_idx ON alert_rules (user_id, task_id, rule_type);

-- +goose Down
DROP TABLE alert_rules;
