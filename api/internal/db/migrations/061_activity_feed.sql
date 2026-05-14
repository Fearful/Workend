-- +goose Up
CREATE TABLE activity_feed (
    id            BIGSERIAL PRIMARY KEY,
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type    TEXT NOT NULL,
    entity_type   TEXT NOT NULL,
    entity_id     UUID,
    summary       TEXT NOT NULL,
    metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX activity_feed_ws_recent ON activity_feed (workspace_id, created_at DESC);

-- +goose Down
DROP TABLE activity_feed;
