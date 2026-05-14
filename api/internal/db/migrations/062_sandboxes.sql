-- +goose Up
CREATE TABLE sandboxes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    branch       TEXT NOT NULL DEFAULT 'main',
    status       TEXT NOT NULL DEFAULT 'provisioning', -- provisioning, running, stopped, destroyed, failed
    url          TEXT NOT NULL DEFAULT '',
    port         INT NOT NULL DEFAULT 0,
    container_id TEXT NOT NULL DEFAULT '',
    created_by   UUID NOT NULL REFERENCES users(id),
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    destroyed_at TIMESTAMPTZ
);
CREATE INDEX sandboxes_project_idx ON sandboxes (project_id);
CREATE INDEX sandboxes_workspace_idx ON sandboxes (workspace_id);
CREATE INDEX sandboxes_status_idx ON sandboxes (status) WHERE status IN ('provisioning', 'running');

-- +goose Down
DROP TABLE sandboxes;
