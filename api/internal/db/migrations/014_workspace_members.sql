-- +goose Up
CREATE TABLE workspace_members (
    workspace_id UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role         TEXT        NOT NULL DEFAULT 'member',  -- 'owner' | 'member'
    added_by     UUID        REFERENCES users(id) ON DELETE SET NULL,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX workspace_members_user_idx ON workspace_members (user_id);

-- Backfill: every existing workspace.user_id becomes the owner.
INSERT INTO workspace_members (workspace_id, user_id, role, added_by)
SELECT id, user_id, 'owner', user_id FROM workspaces;

-- +goose Down
DROP TABLE workspace_members;
