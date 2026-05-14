-- +goose Up
ALTER TABLE runs ADD COLUMN archived_at TIMESTAMPTZ;
CREATE INDEX runs_archived_idx ON runs (archived_at) WHERE archived_at IS NOT NULL;

CREATE TABLE retention_policies (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    max_age_days  INT NOT NULL DEFAULT 90,
    max_runs      INT NOT NULL DEFAULT 1000,
    archive_after_days INT NOT NULL DEFAULT 30,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(workspace_id)
);

-- +goose Down
DROP TABLE retention_policies;
ALTER TABLE runs DROP COLUMN archived_at;
