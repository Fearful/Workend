-- +goose Up
CREATE TABLE registry_configs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    registry_url   TEXT NOT NULL,
    repository     TEXT NOT NULL,
    username       TEXT NOT NULL DEFAULT '',
    password_enc   BYTEA,
    tag_pattern    TEXT NOT NULL DEFAULT '{{branch}}-{{short_sha}}',
    auto_push      BOOLEAN NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(project_id)
);

-- +goose Down
DROP TABLE registry_configs;
