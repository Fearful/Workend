-- +goose Up
-- Stage 37: per-project issue board. Issues are pulled from the upstream
-- git provider (GitHub/GitLab/Gitea) into a local cache; the board columns
-- are subsets of the upstream label set picked at first sync. Drag/drop on
-- the UI changes labels both in our cache and on the upstream API.

-- One board per project. `columns` is an ordered array of {name, position}
-- where each name is an existing upstream label.
CREATE TABLE issue_boards (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID        NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    columns         JSONB       NOT NULL DEFAULT '[]'::jsonb,
    last_synced_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Local cache of upstream issues. provider_number is GitHub's issue number,
-- GitLab's iid, Gitea's index — all small ints unique per repo.
CREATE TABLE issues (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider_number     INT         NOT NULL,
    title               TEXT        NOT NULL,
    body                TEXT,
    state               TEXT        NOT NULL DEFAULT 'open',  -- 'open' | 'closed'
    labels              JSONB       NOT NULL DEFAULT '[]'::jsonb,
    author_handle       TEXT,
    author_url          TEXT,
    html_url            TEXT        NOT NULL,
    upstream_updated_at TIMESTAMPTZ,
    fetched_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX issues_project_number_idx ON issues (project_id, provider_number);
CREATE INDEX issues_project_state_idx ON issues (project_id, state);

CREATE TABLE issue_comments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id      UUID        NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    provider_id   BIGINT      NOT NULL,
    body          TEXT        NOT NULL,
    author_handle TEXT,
    html_url      TEXT,
    created_at    TIMESTAMPTZ NOT NULL,
    fetched_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX issue_comments_issue_provider_idx ON issue_comments (issue_id, provider_id);
CREATE INDEX issue_comments_issue_idx ON issue_comments (issue_id, created_at);

-- +goose Down
DROP TABLE issue_comments;
DROP TABLE issues;
DROP TABLE issue_boards;
