-- +goose Up
CREATE TABLE custom_roles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    is_system    BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, name)
);

CREATE TABLE role_permissions (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id   UUID NOT NULL REFERENCES custom_roles(id) ON DELETE CASCADE,
    resource  TEXT NOT NULL,
    action    TEXT NOT NULL,
    UNIQUE (role_id, resource, action)
);
CREATE INDEX role_permissions_role_idx ON role_permissions (role_id);

CREATE TABLE member_role_assignments (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id      UUID NOT NULL REFERENCES custom_roles(id) ON DELETE CASCADE,
    assigned_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, user_id)
);

-- +goose Down
DROP TABLE member_role_assignments;
DROP TABLE role_permissions;
DROP TABLE custom_roles;
