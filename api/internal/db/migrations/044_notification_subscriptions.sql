-- +goose Up
-- Notification subscriptions tie a user's notification_config (channel) to a
-- scope: global (all the user's runs), workspace, project, or task. Severity
-- filters within that scope: 'all' | 'failures' | 'off'. A user can have any
-- number of subscriptions. The dispatcher only fires a config when at least
-- one of its subscriptions matches the run AND the severity passes.
--
-- Default behavior: when a config is created, we auto-add a single
-- ('global', NULL, 'failures') subscription so existing users keep their
-- on_failure pings. The migration backfills the same row for every existing
-- config.
CREATE TABLE notification_subscriptions (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    config_id   UUID        NOT NULL REFERENCES notification_configs(id) ON DELETE CASCADE,
    scope_type  TEXT        NOT NULL,
    scope_id    UUID,
    severity    TEXT        NOT NULL DEFAULT 'failures',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE notification_subscriptions
    ADD CONSTRAINT notification_subscriptions_scope_chk
    CHECK (scope_type IN ('global','workspace','project','task'));

ALTER TABLE notification_subscriptions
    ADD CONSTRAINT notification_subscriptions_severity_chk
    CHECK (severity IN ('all','failures','off'));

ALTER TABLE notification_subscriptions
    ADD CONSTRAINT notification_subscriptions_scope_id_chk
    CHECK ((scope_type = 'global' AND scope_id IS NULL) OR (scope_type <> 'global' AND scope_id IS NOT NULL));

CREATE UNIQUE INDEX notification_subs_uniq_idx
    ON notification_subscriptions (user_id, config_id, scope_type, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'::uuid));

CREATE INDEX notification_subs_user_idx ON notification_subscriptions (user_id);

-- Backfill: one global-failures subscription per existing notification_config.
INSERT INTO notification_subscriptions (user_id, config_id, scope_type, scope_id, severity)
SELECT user_id, id, 'global', NULL, 'failures'
FROM notification_configs
WHERE NOT EXISTS (
    SELECT 1 FROM notification_subscriptions s WHERE s.config_id = notification_configs.id
);

-- +goose Down
DROP INDEX IF EXISTS notification_subs_user_idx;
DROP INDEX IF EXISTS notification_subs_uniq_idx;
DROP TABLE IF EXISTS notification_subscriptions;
