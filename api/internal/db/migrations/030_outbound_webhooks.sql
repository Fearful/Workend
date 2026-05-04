-- +goose Up
-- Stage 41: HMAC-signed outbound webhooks. Distinct from notification_configs
-- (which is "tell me when a run finishes"); this is "build me a generic
-- automation pipeline" — events for run.created, run.started, run.finished,
-- project.synced, image.built, etc., with a shared secret for verification.
CREATE TABLE outbound_webhook_targets (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    url         TEXT        NOT NULL,
    secret      TEXT        NOT NULL,        -- shared secret for HMAC-SHA256 of the request body
    events      JSONB       NOT NULL DEFAULT '[]'::jsonb, -- subscribed event types; ['*'] = all
    enabled     BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_fired_at TIMESTAMPTZ,
    last_status   INT
);
CREATE INDEX outbound_webhook_targets_user_idx ON outbound_webhook_targets (user_id);

-- +goose Down
DROP TABLE outbound_webhook_targets;
