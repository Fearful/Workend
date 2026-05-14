-- +goose Up
CREATE TABLE commit_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sha         TEXT NOT NULL,
    author_name TEXT NOT NULL DEFAULT '',
    author_email TEXT NOT NULL DEFAULT '',
    message     TEXT NOT NULL DEFAULT '',
    committed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    files_changed INT NOT NULL DEFAULT 0,
    insertions  INT NOT NULL DEFAULT 0,
    deletions   INT NOT NULL DEFAULT 0,
    UNIQUE (project_id, sha)
);
CREATE INDEX commit_history_project_time_idx ON commit_history (project_id, committed_at DESC);

-- +goose Down
DROP TABLE commit_history;
