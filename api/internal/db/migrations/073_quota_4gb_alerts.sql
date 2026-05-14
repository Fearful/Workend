-- +goose Up
ALTER TABLE users ALTER COLUMN quota_bytes SET DEFAULT 4294967296;
UPDATE users SET quota_bytes = 4294967296 WHERE quota_bytes = 5368709120;

CREATE TABLE IF NOT EXISTS quota_alerts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    threshold  INT NOT NULL,
    used_bytes BIGINT NOT NULL,
    quota_bytes BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, threshold)
);

CREATE INDEX idx_quota_alerts_user ON quota_alerts(user_id);

-- +goose Down
ALTER TABLE users ALTER COLUMN quota_bytes SET DEFAULT 5368709120;
DROP TABLE IF EXISTS quota_alerts;
