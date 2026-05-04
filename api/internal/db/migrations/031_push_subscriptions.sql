-- +goose Up
-- Stage 42: Web Push subscriptions. One row per device-keyed endpoint.
-- The endpoint is the URL the browser gives us when it subscribes; auth
-- and p256dh are the two Web Push encryption keys.
CREATE TABLE push_subscriptions (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint    TEXT        NOT NULL,
    auth        TEXT        NOT NULL,
    p256dh      TEXT        NOT NULL,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX push_subscriptions_endpoint_idx ON push_subscriptions (endpoint);
CREATE INDEX push_subscriptions_user_idx ON push_subscriptions (user_id);

-- +goose Down
DROP TABLE push_subscriptions;
