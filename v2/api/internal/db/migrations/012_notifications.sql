-- +goose Up
CREATE TABLE notification_configs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- 'webhook' | 'slack' | 'email'
    kind        TEXT        NOT NULL,
    -- For webhook/slack: target URL. For email: recipient address.
    target      TEXT        NOT NULL,
    -- Trigger filter: 'on_failure' | 'on_status_change' | 'always'
    trigger     TEXT        NOT NULL DEFAULT 'on_failure',
    enabled     BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX notification_configs_user_idx ON notification_configs (user_id);

-- +goose Down
DROP TABLE notification_configs;
