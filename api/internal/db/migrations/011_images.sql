-- +goose Up
CREATE TABLE images (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    run_id          UUID        REFERENCES runs(id) ON DELETE SET NULL,
    dockerfile_path TEXT        NOT NULL,
    digest          TEXT,
    size_bytes      BIGINT,
    commit_sha      TEXT,
    built_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX images_project_idx ON images (project_id, built_at DESC);

-- +goose Down
DROP TABLE images;
