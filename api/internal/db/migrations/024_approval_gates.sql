-- +goose Up
-- Stage 35: approval gates. A task may be flagged as requiring manual
-- approval before it executes; runs of such tasks land in a new
-- 'pending_approval' status until a workspace member approves them.
ALTER TABLE tasks
    ADD COLUMN requires_approval BOOLEAN NOT NULL DEFAULT false;

-- Approval decisions are recorded so we can show "approved by X at Y" on
-- the run page. One approval row per run.
CREATE TABLE run_approvals (
    run_id     UUID        PRIMARY KEY REFERENCES runs(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    approved   BOOLEAN     NOT NULL,
    note       TEXT,
    decided_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE run_approvals;
ALTER TABLE tasks DROP COLUMN requires_approval;
