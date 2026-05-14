-- +goose Up
CREATE TABLE monorepo_packages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    path        TEXT NOT NULL, -- relative path within repo, e.g. "packages/auth"
    pkg_type    TEXT NOT NULL DEFAULT 'unknown', -- npm, go, cargo, python, gradle
    version     TEXT NOT NULL DEFAULT '',
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, path)
);
CREATE INDEX monorepo_packages_project_idx ON monorepo_packages (project_id);

CREATE TABLE package_task_scopes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id  UUID NOT NULL REFERENCES monorepo_packages(id) ON DELETE CASCADE,
    task_id     UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (package_id, task_id)
);

-- +goose Down
DROP TABLE package_task_scopes;
DROP TABLE monorepo_packages;
