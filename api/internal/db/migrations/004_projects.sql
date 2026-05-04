-- +goose Up
-- Projects table is created in Stage 2 to lock the workspace -> project FK
-- early, but is not used by the API until Stage 3 (repo ingestion).
CREATE TABLE projects (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id        UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name                TEXT        NOT NULL,
    git_url             TEXT        NOT NULL,
    default_branch      TEXT,
    local_path          TEXT,
    status              TEXT        NOT NULL DEFAULT 'pending',
    last_commit_sha     TEXT,
    last_commit_message TEXT,
    last_commit_author  TEXT,
    last_synced_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX projects_workspace_name_idx ON projects (workspace_id, lower(name));
CREATE INDEX projects_status_idx ON projects (status);

-- +goose Down
DROP TABLE projects;
