-- +goose Up
-- Per-run branch label. NULL = ran against the project's currently-checked-out
-- branch (legacy). Non-null = ephemeral checkout via Spec.CommitSHA / GitURL.
ALTER TABLE runs ADD COLUMN branch TEXT;
CREATE INDEX runs_branch_idx ON runs (branch) WHERE branch IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS runs_branch_idx;
ALTER TABLE runs DROP COLUMN IF EXISTS branch;
