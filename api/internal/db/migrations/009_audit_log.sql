-- +goose Up
CREATE TABLE audit_log (
    id          BIGSERIAL   PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor_id    UUID        REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT        NOT NULL,    -- e.g., 'user.signup', 'project.create', 'run.start', 'run.cancel'
    target_kind TEXT,                    -- 'user' | 'workspace' | 'project' | 'run' | ...
    target_id   TEXT,                    -- text to accommodate UUIDs from various tables
    ip          TEXT,
    metadata    JSONB
);

CREATE INDEX audit_log_actor_idx       ON audit_log (actor_id, occurred_at DESC);
CREATE INDEX audit_log_action_idx      ON audit_log (action, occurred_at DESC);
CREATE INDEX audit_log_occurred_at_idx ON audit_log (occurred_at DESC);

-- +goose Down
DROP TABLE audit_log;
