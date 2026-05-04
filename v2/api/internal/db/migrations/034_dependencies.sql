-- +goose Up
-- Stage 44: parsed lockfile dependencies. Refreshed on every sync; one row
-- per (project, ecosystem, name, version). Depth=0 for top-level deps,
-- depth>0 for transitives where the ecosystem exposes them.
CREATE TABLE project_dependencies (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ecosystem    TEXT        NOT NULL,    -- 'npm' | 'go' | 'cargo' | 'pypi'
    name         TEXT        NOT NULL,
    version      TEXT        NOT NULL,
    depth        INT         NOT NULL DEFAULT 0,
    parsed_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX project_deps_unique_idx
    ON project_dependencies (project_id, ecosystem, name, version);
CREATE INDEX project_deps_project_idx
    ON project_dependencies (project_id, ecosystem, name);

-- +goose Down
DROP TABLE project_dependencies;
