-- +goose Up
CREATE TABLE previews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    branch          TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'building', -- building, live, failed, stopped
    url             TEXT NOT NULL DEFAULT '',
    deploy_log      TEXT NOT NULL DEFAULT '',
    auto_deploy     BOOLEAN NOT NULL DEFAULT true,
    created_by      UUID NOT NULL REFERENCES users(id),
    last_deployed_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, branch)
);
CREATE INDEX previews_project_idx ON previews (project_id);

-- +goose Down
DROP TABLE previews;
