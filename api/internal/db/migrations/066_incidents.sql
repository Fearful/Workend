-- +goose Up
CREATE TABLE incidents (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    title         TEXT NOT NULL DEFAULT '',
    severity      TEXT NOT NULL DEFAULT 'warning', -- info, warning, critical
    status        TEXT NOT NULL DEFAULT 'open', -- open, investigating, resolved
    started_at    TIMESTAMPTZ NOT NULL,
    resolved_at   TIMESTAMPTZ,
    summary       TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX incidents_workspace_idx ON incidents (workspace_id, started_at DESC);

CREATE TABLE incident_runs (
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    run_id      UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    PRIMARY KEY (incident_id, run_id)
);

-- +goose Down
DROP TABLE incident_runs;
DROP TABLE incidents;
