-- +goose Up
-- Stage 39: gitleaks-driven secret scan, run as part of project sync.
CREATE TABLE secret_findings (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    commit_sha   TEXT,
    file         TEXT        NOT NULL,
    line_no      INT,
    rule         TEXT        NOT NULL,
    severity     TEXT        NOT NULL DEFAULT 'high', -- 'critical' | 'high' | 'medium' | 'low'
    fingerprint  TEXT        NOT NULL,                -- gitleaks's stable fingerprint to dedupe
    found_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at  TIMESTAMPTZ
);
CREATE UNIQUE INDEX secret_findings_fp_idx ON secret_findings (project_id, fingerprint);
CREATE INDEX secret_findings_project_idx  ON secret_findings (project_id, found_at DESC);

-- +goose Down
DROP TABLE secret_findings;
