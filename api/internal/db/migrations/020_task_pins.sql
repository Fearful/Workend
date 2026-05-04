-- +goose Up
-- Stage 31: per-user task pins for one-click access from the dashboard.
CREATE TABLE task_pins (
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id    UUID        NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    position   INT         NOT NULL DEFAULT 0,
    pinned_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, task_id)
);

CREATE INDEX task_pins_user_idx ON task_pins (user_id, position);

-- +goose Down
DROP TABLE task_pins;
